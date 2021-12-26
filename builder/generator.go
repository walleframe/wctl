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
	"fmt"
	"os"

	"github.com/aggronmagi/wctl/protocol/ast"
)

// Output 生成结果
type Output struct {
	File string
	Data []byte
}

// Generater go插件 代码生成接口
type Generater interface {
	// Generate 生成代码接口
	Generate(prog *ast.YTProgram) (outs []*Output, err error)
	// Union 唯一标识符 用于标识不同插件
	Union() string
}

// 代码生成器工厂
var factory = make(map[string]Generater)

// 生效的插件
var use []Generater

// EnableGenerator 生效内置生成器
func EnableGenerator(key string) (err error) {
	if gen, ok := factory[key]; ok {
		addUse(gen)
		return
	}
	err = fmt.Errorf("生成器: %s 不存在", key)
	return
}

// GetInnerGenerator 获取内置插件
func GetInnerGenerator() (list []string) {
	for k := range factory {
		list = append(list, k)
	}
	return
}

// RegisterGenerater 注册生成器
func RegisterGenerater(gen Generater) {
	if _, ok := factory[gen.Union()]; ok {
		fmt.Println("内置生成器重复:", gen.Union())
		os.Exit(1)
	}
	// 保存生成器
	factory[gen.Union()] = gen
}

// 生效插件
func addUse(gen Generater) {
	// 生效插件
	find := false
	for k, v := range use {
		// 防止同一插件多次调用
		if v.Union() == gen.Union() {
			find = true
			use[k] = gen
			break
		}
	}
	if !find {
		use = append(use, gen)
	}
}

// 内部生成器名称
const (
	// InnerTemplate  - 模板生成器
	InnerTemplate = "template"
	// InnerPrinter - 打印插件参数. 用于调试
	InnerPrinter = "printer"
)
