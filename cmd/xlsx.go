/*
   Copyright © 2020 aggronmagi <czy463@163.com>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cmd

import (
	"github.com/walleframe/wctl/commands/xlsxgen"

	"github.com/spf13/cobra"
)

// xlsxCmd represents the xlsx command
var xlsxCmd = &cobra.Command{
	Use:     "xlsx",
	Short:   "xlsx配置生成",
	Long:    xlsxgen.Help,
	Example: xlsxgen.Example,
	Version: xlsxgen.Version,
	Run:     xlsxgen.RunCommand,
}

func init() {
	rootCmd.AddCommand(xlsxCmd)
	// 命令参数
	xlsxgen.Flags(xlsxCmd.Flags())
}
