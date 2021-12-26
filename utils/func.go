package utils

import "strings"

func Title(in string) (out string) {
	list := strings.Split(in, "_")
	for _, v := range list {
		out += strings.Title(v)
	}
	return
}
