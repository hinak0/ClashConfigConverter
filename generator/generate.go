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
	if exclude != "" {
		excludeReg = regexp.MustCompile(exclude)
	}
	client := &http.Client{}

	proxies = append(proxies, preloadProxies...)

	for _, subscription := range subscriptions {
		res, err := getSingleSubscription(client, subscription)
		if err != nil {
			fmt.Printf("Error pulling %s : %v\n", subscription.URL, err)
			continue
		}
		nativeConfig := proto.RawConfig{}
		err = yaml.Unmarshal([]byte(res), &nativeConfig)
		if err != nil {
			log.Warnln("Error when parse subscription:%s", err)
		} else {
			log.Infoln("Success pull subscription: %s", subscription.URL)
		}
		nativePoxies := nativeConfig.Proxy
		if !*subscription.UdpEnable {
			for i := range nativePoxies {
				nativePoxies[i].UdpEnable = false
			}
		}
		proxies = append(proxies, nativePoxies...)
	}

	if excludeReg != nil {
		for i := 0; i < len(proxies); i++ {
			// remove exclude
			if excludeReg.MatchString(proxies[i].Name) {
				log.Infoln("Proxy %s match exclude, delete it.", proxies[i].Name)
				proxies = append(proxies[:i], proxies[i+1:]...)
				i--
			}
		}
	}

	proxiesNames = getAllProxyName(proxies)
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
