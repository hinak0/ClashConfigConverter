package generator

import (
	"regexp"

	"github.com/hinak0/ClashConfigConverter/proto"
)

// RemoveEmojis 移除字符串中的 Emoji 字符
func RemoveEmojis(input string) string {
	// 使用正则表达式匹配大部分 Emoji 字符
	re := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F900}-\x{1F9FF}]|[\x{1FA70}-\x{1FAFF}]|[\x{1F1E6}-\x{1F1FF}]`)
	return re.ReplaceAllString(input, "")
}

func getAllProxyName(proxies []proto.Proxy) (proxiesNameList []string) {
	for _, p := range proxies {
		proxiesNameList = append(proxiesNameList, p.Name)
	}
	return proxiesNameList
}
