/*
Copyright © 2023 aggronmagi <czy463@163.com>

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
package ast

import "github.com/aggronmagi/wctl/protocol/token"

// 检查命名是否重定义
type ytCheck struct {
	data map[string]token.Pos
	no   map[int64]token.Pos
	// 	opt  map[string]string
}

func (ck *ytCheck) addUnionName(name string, pos token.Pos) {
	if ck.data == nil {
		ck.data = make(map[string]token.Pos)
	}
	ck.data[name] = pos
}

func (ck *ytCheck) checkUnionName(name string) (last token.Pos, repeat bool) {
	if v, ok := ck.data[name]; ok {
		repeat = true
		last = v
	}
	return
}

func (ck *ytCheck) addUionOption(name string, pos token.Pos) {
	if ck.data == nil {
		ck.data = make(map[string]token.Pos)
	}
	ck.data[name] = pos
}

func (ck *ytCheck) checkUnionOption(name string) (last token.Pos, repeat bool) {
	if v, ok := ck.data[name]; ok {
		repeat = true
		last = v
	}
	return
}

func (ck *ytCheck) checkUnionNo(no int64) (last token.Pos, repeat bool) {
	if v, ok := ck.no[no]; ok {
		repeat = true
		last = v
	}
	return
}

func (ck *ytCheck) addUnionNo(no int64, pos token.Pos) {
	if ck.no == nil {
		ck.no = make(map[int64]token.Pos)
	}
	ck.no[no] = pos
}
