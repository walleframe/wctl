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
	"fmt"
	"strings"
)

// RecursionAnalyser 递归分析接口
type RecursionAnalyser interface {
	Analyse(file string) (prog *YTProgram, err error)
}

// RecursionAnalyseFunc 递归解析函数
type RecursionAnalyseFunc func(file string) (prog *YTProgram, err error)

// Analyse 实现 RecursionAnalyser 接口
func (f RecursionAnalyseFunc) Analyse(file string) (prog *YTProgram, err error) {
	return f(file)
}

var _ RecursionAnalyser = RecursionAnalyseFunc(nil)

// 本地全局.递归解析函数
var gRegisterRecursionAnalyser RecursionAnalyser

// RegisterRecursionAnalyser 注册递归解析函数
func RegisterRecursionAnalyser(act RecursionAnalyser) {
	gRegisterRecursionAnalyser = act
}

// 检查并修复文件引用合理性
func (prog *YTProgram) checkFixFileRefrence() (err error) {
	// 服务检查
	pb := false
	for _, v := range prog.Services {
		// 接口检查
		for _, mtd := range v.Methods {
			pb, err = mtd.checkMessageProtobuf(prog, fmt.Sprintf("not define in service <%s> methond <%s> ", v.Name, mtd.Name))
			if err != nil {
				return
			}
			if pb {
				continue
			}
			// 请求
			if mtd.Request != nil {
				err = prog.checkMsg(mtd.Request, fmt.Sprintf("not define in service <%s> methond <%s> request", v.Name, mtd.Name))
				if err != nil {
					return
				}
			}
			// 回复
			if mtd.Reply != nil {
				err = prog.checkMsg(mtd.Reply, fmt.Sprintf("not define in service <%s> methond <%s> reply", v.Name, mtd.Name))
				if err != nil {
					return
				}
			}
		}
	}
	// 消息检查
	for _, v := range prog.Messages {
		err = prog.checkMsg(v, fmt.Sprintf("not define in message <%s> ", v.Name))
		if err != nil {
			return
		}
	}
	return
}

// 检查消息内字段类型
func (prog *YTProgram) checkMsg(msg *YTMessage, tip string) (err error) {
	for _, field := range msg.Fields {
		if err = field.Type.checkType(prog, fmt.Sprintf("%s< $FILE:%d >in %s", tip, field.pos.Line, msg.Name)); err != nil {
			return
		}
	}
	return
}

// 字段类型检查.自定义类型
func (typ *YTFieldType) checkType(prog *YTProgram, tip string) (err error) {

	// 忽略基础类型
	if typ.YTBaseType != nil {
		return
	}
	// list 类型
	if typ.YTListType != nil {
		if typ.YTListType.YTBaseType != nil {
			return
		}
		err = typ.YTListType.YTCustomType.checkCustom(prog, tip)
		return
	}
	if typ.YTMapTypee != nil {
		if typ.YTMapTypee.Value.YTBaseType != nil {
			return
		}
		err = typ.YTMapTypee.Value.YTCustomType.checkCustom(prog, tip)
		return
	}
	err = typ.YTCustomType.checkCustom(prog, tip)
	return
}

// 检查自定义类型是否在本地文件包含
func (cst *YTCustomType) checkCustom(prog *YTProgram, tip string) (err error) {
	// 不包含"." 不是是当前文件
	if !strings.Contains(cst.Name, ".") {
		// 当前文件. 直接查找
		if msg, ok := prog.msgMap[cst.Name]; ok {
			cst.Msg = msg
		} else {
			err = fmt.Errorf("custom type [%s] %s", cst.Name, tip)
		}
		return
	}
	// 切分
	list := strings.Split(cst.Name, ".")
	// len(list) == 2
	// 查找引用包
	refName := list[0]
	iprogs, ok := prog.impMap[refName]
	if !ok {
		err = fmt.Errorf("import cutsom type [%s] %s", cst.Name, tip)
		return
	}
	// 查找类型
	stName := list[1]
	// 遍历引用包
	for _, iprog := range iprogs {
		if msg, ok := iprog.msgMap[stName]; ok {
			cst.Msg = msg
			err = nil
			return
		}
		err = fmt.Errorf("import cutsom type [%s] %s(from ref[%s] pkg[%s] struct[%s])", cst.Name, tip,
			refName, iprog.Pkg.Name, stName,
		)
	}
	return
}

// 检测修复pb类型
func (method *YTMethod) checkMessageProtobuf(prog *YTProgram, tip string) (pb bool, err error) {
	if method.Request != nil && method.Request.protobuf {
		pb = true
		method.Request, err = getProtoMsg(prog, method.Request.Name, tip)
		if err != nil {
			return
		}
	}
	if method.Reply != nil && method.Reply.protobuf {
		pb = true
		method.Reply, err = getProtoMsg(prog, method.Reply.Name, tip)
		if err != nil {
			return
		}
	}
	return
}
func getProtoMsg(prog *YTProgram, name, tip string) (result *YTMessage, err error) {
	// 不包含"." 不是是当前文件
	if !strings.Contains(name, ".") {
		// 当前文件. 直接查找
		if msg, ok := prog.msgMap[name]; ok {
			result = msg
		} else {
			err = fmt.Errorf("custom type [%s] %s", name, tip)
		}
		return
	}
	// 切分
	list := strings.Split(name, ".")
	// len(list) == 2
	// 查找引用包
	refName := list[0] // 查找类型
	stName := list[1]

	for _, iprogs := range prog.impMap {
		// 遍历引用包
		for _, iprog := range iprogs {
			if refName != iprog.Pkg.Name {
				continue
			}
			if msg, ok := iprog.msgMap[stName]; ok {
				// 复制一份数据.并修改名称
				result = &YTMessage{}
				*result = *msg
				result.Name = name
				err = nil
				return
			}
		}
	}

	err = fmt.Errorf("import cutsom type [%s] %s. not found define", name, tip)
	return
}
