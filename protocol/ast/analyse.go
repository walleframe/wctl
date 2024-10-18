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
var RegisterRecursionAnalyser RecursionAnalyser = nil

// AnalyseProgram 分析检测Program合理性
func (prog *YTProgram) AnalyseProgram() (err error) {
	// 检测import合理性
	err = prog.checkImport()
	if err != nil {
		return
	}

	// 重复定义检测
	err = prog.checkRepeatedDefine()
	if err != nil {
		return
	}

	// 检查并修复文件引用合理性
	err = prog.checkFixFileRefrence()
	if err != nil {
		return
	}

	return
}

// 检测import合理性
func (prog *YTProgram) checkImport() (err error) {
	imp := make(map[string]*YTImport)
	for _, v := range prog.Imports {
		// 导入文件名重复检测
		if last, ok := imp[v.File]; ok {
			return NewErrorPos(v.DefPos, "import file [%s] repeated with %s", v.File, last.DefPos.String())
		}
		imp[v.File] = v
		// 导入别名检测
		if v.AliasName == "" {
			continue
		}
		if last, ok := imp[v.AliasName]; ok {
			return NewErrorPos(v.DefPos, "import alias [%s] repeated with %s", v.AliasName, last.DefPos.String())
		}
		imp[v.AliasName] = v
	}

	prog.impMap = make(map[string][]*YTProgram) // 改为slice结构. 允许多个文件作为同一个包
	for _, val := range prog.Imports {
		refName := val.AliasName
		if refName == "" {
			refName = val.Prog.Pkg.Name
		}
		if last, ok := prog.impMap[refName]; ok {
			for _, v := range last {
				// 相同引用名,必须相同包名
				if v.Pkg.Name != val.Prog.Pkg.Name {
					return NewErrorPos(val.DefPos, "import invalid. import package name do not same. <import %s %s> pkg[%s] != pkg[%s] refName[%s]",
						val.AliasName, val.File, val.Prog.Pkg.Name,
						v.Pkg.Name, refName)
				}
			}
		}

		prog.impMap[refName] = append(prog.impMap[refName], val.Prog)
		//prog.impMap[val.Prog.Pkg.Name] = append(prog.impMap[val.Prog.Pkg.Name], val.Prog)
	}
	return
}

// 重复定义检测
func (prog *YTProgram) checkRepeatedDefine() (err error) {
	for _, val := range prog.YTOptions.Opts {
		if last, ok := prog.checkUnionOption(val.Key); ok {
			return NewErrorPos(val.DefPos, "pakage level option define name repeated [%s] %s", val.Key, last.String())
		}
		prog.addUionOption(val.Key, val.DefPos)
	}

	for _, val := range prog.EnumDefs {
		if last, ok := prog.checkUnionName(val.Name); ok {
			return NewErrorPos(val.DefPos, "enum define name repeated [%s] %s", val.Name, last.String())
		}
		prog.addUnionName(val.Name, val.DefPos)

		for _, ev := range val.Values {
			if last, ok := val.checkUnionNo(ev.Value); ok {
				return NewErrorPos(ev.DefPos, "enum value name repeated [%s.%s] %s", val.Name, ev.Name, last.String())
			}
			val.addUnionNo(ev.Value, ev.DefPos)
		}

		for _, opt := range val.YTOptions.Opts {
			if last, ok := prog.checkUnionOption(opt.Key); ok {
				return NewErrorPos(opt.DefPos, "enum option define name repeated [%s %s] %s", val.Name, opt.Key, last.String())
			}
			prog.addUionOption(opt.Key, opt.DefPos)
		}
	}

	prog.msgMap = make(map[string]*YTMessage)
	for _, val := range prog.Messages {
		prog.msgMap[val.Name] = val
		err = prog.checkMsgRepeatedDefine(val)
		if err != nil {
			return
		}
	}

	for _, val := range prog.Services {
		if last, ok := prog.checkUnionName(val.Name); ok {
			return NewErrorPos(val.DefPos, "service define name repeated [%s] %s", val.Name, last.String())
		}
		prog.addUnionName(val.Name, val.DefPos)
		// method
		for _, method := range val.Methods {
			if last, ok := val.checkUnionName(method.Name); ok {
				return NewErrorPos(method.DefPos, "service method name repeated [%s] %s", val.Name, last.String())
			}
			val.addUnionName(method.Name, method.DefPos)
		}
		// option
		for _, opt := range val.YTOptions.Opts {
			if last, ok := val.checkUnionName(opt.Key); ok {
				return NewErrorPos(opt.DefPos, "service option name repeated [%s] %s", val.Name, last.String())
			}
			val.addUionOption(opt.Key, opt.DefPos)
		}
		// method no
		if Flag.ServiceUseMethodID {
			for _, method := range val.Methods {
				if last, ok := val.checkUnionNo(*method.No.Value); ok {
					return NewErrorPos(method.DefPos, "service method id repeated [%s] %s", val.Name, last.String())
				}
				val.addUnionNo(*method.No.Value, method.DefPos)
			}
		}
	}

	for _, val := range prog.Projects {
		// if last, ok := prog.checkUnionName(val.Name); ok {
		// 	return NewErrorPos(val.DefPos, "enum define name repeated [%s] %s", val.Name, last.String())
		// }
		// prog.addUnionName(val.Name, val.DefPos)

		for area, opts := range val.Conf {
			check := ytCheck{}
			for _, opt := range opts.Opts {
				if last, ok := check.checkUnionOption(opt.Key); ok {
					return NewErrorPos(opt.DefPos,
						"project %s area %s option name repeated [%s] %s",
						val.Name, area, opt.Key, last.String(),
					)
				}
			}
		}
	}
	return
}

func (prog *YTProgram) checkMsgRepeatedDefine(val *YTMessage) error {
	if last, ok := prog.checkUnionName(val.Name); ok {
		return NewErrorPos(val.DefPos, "message define name repeated [%s] %s", val.Name, last.String())
	}
	prog.addUnionName(val.Name, val.DefPos)
	//
	for _, field := range val.Fields {
		//fmt.Println("val:%t field:%t", val != nil, field != nil)
		if last, ok := val.checkUnionNo(int64(field.No)); ok {
			return NewErrorPos(field.DefPos, "message field id repeated [%s.%s] %s", val.Name, field.Name, last.String())
		}
		val.addUnionNo(int64(field.No), field.DefPos)

		if last, ok := val.checkUnionName(field.Name); ok {
			return NewErrorPos(field.DefPos, "message field name repeated [%s.%s] %s", val.Name, field.Name, last.String())
		}
		val.addUnionName(field.Name, field.DefPos)
	}

	for _, opt := range val.YTOptions.Opts {
		if last, ok := val.checkUnionOption(opt.Key); ok {
			return NewErrorPos(opt.DefPos, "message option name repeated [%s %s] %s", val.Name, opt.Key, last.String())
		}
		val.addUionOption(opt.Key, opt.DefPos)
	}
	//
	for _, sub := range val.SubMsgs {
		if err := prog.checkMsgRepeatedDefine(sub); err != nil {
			return err
		}
		prog.msgMap[sub.Name] = sub
	}

	return nil
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
		if field.Type.YTCustomType != nil {
			find := false
			for _, v := range msg.SubEnums {
				if v.Name == field.Type.Name {
					find = true
					break
				}
			}
			if find {
				continue
			}
			for _, v := range msg.SubMsgs {
				if v.Name == field.Type.Name {
					find = true
					break
				}
			}
			if find {
				continue
			}
		}
		if err = field.Type.checkType(prog, fmt.Sprintf("%s< %s >in %s", tip, field.DefPos.String(), msg.Name)); err != nil {
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
			return
		}
		// 查找enum
		for _, def := range prog.EnumDefs {
			if def.Name == cst.Name {
				return
			}
		}
		err = fmt.Errorf("custom type [%s] %s", cst.Name, tip)
		return
	}
	// 切分
	list := strings.Split(cst.Name, ".")
	// len(list) == 2
	// 查找引用包
	refName := list[0]
	iprogs, ok := prog.impMap[refName]
	if !ok {
		// 兼容protobuf. 包名是当前的包名.
		if refName == prog.Pkg.Name {
			cst.Name = list[1]
			if err = cst.checkCustom(prog, tip); err == nil {
				return
			}
			cst.Name = strings.Join(list, ".")
		}
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
	if method.Request != nil && method.Request.ProtobufFlag {
		pb = true
		method.Request, err = getProtoMsg(prog, method.Request.Name, tip)
		if err != nil {
			return
		}
	}
	if method.Reply != nil && method.Reply.ProtobufFlag {
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
