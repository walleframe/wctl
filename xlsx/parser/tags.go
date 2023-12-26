package parser

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/multierr"
)

// ExportTags export flag
type ExportTags uint32

const (
	ExportNone   ExportTags = 0x00
	ExportServer ExportTags = 0x01
	ExportClient ExportTags = 0x02
	ExportAll    ExportTags = 0x03
	ExportPM     ExportTags = 0x04
	ExportID     ExportTags = 0x08
)

func (f *ExportTags) Server() bool {
	return ((*f) & ExportServer) > 0
}

func (f *ExportTags) Client() bool {
	return ((*f) & ExportClient) > 0
}

func (f *ExportTags) All() bool {
	return ((*f) & ExportAll) > 0
}

func (f *ExportTags) ID() bool {
	return ((*f) & ExportID) > 0
}


func SplitFlags(raw string) (dst string, flags ExportTags, err error) {
	// 拆分可能的标记
	lists := strings.Split(raw, "@")
	dst = lists[0]
	for i := 1; i < len(lists); i++ {
		f := strings.ToLower(lists[i])
		if f == "" {
			err = fmt.Errorf("flag empty")
			return
		}
		switch f {
		case "c":
			flags |= ExportClient
		case "s":
			flags |= ExportServer
		case "pm":
			flags |= ExportPM
		case "id":
			flags |= ExportID
		default:
			err = multierr.Append(err, errors.New(fmt.Sprintf("unsupport flag [%s]", f)))
		}
	}
	if err != nil {
		return
	}

	// 没有单独指定,那么就是都导出
	if !flags.Client() && !flags.Server() {
		flags |= ExportAll
	}
	// ID 应该服务器客户端都导出
	if flags.ID() {
		flags |= ExportAll
	}
	return
}
