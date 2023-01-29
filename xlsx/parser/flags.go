package parser

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/multierr"
)

// ExportFlag export flag
type ExportFlag int64

const (
	ExportAll ExportFlag = iota
	ExportServer
	ExportClient
)

// Flag
type Flag struct {
	// 仅导出服务器
	Server bool
	// 仅导出客户端
	Client bool
	// pm 不导出
	Pm bool
	// ID 索引标记
	ID bool
}

func SplitFlags(raw string) (dst string, flags Flag, err error) {
	// 拆分可能的标记
	lists := strings.Split(raw, "@")
	dst = lists[0]
	for i := 1; i < len(lists); i++ {
		if lists[i] == "" {
			err = fmt.Errorf("flag empty")
			return
		}
		switch lists[i] {
		case "c":
			flags.Client = true
		case "s":
			flags.Server = true
		case "pm":
			flags.Pm = true
		case "id":
			flags.ID = true
		default:
			err = multierr.Append(err, errors.New(fmt.Sprintf("unsupport flag [[%s]", lists[i])))
		}
	}
	if err != nil {
		return
	}
	// 没有单独指定,那么就是都导出
	if !flags.Client && !flags.Server {
		flags.Server = true
		flags.Client = true
	}
	// ID 应该服务器客户端都导出
	if flags.ID {
		flags.Server = true
		flags.Client = true
	}
	return
}
