package generator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hinak0/ClashConfigConverter/config"
	"github.com/hinak0/ClashConfigConverter/log"
	"github.com/hinak0/ClashConfigConverter/proto"
	"gopkg.in/yaml.v3"
)

var (
	proxiesNames []string
	excludeReg   *regexp.Regexp
)

func ParseProxies(subscriptions []config.Subscription, exclude string, preloadProxies []proto.Proxy) (proxies []proto.Proxy) {
	// 如果有排除条件，编译正则表达式
	if exclude != "" {
		excludeReg = regexp.MustCompile(exclude)
	}

	client := &http.Client{}

	// 将预加载的代理添加到最终的代理列表中
	proxies = append(proxies, preloadProxies...)

	// 创建一个存储代理名称的集合，避免重名
	nameSet := make(map[string]struct{})
	for _, proxy := range proxies {
		nameSet[proxy.Name] = struct{}{}
	}

	// 遍历订阅，处理每个订阅的数据
	for _, subscription := range subscriptions {
		// 拉取单个订阅内容
		res, err := getSingleSubscription(client, subscription)
		if err != nil {
			fmt.Printf("Error pulling %s : %v\n", subscription.URL, err)
			continue
		}

		// 解析订阅内容为配置
		nativeConfig := proto.RawConfig{}
		err = yaml.Unmarshal([]byte(res), &nativeConfig)
		if err != nil {
			log.Warnln("Error when parse subscription: %s", err)
			log.Warnln("Subscription: %s", res)
		} else {
			log.Infoln("Success pull subscription: %s", subscription.URL)
		}

		// 获取解析后的代理列表
		nativePoxies := nativeConfig.Proxy

		// 如果 UDP 被禁用，禁用所有的代理的 UDP 功能
		if !*subscription.UdpEnable {
			for i := range nativePoxies {
				nativePoxies[i].UdpEnable = false
			}
		}

		// 检查并处理代理的重命名
		for i := range nativePoxies {
			originalName := nativePoxies[i].Name
			newName := originalName
			count := 1

			// 如果代理名已经存在，则添加 [n] 后缀，直到找到唯一的名字
			for {
				if _, exists := nameSet[newName]; exists {
					newName = fmt.Sprintf("%s[%d]", originalName, count)
					count++
				} else {
					break
				}
			}

			// 更新代理名并保存到集合中
			nativePoxies[i].Name = newName
			nameSet[newName] = struct{}{}
		}

		// 将处理后的代理添加到最终代理列表中
		proxies = append(proxies, nativePoxies...)
	}

	// 如果有排除规则，按规则移除匹配的代理
	if excludeReg != nil {
		for i := 0; i < len(proxies); i++ {
			if excludeReg.MatchString(proxies[i].Name) {
				log.Infoln("Proxy %s match exclude, delete it.", proxies[i].Name)
				proxies = append(proxies[:i], proxies[i+1:]...)
				i-- // 因为删除了一个元素，调整索引
			}
		}
	}

	// 获取所有代理的名称列表，用于日志输出
	proxiesNames := getAllProxyName(proxies)
	log.Infoln("Parse subscription success: %v", proxiesNames)

	return proxies
}

func getSingleSubscription(client *http.Client, sub config.Subscription) (string, error) {
	request, _ := http.NewRequest("GET", sub.URL, nil)

	for key, value := range sub.Headers {
		request.Header.Set(key, value)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Warnln("Error when pull subscription %s: %v", sub.URL, err)
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	return string(body), nil
}

func ParseRuleSet(rulesets []config.RuleSet) (rules []string) {

	for _, s := range rulesets {
		// not a file ref
		if s.Value != "" {
			rules = append(rules, s.Value+","+s.Name)
			continue
		}

		file, _ := os.Open(s.Location)
		defer file.Close()
		data, _ := io.ReadAll(file)

		rulelist := strings.Split(string(data), "\n")
		for _, rule := range rulelist {

			rule = strings.TrimSpace(rule)
			// blank or comment
			if rule == "" || strings.HasPrefix(rule, "#") {
				continue
			}

			ruleParams := strings.Split(rule, ",")
			ruleParams = append(ruleParams[:2], append([]string{s.Name}, ruleParams[2:]...)...)
			ruleStr := strings.Join(ruleParams, ",")
			rules = append(rules, ruleStr)
		}
	}
	log.Infoln("Success parse rules.")
	return rules
}

func getAllProxyName(proxies []proto.Proxy) (proxiesNameList []string) {
	for _, p := range proxies {
		name := p.Name
		proxiesNameList = append(proxiesNameList, name)
	}
	return proxiesNameList
}

func ParseProxyGroup(rowGroups []proto.ProxyGroup, proxies []proto.Proxy) (groups []proto.ProxyGroup) {
	// proxiesNames := getAllProxyName(proxies)
	for _, rowGroup := range rowGroups {
		for index, proxyName := range rowGroup.Proxies {
			// replase * to all proxies
			if proxyName == "*" {
				rowGroup.Proxies = append(rowGroup.Proxies[:index], append(proxiesNames, rowGroup.Proxies[index+1:]...)...)
				break
			}
		}
		groups = append(groups, rowGroup)
	}
	log.Infoln("Success parse ProxyGroups.")
	return
}

func Integrate(c config.AppConfig) {

	proxies := ParseProxies(c.Subscriptions, c.Exclude, c.Proxies)
	proxyGroups := ParseProxyGroup(c.ProxyGroup, proxies)
	rules := ParseRuleSet(c.RuleSets)

	baseConfig := proto.RawConfig{}

	f, _ := os.Open(c.BaseFile)
	defer f.Close()
	data, _ := io.ReadAll(f)

	err := yaml.Unmarshal(data, &baseConfig)
	if err != nil {
		log.Fatalln("Failed to parse base config: %v.", err)
	}

	baseConfig.Proxy = proxies
	baseConfig.ProxyGroup = proxyGroups
	baseConfig.Rule = rules

	result, _ := yaml.Marshal(baseConfig)
	WriteTarget(c.TargetPath, string(result))
}

func WriteTarget(path string, content string) {
	f, _ := os.Create(path)
	defer f.Close()

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	comment := fmt.Sprintf("# Generate by ClashConfigConverter.\n# %s\n", currentTime)
	_, err := f.WriteString(comment)
	if err != nil {
		log.Errorln("Failed to write timestamp into target.yaml:	%v", err)
	}

	_, err = f.Write([]byte(content))
	if err != nil {
		log.Errorln("Failed to write target.yaml:	%v", err)
	}
}
