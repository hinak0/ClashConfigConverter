package config

import (
	"io"
	"os"

	"github.com/hinak0/ClashConfigConverter/log"
	"github.com/hinak0/ClashConfigConverter/proto"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	BaseFile      string                   `yaml:"base-file"`
	TargetPath    string                   `yaml:"target"`
	Subscriptions []Subscription           `yaml:"sub-links"`
	RuleSets      []RuleSet                `yaml:"ruleset"`
	Proxies       []map[string]interface{} `yaml:"proxies"`
	ProxyGroup    []proto.ProxyGroup       `yaml:"proxy-groups"`
}

type Subscription struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
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
		log.Errorln("Failed to parse config.yaml.")
	}

	return Appconfig
}
