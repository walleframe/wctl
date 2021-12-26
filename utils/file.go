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
package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

// RangeFiles 遍历获取指定目录下的所有文件(递归深层目录)
func RangeFiles(dirPth string, rf func(file string) error) (err error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			err = RangeFiles(dirPth+PthSep+fi.Name(), rf)
			if err != nil {
				return
			}
		} else {
			// 过滤指定格式
			err = rf(dirPth + PthSep + fi.Name())
			if err != nil {
				return
			}
		}
	}

	return
}

// RangeFilesWithExt 遍历指定目录下所有 有指定后缀名的文件
func RangeFilesWithExt(dir, ext string, rf func(string) error) error {
	return RangeFiles(dir, func(file string) error {
		if !strings.HasSuffix(file, ext) {
			return nil
		}
		return rf(file)
	})
}
