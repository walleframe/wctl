package parser

import (
	"fmt"
	"strings"
)

func SplitFlags(raw string) (dst string, flags map[string]struct{}, err error) {
	// 拆分可能的标记
	lists := strings.Split(raw, "@")
	dst = lists[0]
	flags = make(map[string]struct{})
	for i := 1; i < len(lists); i++ {
		if lists[i] == "" {
			err = fmt.Errorf("flag empty")
			return
		}
		if !supportFlag(lists[i]) {
			err = fmt.Errorf("unsupport flag %s", lists[i])
			return
		}
		flags[lists[i]] = struct{}{}
	}
	return
}

func supportFlag(flag string) bool {
	switch flag {
	case "c", // 仅支持客户端
		"s",  // 仅支持服务器
		"id", // 数据索引项
		"pm": // 策划
		return true
	}
	return false
}
