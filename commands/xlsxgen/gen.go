package xlsxgen

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var config = struct {
}{}

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
}

// RunCommand run generate command
func RunCommand(cmd *cobra.Command, args []string) {

}
