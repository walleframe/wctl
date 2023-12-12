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
package protocol

import (
	"fmt"
	"os"

	"github.com/aggronmagi/wctl/protocol/ast"
	"github.com/aggronmagi/wctl/protocol/protobuf"
	"github.com/aggronmagi/wctl/protocol/wproto"
	"github.com/aggronmagi/wctl/utils"
)

type Parser interface {
	Parse(file string, input []byte) (prog *ast.YTProgram, err error)
}

type ParseFunc func(file string, input []byte) (prog *ast.YTProgram, err error)

func (f ParseFunc) Parse(file string, input []byte) (prog *ast.YTProgram, err error) {
	return f(file, input)
}

var _ Parser = ParseFunc(nil)

type astItem struct {
	FullName string
	FileName string
	// RelName  string
	Ast *ast.YTProgram
}

// 已分析文件仓库
type warehouse struct {
	full          map[string]*astItem
	file          map[string]*astItem
	path          string
	startWorkPath string
	parsers       map[string]Parser
}

// 全局仓库
var gWarehouse = &warehouse{
	full: make(map[string]*astItem),
	file: make(map[string]*astItem),
}

func init() {
	path, err := os.Getwd()
	if err != nil {
		utils.PanicIf(fmt.Errorf("获取运行目录出错.%w", err))
	}
	gWarehouse.startWorkPath = path
	gWarehouse.path = path
	gWarehouse.parsers = make(map[string]Parser)
}

// SetBasePath 设置基础目录. 所有输入将基于这个目录进行查找.(会将目录切换到输入目录)
func SetBasePath(path string) {
	gWarehouse.path = path
}

func RegisterParser(suffix string, parser Parser) {
	gWarehouse.parsers[suffix] = parser
}

func GetParser(suffix string) Parser {
	if val, ok := gWarehouse.parsers[suffix]; ok {
		return val
	}
	return nil
}

func init() {
	RegisterParser(".wproto", ParseFunc(wproto.Parse))
	RegisterParser(".proto", ParseFunc(protobuf.Parse))
}
