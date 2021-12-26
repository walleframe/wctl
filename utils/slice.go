package utils

import "strings"

func MergeNotSameSlice(l1, l2 []string) (res []string) {
	for _, v := range l1 {
		if v == "" {
			continue
		}
		res = append(res, strings.ToLower(v))
	}
	for _, v := range l2 {
		if v == "" {
			continue
		}
		find := false
		for _, v2 := range res {
			if v == v2 {
				find = true
				break
			}
		}
		if find {
			continue
		}
		res = append(res, strings.ToLower(v))
	}
	return
}
