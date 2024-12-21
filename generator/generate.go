package generator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hinak0/ClashConfigConverter/config"
	"github.com/hinak0/ClashConfigConverter/log"
	"github.com/hinak0/ClashConfigConverter/proto"
	"gopkg.in/yaml.v3"
)

var (
	excludeReg *regexp.Regexp
)

func ParseProxies(subscriptions []config.Subscription, exclude string, preloadProxies []proto.Proxy) (proxies []proto.Proxy) {
	if exclude != "" {
		excludeReg = regexp.MustCompile(exclude)
	}

	client := &http.Client{}
	proxies = append(proxies, preloadProxies...)

	var wg sync.WaitGroup
	proxyChan := make(chan []proto.Proxy, len(subscriptions))

	for _, subscription := range subscriptions {
		wg.Add(1)
		go func(sub config.Subscription) {
			defer wg.Done()
			proxies, err := getSingleSubscription(client, sub)
			if err != nil {
				log.Warnln("Error pulling %s: %v", sub.URL, err)
				return
			}
			proxyChan <- proxies
		}(subscription)
	}

	go func() {
		wg.Wait()
		close(proxyChan)
	}()

	for currentProxies := range proxyChan {
		proxies = append(proxies, currentProxies...)
	}

	log.Infoln("Pull all subscriptions successfully.")

	// 正则匹配排除
	if excludeReg != nil {
		for i := 0; i < len(proxies); i++ {
			if excludeReg.MatchString(proxies[i].Name) {
				log.Infoln("Proxy %s match exclude, delete it.", proxies[i].Name)
				proxies = append(proxies[:i], proxies[i+1:]...)
				i--
			}
		}
	}

	// 去重
	nameSet := make(map[string]struct{})
	for i, proxy := range proxies {
		originalName := proxy.Name
		newName := originalName
		count := 1

		for {
			if _, exists := nameSet[newName]; exists {
				newName = fmt.Sprintf("%s[%d]", originalName, count)
				count++
			} else {
				break
			}
		}

		proxies[i].Name = newName
		nameSet[newName] = struct{}{}
	}

	proxiesNames := getAllProxyName(proxies)
	log.Infoln("Parse subscription success: %s", strings.Join(proxiesNames, ","))

	return proxies
}

// 拉取单个订阅
func getSingleSubscription(client *http.Client, sub config.Subscription) ([]proto.Proxy, error) {
	request, _ := http.NewRequest("GET", sub.URL, nil)

	for key, value := range sub.Headers {
		request.Header.Set(key, value)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Warnln("Error when pull subscription %s: %v", sub.URL, err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Warnln("Error pulling %s: %v", sub.URL, err)
		return nil, err
	}

	nativeConfig := proto.RawConfig{}
	err = yaml.Unmarshal(body, &nativeConfig)
	if err != nil {
		log.Warnln("Error when parse subscription %s: %v", sub.URL, err)
		return nil, err
	}

	log.Infoln("Successfully pull subscription: %s", sub.URL)

	proxies := nativeConfig.Proxy

	// 设置updEnable
	if sub.UdpEnable != nil && !*sub.UdpEnable {
		for i := range proxies {
			proxies[i].UdpEnable = false
		}
	}

	// 去除emoji
	for i := range proxies {
		proxies[i].Name = RemoveEmojis(proxies[i].Name)
	}

	// 设置组名
	if sub.Name != "" {
		for i := range proxies {
			// [name]proxyname
			proxies[i].Name = fmt.Sprintf("[%s]%s", sub.Name, proxies[i].Name)
		}
	}

	return proxies, nil
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

func ParseProxyGroup(rowGroups []proto.ProxyGroup, proxies []proto.Proxy) (groups []proto.ProxyGroup) {
	proxiesNames := getAllProxyName(proxies)

	for _, rowGroup := range rowGroups {
		for index, proxyName := range rowGroup.Proxies {
			// replace * to all proxies
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
