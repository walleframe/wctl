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
package ast

import (
	"github.com/aggronmagi/wctl/protocol/token"
)

// 检查命名是否重定义
type ytCheck struct {
	pos  token.Pos
	data map[string]string
	opt  map[string]string
	no   map[int64]string
}

func (ck *ytCheck) addUnionName(name, tip string) {
	if ck.data == nil {
		ck.data = make(map[string]string)
	}
	ck.data[name] = tip
}

func (ck *ytCheck) checkUnionName(name string) (repeat bool, tip string) {
	if v, ok := ck.data[name]; ok {
		repeat = true
		tip = v
	}
	return
}

func (ck *ytCheck) addUionOption(name, tip string) {
	if ck.data == nil {
		ck.data = make(map[string]string)
	}
	ck.data[name] = tip
}

func (ck *ytCheck) checkUnionOption(name string) (repeat bool, tip string) {
	if v, ok := ck.data[name]; ok {
		repeat = true
		tip = v
	}
	return
}

func (ck *ytCheck) checkUnionNo(no int64) (repeat bool, tip string) {
	if v, ok := ck.no[no]; ok {
		repeat = true
		tip = v
	}
	return
}

func (ck *ytCheck) addUnionNo(no int64, tip string) {
	if ck.no == nil {
		ck.no = make(map[int64]string)
	}
	ck.no[no] = tip
}
