package gen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aggronmagi/wctl/xlsx/parser"
	"github.com/spf13/pflag"
)

type ExportOption struct {
	// 当前输出路径
	Outpath string
	// Type 路径
	TypePath string
	// Data 路径
	DataPath string
	// 导出标记
	ExportFlag parser.ExportFlag
}

// ServerOption
//go:generate gogen option -n SupportOption -o options.go
func xlsxSupportConfig() interface{} {
	return map[string]interface{}{
		// 导出类型文件
		"ExportDefine": (func(sheet *parser.XlsxSheet, opts *ExportOption) (err error))(nil),
		// 合并导出类型
		"ExportMergeDefine": (func(sheets []*parser.XlsxSheet, opts *ExportOption) (err error))(nil),
		// 导出数据文件
		"ExportData": (func(sheet *parser.XlsxSheet, opts *ExportOption) (err error))(nil),
		// 合并导出数据
		"ExportMergeData": (func(sheets []*parser.XlsxSheet, opts *ExportOption) (err error))(nil),
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
	outType string
	// 导出标记 0:服务器和客户端 1;服务器 2:客户端
	exportFlag int64
}

func NewExportConfig(language string, opts ...SupportOption) *ExportSupportConfig {
	cfg := &ExportSupportConfig{
		Language: language,
		Opts:     NewSupportOptions(opts...),
		setFlag:  false,
		// 默认配置
		outData:    fmt.Sprintf("./%s/data", language),
		outType:    fmt.Sprintf("./%s/type", language),
		exportFlag: 0,
	}
	cfg.BoolVar(&cfg.setFlag, "gen", cfg.setFlag, "生成标记,其他选项都使用默认值时候,开启生成")
	// 导出标记
	cfg.Int64OptionsVar(&cfg.exportFlag, "flag", cfg.exportFlag, " 0:服务器和客户端 1;服务器 2:客户端", []int64{0, 1, 2})
	set := false
	if cfg.Opts.ExportDefine != nil || cfg.Opts.ExportMergeDefine != nil {
		cfg.StringVar(&cfg.outType, "type", cfg.outType, "类型导出目录")
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
func (cfg *ExportSupportConfig) OutpathType() string {
	if cfg.Opts.ExportDefine == nil && cfg.Opts.ExportMergeDefine == nil {
		return "not support type define export"
	}
	if !filepath.IsAbs(cfg.outType) {
		if path, err := filepath.Abs(cfg.outType); err == nil {
			cfg.outType = path
		}
	}
	return cfg.outType
}

func (cfg *ExportSupportConfig) ExportFlag() parser.ExportFlag {
	return parser.ExportFlag(cfg.exportFlag)
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

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (cfg *ExportSupportConfig) Int64Var(p *int64, name string, value int64, usage string, updates ...func()) {
	usage = cfg.Language + " " + usage
	cfg.configs = append(cfg.configs, func(set *pflag.FlagSet) {
		name = fmt.Sprintf("%s-%s", cfg.Language, name)
		*p = value
		set.VarP(&CustomFlagIntValue{val: p, root: cfg, update: updates}, name, "", usage)
	})
}

func (cfg *ExportSupportConfig) Int64OptionsVar(p *int64, name string, value int64, usage string, opts []int64, updates ...func()) {
	usage = cfg.Language + " " + usage
	cfg.configs = append(cfg.configs, func(set *pflag.FlagSet) {
		name = fmt.Sprintf("%s-%s", cfg.Language, name)
		*p = value
		set.VarP(&CustomFlagIntValue{val: p, root: cfg, update: updates, limits: opts}, name, "", usage)
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

// -- int Value
type CustomFlagIntValue struct {
	val    *int64
	root   *ExportSupportConfig
	limits []int64
	update []func()
}

var _ pflag.Value = (*CustomFlagIntValue)(nil)

func (s *CustomFlagIntValue) Set(val string) error {
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	// 数值限制
	if len(s.limits) > 0 {
		find := false
		for _, c := range s.limits {
			if c == v {
				find = true
			}
		}
		if !find {
			return fmt.Errorf("value %d not in %+v", v, s.limits)
		}
	}
	*s.val = v
	s.root.setFlag = true
	for _, uf := range s.update {
		uf()
	}
	return nil
}
func (s *CustomFlagIntValue) Type() string {
	return "string"
}

func (s *CustomFlagIntValue) String() string { return strconv.FormatInt(*s.val, 10) }

func WriteFile(fname string, data []byte) (err error) {
	os.MkdirAll(filepath.Dir(fname), 0755)
	err = ioutil.WriteFile(fname, data, 0644)
	if err != nil {
		return err
	}
	//log.Println("genrate ", fname)
	return nil
}
