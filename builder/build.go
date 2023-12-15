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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/walleframe/wctl/protocol/ast"
	"github.com/walleframe/wctl/utils"
	"go.uber.org/multierr"
)

// Build 生成代码
func Build(progs []*ast.YTProgram, outPath string, merge bool) (err error) {
	if len(use) < 1 {
		fmt.Println("未使用任何生成器. 内置生成器:", GetInnerGenerator())
		return
	}
	var mergeCache map[string]*mergeData
	mergeCache = make(map[string]*mergeData, len(progs)*10)
	var outFile string
	for _, prog := range progs {
		for _, gen := range use {
			outs, err := gen.Generate(prog)
			if err != nil {
				err = fmt.Errorf("generate [%s] %s failed. \n%s", gen.Union(), prog.File, err.Error())
				return err
			}
			for _, v := range outs {
				if v.File == "" {
					ne := fmt.Errorf("generate [%s] %s failed. output file empty. len(%d)",
						gen.Union(), prog.File, len(v.Data))
					log.Println(ne)
					err = multierr.Append(err, ne)
					continue
				}
				outFile = filepath.Clean(filepath.Join(outPath, v.File))
				checkDir(outFile)
				if merge {
					last, ok := mergeCache[outFile]
					if !ok {
						fmt.Println(gen.Union(), prog.File, "==>", v.File)
						mergeCache[outFile] = &mergeData{
							datas: []*mergeFile{{
								lastUnion:  gen.Union(),
								lastSource: prog.File,
								data:       v.Data,
							}},
						}
						err = ioutil.WriteFile(outFile, v.Data, 0644)
					} else {
						fmt.Println(gen.Union(), prog.File, " rewrite ==>", v.File)
						last.datas = append(last.datas, &mergeFile{
							lastUnion:  gen.Union(),
							lastSource: prog.File,
							data:       v.Data,
						})
						err = ioutil.WriteFile(outFile, mergeFileData(last), 0644)
					}
				} else {
					fmt.Println(gen.Union(), prog.File, "==>", v.File)
					// 覆盖重写检测
					if overwrite, ok := mergeCache[outFile]; ok {
						ow := overwrite.datas[0]
						err = fmt.Errorf("generate [%s] %s ==> %s will overwrite [%s] %s generate output",
							gen.Union(), prog.File, v.File,
							ow.lastUnion, ow.lastSource,
						)
						return err
					}
					// 缓存信息 (不保存 文件数据信息)
					mergeCache[outFile] = &mergeData{
						datas: []*mergeFile{{
							lastUnion:  gen.Union(),
							lastSource: prog.File,
						}},
					}
					err = ioutil.WriteFile(outFile, v.Data, 0644)
				}
				if err != nil {
					err = fmt.Errorf("generate [%s] %s save %s \n%s", gen.Union(), prog.File, v.File, err.Error())
					return err
				}
				if utils.ShowDetail() {
					fmt.Println("data:", string(v.Data))
				}
			}
		}
	}
	return
}

func checkDir(pn string) {
	dir := filepath.Dir(filepath.Clean(pn))
	//fmt.Println(dir)
	os.MkdirAll(dir, 0755) // os.ModeDir) // /0666)
}

type mergeFile struct {
	lastUnion  string
	lastSource string
	data       []byte
}

type mergeData struct {
	datas []*mergeFile
}

func mergeFileData(in *mergeData) (data []byte) {
	srcs := in.datas
	l := 0
	for _, v := range srcs {
		l += len(v.data)
	}
	data = make([]byte, 0, l)
	for _, v := range srcs {
		data = append(data, v.data...)
	}
	return
}
