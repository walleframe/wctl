// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n SupportOption -o options.go"
// Version: 0.0.4

package gen

import (
	"github.com/walleframe/wctl/xlsx/parser"
)

var _ = xlsxSupportConfig()

// ServerOption
type SupportOptions struct {
	// 导出类型文件
	ExportDefine func(sheet *parser.XlsxSheet, opts *ExportOption) (err error)
	// 合并导出类型
	ExportMergeDefine func(sheets []*parser.XlsxSheet, opts *ExportOption) (err error)
	// 导出数据文件
	ExportData func(sheet *parser.XlsxSheet, opts *ExportOption) (err error)
	// 合并导出数据
	ExportMergeData func(sheets []*parser.XlsxSheet, opts *ExportOption) (err error)
	// 检测配置
	CheckOptions func() error
}

// 导出类型文件
func WithExportDefine(v func(sheet *parser.XlsxSheet, opts *ExportOption) (err error)) SupportOption {
	return func(cc *SupportOptions) SupportOption {
		previous := cc.ExportDefine
		cc.ExportDefine = v
		return WithExportDefine(previous)
	}
}

// 合并导出类型
func WithExportMergeDefine(v func(sheets []*parser.XlsxSheet, opts *ExportOption) (err error)) SupportOption {
	return func(cc *SupportOptions) SupportOption {
		previous := cc.ExportMergeDefine
		cc.ExportMergeDefine = v
		return WithExportMergeDefine(previous)
	}
}

// 导出数据文件
func WithExportData(v func(sheet *parser.XlsxSheet, opts *ExportOption) (err error)) SupportOption {
	return func(cc *SupportOptions) SupportOption {
		previous := cc.ExportData
		cc.ExportData = v
		return WithExportData(previous)
	}
}

// 合并导出数据
func WithExportMergeData(v func(sheets []*parser.XlsxSheet, opts *ExportOption) (err error)) SupportOption {
	return func(cc *SupportOptions) SupportOption {
		previous := cc.ExportMergeData
		cc.ExportMergeData = v
		return WithExportMergeData(previous)
	}
}

// 检测配置
func WithCheckOptions(v func() error) SupportOption {
	return func(cc *SupportOptions) SupportOption {
		previous := cc.CheckOptions
		cc.CheckOptions = v
		return WithCheckOptions(previous)
	}
}

// SetOption modify options
func (cc *SupportOptions) SetOption(opt SupportOption) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *SupportOptions) ApplyOption(opts ...SupportOption) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *SupportOptions) GetSetOption(opt SupportOption) SupportOption {
	return opt(cc)
}

// SupportOption option define
type SupportOption func(cc *SupportOptions) SupportOption

// NewSupportOptions create options instance.
func NewSupportOptions(opts ...SupportOption) *SupportOptions {
	cc := newDefaultSupportOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogSupportOptions != nil {
		watchDogSupportOptions(cc)
	}
	return cc
}

// InstallSupportOptionsWatchDog install watch dog
func InstallSupportOptionsWatchDog(dog func(cc *SupportOptions)) {
	watchDogSupportOptions = dog
}

var watchDogSupportOptions func(cc *SupportOptions)

// newDefaultSupportOptions new option with default value
func newDefaultSupportOptions() *SupportOptions {
	cc := &SupportOptions{
		ExportDefine:      nil,
		ExportMergeDefine: nil,
		ExportData:        nil,
		ExportMergeData:   nil,
		CheckOptions: func() error {
			return nil
		},
	}
	return cc
}
