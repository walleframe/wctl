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
	"path/filepath"

	"github.com/aggronmagi/wctl/protocol/ast"
	"github.com/aggronmagi/wctl/utils"
)

func init() {
	// 注册递归解析函数.用于递归处理依赖
	ast.RegisterRecursionAnalyser = ast.RecursionAnalyseFunc(analyseOneFile)
}

// analyseOneFile 解析文件
func analyseOneFile(file string) (prog *ast.YTProgram, err error) {
	suffix := filepath.Ext(file)
	parser := GetParser(suffix)
	if parser == nil {
		err = fmt.Errorf("文件格式[%s]未注册！%s", suffix, file)
		return
	}

	// // import文件没有写后缀名. 强制必须写后缀名
	// if !strings.HasSuffix(file, ast.Flag.FileSuffix) {
	// 	err = fmt.Errorf("import 或者 指定的文件名不对. 必须包含[%s]后缀.(%s)", ast.Flag.FileSuffix, file)
	// 	return
	// 	// file = file + ast.Flag.FileSuffix
	// }
	// 查找仓库
	if item, ok := gWarehouse.file[file]; ok {
		return item.Ast, nil
	}
	// 转换绝对路径
	full, err := filepath.Abs(filepath.Join(gWarehouse.path, file))
	if err != nil {
		return
	}
	// 查找仓库
	if item, ok := gWarehouse.full[full]; ok {
		return item.Ast, nil
	}
	// 读取文件
	data, err := os.ReadFile(full)
	if err != nil {
		return
	}

	// 进行解析
	prog, err = parser.Parse(full, data)
	if err != nil {
		return
	}
	prog.File = file

	// 分析合理性
	err = prog.AnalyseProgram()
	if err != nil {
		return
	}

	// 保存
	item := &astItem{
		FullName: full,
		FileName: file,
		Ast:      prog,
	}
	gWarehouse.full[item.FullName] = item
	gWarehouse.file[item.FileName] = item

	return
}

// AnalyseFile 解析文件
func AnalyseFile(file string) (prog *ast.YTProgram, err error) {
	return analyseOneFile(file)
}

// AnlysePath 分析制定路径下所有文件
func AnlysePath(dir, ext string) (progs []*ast.YTProgram, err error) {
	var prog *ast.YTProgram
	var file string
	utils.RangeFilesWithExt(dir, ext, func(s string) error {
		file, err = filepath.Rel(dir, s)
		if err != nil {
			return err
		}
		prog, err = analyseOneFile(file)
		if err != nil {
			return err
		}
		progs = append(progs, prog)
		return nil
	})
	return
}
