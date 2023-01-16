package xlsxgen

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var config = struct {
}{}

const (
	// Version 命令行工具版本
	Version = "0.0.1"

	Help = `
`
	Example = `
`
)

func Flags(genCmd *pflag.FlagSet) {
	// 参数不排序
	genCmd.SortFlags = false
}

// RunCommand run generate command
func RunCommand(cmd *cobra.Command, args []string) {
	
}
