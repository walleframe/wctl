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
	"github.com/walleframe/wctl/commands/generate"

	"github.com/spf13/cobra"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:     "gen",
	Short:   "代码生成",
	Long:    generate.Help,
	Example: generate.Example,
	Version: generate.Version,
	Run:     generate.RunCommand,
}

func init() {
	rootCmd.AddCommand(genCmd)
	// 命令参数
	generate.Flags(genCmd.Flags())
}
