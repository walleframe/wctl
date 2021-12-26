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
package builder

import (
	"github.com/aggronmagi/wctl/protocol/ast"
	"github.com/aggronmagi/wctl/utils"
)

type printGenerater struct{}

// Generate 生成代码接口
func (gen *printGenerater) Generate(prog *ast.YTProgram) (outs []*Output, err error) {
	utils.Dump(prog)
	return
}

// Union 唯一标识符 用于标识不同插件
func (gen *printGenerater) Union() string {
	return "printer"
}

func init() {
	RegisterGenerater(&printGenerater{})
}
