package gen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aggronmagi/wctl/xlsx/parser"
	"github.com/spf13/pflag"
)

// ServerOption
//go:generate gogen option -n SupportOption -o options.go
func xlsxSupportConfig() interface{} {
	return map[string]interface{}{
		// 导出类型文件
		"ExportDefine": (func(sheet *parser.XlsxSheet, outpath string) (err error))(nil),
		// 合并导出类型
		"ExportMergeDefine": (func(sheets []*parser.XlsxSheet, outpath string) (err error))(nil),
		// 导出数据文件
		"ExportData": (func(sheet *parser.XlsxSheet, outpath string) (err error))(nil),
		// 合并导出数据
		"ExportMergeData": (func(sheets []*parser.XlsxSheet, outpath string) (err error))(nil),
		// 检测配置
		"CheckOptions": func() error { return nil },
	}
}

// ExportSupport 导出生成支持的配置项
type ExportSupportConfig struct {
	// Language 导出名字
	Language string
	// 导出选项
	Opts *SupportOptions

	// 是否设置了标记
	setFlag bool
	// configs
	configs []func(set *pflag.FlagSet)

	// 默认配置项
	// 数据输出目录
	outData string
	// 类型输出目录
	outDef string
}

func NewExportConfig(language string, opts ...SupportOption) *ExportSupportConfig {
	cfg := &ExportSupportConfig{
		Language: language,
		Opts:     NewSupportOptions(opts...),
		setFlag:  false,
		// 默认配置
		outData: fmt.Sprintf("./%s/data", language),
		outDef:  fmt.Sprintf("./%s/def", language),
	}
	cfg.BoolVar(&cfg.setFlag, "gen", cfg.setFlag, "生成标记,其他选项都使用默认值时候,开启生成")
	set := false
	if cfg.Opts.ExportDefine != nil || cfg.Opts.ExportMergeDefine != nil {
		cfg.StringVar(&cfg.outDef, "type", cfg.outDef, "类型导出目录")
		set = true
	}
	if cfg.Opts.ExportData != nil || cfg.Opts.ExportMergeData != nil {
		cfg.StringVar(&cfg.outData, "data", cfg.outData, "数据导出目录")
		set = true
	}

	if !set {
		panic(fmt.Sprintf("language [%s] not support any export,check your code.", language))
	}
	return cfg
}

func (cfg *ExportSupportConfig) SetFlagSet(set *pflag.FlagSet) {
	for _, sf := range cfg.configs {
		sf(set)
	}
}

// 数据导出目录
func (cfg *ExportSupportConfig) OutpathData() string {
	if cfg.Opts.ExportData == nil && cfg.Opts.ExportMergeData == nil {
		return "not support data value export"
	}
	if !filepath.IsAbs(cfg.outData) {
		if path, err := filepath.Abs(cfg.outData); err == nil {
			cfg.outData = path
		}
	}
	return cfg.outData
}

// 类型导出目录
func (cfg *ExportSupportConfig) OutpathDef() string {
	if cfg.Opts.ExportDefine == nil && cfg.Opts.ExportMergeDefine == nil {
		return "not support type define export"
	}
	if !filepath.IsAbs(cfg.outDef) {
		if path, err := filepath.Abs(cfg.outDef); err == nil {
			cfg.outDef = path
		}
	}
	return cfg.outDef
}

func (cfg *ExportSupportConfig) HasSetFlag() bool {
	return cfg.setFlag
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (cfg *ExportSupportConfig) StringVar(p *string, name string, value string, usage string, updates ...func()) {
	usage = cfg.Language + " " + usage
	cfg.configs = append(cfg.configs, func(set *pflag.FlagSet) {
		name = fmt.Sprintf("%s-%s", cfg.Language, name)
		*p = value
		set.VarP(&CustomFlagStringValue{val: p, root: cfg, update: updates}, name, "", usage)
	})
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (cfg *ExportSupportConfig) BoolVar(p *bool, name string, value bool, usage string, updates ...func()) {
	usage = cfg.Language + " " + usage
	cfg.configs = append(cfg.configs, func(set *pflag.FlagSet) {
		name = fmt.Sprintf("%s-%s", cfg.Language, name)
		*p = value
		flag := set.VarPF(&CustomFlagBoolValue{val: p, root: cfg, update: updates}, name, "", usage)
		flag.NoOptDefVal = "true"
	})
}

// -- string Value
type CustomFlagStringValue struct {
	val    *string
	root   *ExportSupportConfig
	update []func()
}

var _ pflag.Value = (*CustomFlagStringValue)(nil)

func (s *CustomFlagStringValue) Set(val string) error {
	*s.val = val
	s.root.setFlag = true
	for _, uf := range s.update {
		uf()
	}
	return nil
}
func (s *CustomFlagStringValue) Type() string {
	return "string"
}

func (s *CustomFlagStringValue) String() string { return *s.val }

// -- bool Value
type CustomFlagBoolValue struct {
	val    *bool
	root   *ExportSupportConfig
	update []func()
}

var _ pflag.Value = (*CustomFlagBoolValue)(nil)

func (s *CustomFlagBoolValue) Set(val string) error {
	v, err := parser.ParseBool(val)
	if err != nil {
		return err
	}
	*s.val = v
	s.root.setFlag = true
	for _, uf := range s.update {
		uf()
	}
	return nil
}
func (s *CustomFlagBoolValue) Type() string {
	return "string"
}

func (s *CustomFlagBoolValue) String() string { return parser.FormatBool(*s.val) }

func WriteFile(fname string, data []byte) (err error) {
	os.MkdirAll(filepath.Dir(fname), 0755)
	err = ioutil.WriteFile(fname, data, 0644)
	if err != nil {
		return err
	}
	//log.Println("genrate ", fname)
	return nil
}
