package generator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hinak0/ClashConfigConverter/config"
	"github.com/hinak0/ClashConfigConverter/log"
	"github.com/hinak0/ClashConfigConverter/proto"
	"gopkg.in/yaml.v3"
)

var (
	proxiesNames []string
)

func ParseSubscriptions(subscriptions []config.Subscription) (proxies []map[string]interface{}) {

	client := &http.Client{}

	for _, subscription := range subscriptions {
		res, err := getSingleSubscription(client, subscription)
		if err != nil {
			fmt.Printf("Error pulling %s : %v\n", subscription.URL, err)
			continue
		}
		tmpConfig := proto.RawConfig{}
		err = yaml.Unmarshal([]byte(res), &tmpConfig)
		if err != nil {
			log.Warnln("Error when parse subscription:%s", res)
		}
		proxies = append(proxies, tmpConfig.Proxy...)
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

func getAllProxyName(proxies []map[string]interface{}) (proxiesNameList []string) {
	for _, p := range proxies {
		name := p["name"].(string)
		proxiesNameList = append(proxiesNameList, name)
	}
	return proxiesNameList
}

func ParseProxyGroup(rowGroups []proto.ProxyGroup, proxies []map[string]interface{}) (groups []proto.ProxyGroup) {
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

	proxies := ParseSubscriptions(c.Subscriptions)
	proxyGroups := ParseProxyGroup(c.ProxyGroup, proxies)
	rules := ParseRuleSet(c.RuleSets)

	baseConfig := proto.RawConfig{}

	f, _ := os.Open(c.BaseFile)
	defer f.Close()
	data, _ := io.ReadAll(f)

	err := yaml.Unmarshal(data, &baseConfig)
	if err != nil {
		log.Errorln("Failed to parse base config.")
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

	f.Write([]byte(content))
}
