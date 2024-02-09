package config

import (
	"io"
	"os"

	"github.com/hinak0/ClashConfigConverter/log"
	"github.com/hinak0/ClashConfigConverter/proto"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	BaseFile      string             `yaml:"base-file"`
	TargetPath    string             `yaml:"target"`
	Exclude       string             `yaml:"exclude"`
	Subscriptions []Subscription     `yaml:"sub-links"`
	RuleSets      []RuleSet          `yaml:"ruleset"`
	Proxies       []proto.Proxy      `yaml:"proxies"`
	ProxyGroup    []proto.ProxyGroup `yaml:"proxy-groups"`
}

type Subscription struct {
	URL       string            `yaml:"url"`
	Headers   map[string]string `yaml:"headers"`
	UdpEnable *bool             `yaml:"udp,omitempty"`
}

type RuleSet struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location,omitempty"`
	Value    string `yaml:"value,omitempty"`
}

func Parse() AppConfig {
	file, _ := os.Open("config.yaml")
	defer file.Close()

	data, _ := io.ReadAll(file)

	Appconfig := AppConfig{}
	err := yaml.Unmarshal(data, &Appconfig)
	if err != nil {
		log.Fatalln("Failed to parse config.yaml: %v.", err)
	}

	defaultUdpEnableudpEnable := true
	for i := range Appconfig.Subscriptions {
		if Appconfig.Subscriptions[i].UdpEnable == nil {
			Appconfig.Subscriptions[i].UdpEnable = &defaultUdpEnableudpEnable
		}
	}

	return Appconfig
}
