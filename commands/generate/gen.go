package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/walleframe/wctl/builder"
	"github.com/walleframe/wctl/builder/yttpl"
	"github.com/walleframe/wctl/protocol"
	"github.com/walleframe/wctl/protocol/ast"
	"github.com/walleframe/wctl/utils"
)

var config = struct {
	// 输入,输出目录
	input, output string
	// 生成指定文件
	files []string
	// 内部生成器
	useGens []string
	// 模板生成器
	tplCfg []string
	// go插件生成器
	goPlugins []string
	// 命令行插件
	cmdPlguins []string
	// 全局选项,属性配置
	options []string
	// 是否合并文件
	mergeFile bool
	// 文件名后缀
	fileSuffix string
}{
	fileSuffix: ".wproto",
}

const (
	// Version 命令行工具版本
	Version = "0.0.1"

	Help = `代码生成:

1. 解析目标文件没有依赖其它文件的情况(没有指定 -o情况. 输出到相同目录):
	wctl gen -f path/xx.yt
	wctl gen path/xx.yt
命令行 -f 可以省略.直接输入文件名. 与-f效果相同. (两者同时存在,都会解析)
-f 参数 可以多次设置. 当有指定文件时候,只解析并生成指定文件.


2. 解析目标文件依赖其他文件.并且它们目录都是基于某个父目录. 使用 -i/--input 参数
	wctl gen -i base_dir -f path/xx.yt
	wctl gen -i base_idr path/xx.yt
这个时候,文件需要使用相对路径.(基于-i参数指定的目录)
输入文件会 输出到同样的相对路径(基于-o参数)
如果没有指定 -f 参数. 将递归解析-i所在目录下所有 .yt 文件

内置生成器：
  printer 用于打印输出信息，调试用
`
	Example = `解析单个文件
  wctl gen -f path/xx.yt
  wctl gen path/xx.yt
解析多个文件
  wctl gen -f path/xx.yt -f path/x2.yt
  wctl gen path/xx.yt path/x2.yt
解析基于某个目录的文件
  wctl gen -i base_dir -f path/xx.yt
解析某个目录
  wctl gen -i base_dir 
`
)

func Flags(genCmd *pflag.FlagSet) {
	// 参数不排序
	genCmd.SortFlags = false

	// 输入输出
	genCmd.StringSliceVarP(&config.files, "file", "f", nil, "解析文件")
	genCmd.StringVarP(&config.input, "input", "i", "./", "输入基础路径.查找文件基于这个目录进行查找.")
	genCmd.StringVarP(&config.output, "output", "o", "", "输出文件路径,默认使用input目录")

	// 插件支持
	genCmd.StringSliceVar(&config.goPlugins, "go-plugin", nil, "go版本插件")
	genCmd.StringSliceVar(&config.useGens, "lang", nil, "内置插件")
	genCmd.StringSliceVarP(&config.tplCfg, "template", "t", nil, "创建模板生成器 配置文件名")
	genCmd.StringSliceVarP(&config.cmdPlguins, "cmd", "c", nil, "创建命令行生成器 可执行文件名")

	// 全局选项
	genCmd.StringSliceVar(&config.options, "options", nil, `全局Options. 格式为 "xx.xxx=66" "xx.x1" "xx.xx2=xxx"`)
	genCmd.BoolVar(&ast.Flag.ServiceUseMethodID, "use-method-id", ast.Flag.ServiceUseMethodID, "是否使用数值做请求ID")
	genCmd.StringVar(&config.fileSuffix, "suffix", config.fileSuffix, "解析文件后缀名")
	genCmd.BoolVarP(&config.mergeFile, "merge-same-file", "m", false, "执行命令时,不论是否是同一个插件. 生成文件名相同时候,是否合并文件(开启后,会在内存缓存生成的文件信息)")
}

// RunCommand run generate command
func RunCommand(cmd *cobra.Command, args []string) {
	prepareGenCmd(args)
	executeGenCmd()
}

func prepareGenCmd(args []string) {
	var err error
	config.files = append(config.files, args...)
	config.input, _ = filepath.Abs(config.input)
	// 修改输出目录.默认和输入目录相同
	if config.output == "" {
		config.output = config.input
	}
	config.output, _ = filepath.Abs(config.output)
	// 加载插件
	for _, v := range config.goPlugins {
		err = builder.LoadGoPluginGenerater(v)
		if err != nil {
			fmt.Printf("Error load plugin failed. [%s]. %+v\n", v, err)
			os.Exit(1)
		}
		fmt.Println("INFO", "加载go插件", v)
	}
	// 使用的生成器
	for _, v := range config.useGens {
		err = builder.EnableGenerator(v)
		if err != nil {
			fmt.Printf("enable generater failed. [%s]. %+v\n", v, err)
			os.Exit(1)
		}
		fmt.Println("INFO", "使用", v, "生成器")
	}
	// 模板生成器
	for _, v := range config.tplCfg {
		err = yttpl.NewTemplateGenerator(v)
		if err != nil {
			fmt.Printf("create template generater failed. [%s]. %+v\n", v, err)
			os.Exit(1)
		}
	}
	// 命令行生成器
	for _, v := range config.cmdPlguins {
		err = builder.NewCmdPluginGenerater(v)
		if err != nil {
			fmt.Printf("create command generater failed. [%s]. %+v\n", v, err)
			os.Exit(1)
		}
	}
}

func executeGenCmd() {
	if utils.Debug() {
		fmt.Printf("config %#v", config)
	}
	protocol.SetBasePath(config.input)
	var progList []*ast.YTProgram
	var prog *ast.YTProgram
	var err error
	// 生成指定文件
	if len(config.files) > 0 {
		// 文件列表
		for _, file := range config.files {
			prog, err = protocol.AnalyseFile(file)
			utils.PanicIf(err)
			progList = append(progList, prog)
		}
	} else {
		// 解析目录
		progList, err = protocol.AnlysePath(config.input, config.fileSuffix)
		utils.PanicIf(err)
	}
	for _, v := range progList {
		v.ApplyCmdOptions(config.options...)
	}
	// 解析完成. 进行生成
	err = builder.Build(progList, config.output, config.mergeFile)
	utils.PanicIf(err)
	// 完成
	if utils.Debug() {
		fmt.Println("finish")
	}
	return
}
