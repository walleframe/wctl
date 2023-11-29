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
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/aggronmagi/wctl/protocol/token"
	"github.com/aggronmagi/wctl/utils"
)

var flag = utils.Flag

func debug(vals ...interface{}) {
	if !flag.Debug && !flag.ShowDetail {
		return
	}
	if !flag.ShowDetail {
		fmt.Println(vals...)
		return
	}
	if len(vals) > 0 {
		fmt.Println(vals[0])
	}
	for k := 1; k < len(vals); k++ {
		if v, ok := vals[k].(interface {
			IDValue() string
		}); ok {
			fmt.Printf("\t%s %T %#v\n", v.IDValue(), v, v)
		} else {
			fmt.Print("\t", utils.Sdump(vals[k], ""), "\n")
		}
	}
	fmt.Println()
}

// CheckProgram 检查Program合理性
// Start: FileElements
func CheckProgram(v0 interface{}) (prog *YTProgram, err error) {
	debug("CheckProgram", v0)
	prog = v0.(*YTProgram)
	// 检查并修复文件引用合理性
	err = prog.checkFixFileRefrence()
	if err != nil {
		return nil, err
	}

	return
}

// NewFileElements1 新建prog
// FileElements: empty
func NewFileElements1() (*YTProgram, error) {
	debug("NewFileElements1")
	return &YTProgram{
		impMap: make(map[string][]*YTProgram),
		msgMap: make(map[string]*YTMessage),
	}, nil
}

// AppendFileElements xxx
// FileElements: FileElements Doc Element
func AppendFileElements(v0, v1, v2 interface{}) (prog *YTProgram, err error) {
	debug("AppendFileElements", v0, v1, v2)
	var doc *YTDoc
	if v1 != nil {
		doc = v1.(*YTDoc)
	}
	prog = v0.(*YTProgram)
	// 先检查包定义
	switch val := v2.(type) {
	case *YTPackage: // 包定义
		if prog.Pkg != nil {
			return nil, errors.New("package 重定义")
		}
		prog.Pkg = val
		prog.Pkg.YTDoc = doc
		return
	}
	// package 必须在文件开头
	if prog.Pkg == nil {
		// pb 文件, syntax使用option保存
		if prog.YTOptions.GetOption(ProtobufOptionSyntax) == nil {
			return nil, errors.New("package 未定义.package 应该在文件起始处定义")
		}
	}

	switch val := v2.(type) {
	case *YTPackage: // 包定义
	case *YTImport: // 依赖,导入文件
		if doc != nil {
			val.YTDoc = doc
		}
		for _, v := range prog.Imports {
			if v.File == val.File {
				return nil, fmt.Errorf("import无效.重复import同一文件.%s", val.File)
			}
			if val.AliasName != "" && v.AliasName == val.AliasName {
				return nil, fmt.Errorf("import无效.别名重复.%s[%s,%s]", val.AliasName, v.File, val.File)
			}
		}
		refName := val.AliasName
		if refName == "" {
			refName = val.Prog.Pkg.Name
		}
		// 改为slice结构. 允许多个文件作为同一个包
		if last, ok := prog.impMap[refName]; ok {
			for _, v := range last {
				// 相同引用名,必须相同包名
				if v.Pkg.Name != val.Prog.Pkg.Name {
					return nil, fmt.Errorf("import无效.引用包的包名不一样.<import %s %s> pkg[%s] != pkg[%s] refName[%s]",
						val.AliasName, val.File, val.Prog.Pkg.Name,
						v.Pkg.Name, refName)
				}
			}
		}
		prog.Imports = append(prog.Imports, val)
		prog.impMap[refName] = append(prog.impMap[refName], val.Prog)
	case *YTOption: // 包级选项
		if doc != nil {
			val.YTDoc = doc
		}
		if find, tip := prog.checkUnionOption(val.Key); find {
			return nil, fmt.Errorf("包级option定义无效,重复定义.%s %s", val.Key, tip)
		}
		prog.Opts = append(prog.Opts, val)
		prog.addUionOption(val.Key, "")
	case *YTEnumDef: // 枚举定义
		if doc != nil {
			val.YTDoc = doc
		}
		if find, tip := prog.checkUnionName(val.Name); find {
			return nil, fmt.Errorf("Enum定义无效,重复定义.%s %s", val.Name, tip)
		}
		prog.EnumDefs = append(prog.EnumDefs, val)
		prog.addUnionName(val.Name, fmt.Sprintf("last define on line:%d", val.pos.Line))
	case *YTMessage: // 消息定义
		if doc != nil {
			val.YTDoc = doc
		}
		if find, tip := prog.checkUnionName(val.Name); find {
			return nil, fmt.Errorf("message 定义无效,重复定义.%s %s", val.Name, tip)
		}
		prog.Messages = append(prog.Messages, val)
		prog.addUnionName(val.Name, fmt.Sprintf("last define on line:%d", val.pos.Line))
		prog.msgMap[val.Name] = val
	case *YTService: // 服务定义
		if doc != nil {
			val.YTDoc = doc
		}
		if find, tip := prog.checkUnionName(val.Name); find {
			return nil, fmt.Errorf("service定义无效,重复定义.%s %s", val.Name, tip)
		}
		prog.Services = append(prog.Services, val)
		prog.addUnionName(val.Name, fmt.Sprintf("last define on line:%d", val.pos.Line))
	case *YTProject: // 项目定义
		if doc != nil {
			val.YTDoc = doc
		}
		// if find, tip := prog.checkUnionName(val.Name); find {
		// 	return nil, fmt.Errorf("service定义无效,重复定义.%s %s", val.Name, tip)
		// }
		prog.Projects = append(prog.Projects, val)
		// prog.addUnionName(val.Name, fmt.Sprintf("last define on line:%d", val.pos.Line))
	default:
		return nil, fmt.Errorf("未知定义. %T %#v %s", v2, v2, utils.Sdump(v2, "unkown"))
	}
	return
}
func strptr(sv string) *string {
	return &sv
}
func intptr(v int64) *int64 {
	return &v
}

// NewProject 新建项目
// "project" tok_identifier "{" ProjElements "}"
func NewProject(v1, v3 interface{}) (proj *YTProject, _ error) {
	debug("NewProject", v1, v3)
	if v3 != nil {
		proj = v3.(*YTProject)
	} else {
		proj = &YTProject{}
	}
	proj.Name = v1.(*token.Token).IDValue()
	return
}

// NewProjectEmpty 初始化project
// ProjElements: empty
func NewProjectEmpty() (proj *YTProject, _ error) {
	proj = &YTProject{}
	proj.check = make(map[string]*ytCheck)
	proj.Conf = make(map[string]*YTOptions)
	return
}

// ChangeProjectArea 更改area
// ProjElements Doc ProjArea
func ChangeProjectArea(v0, v1, v2 interface{}) (proj *YTProject, _ error) {
	debug("ChangeProjectArea", v0, v1, v2)
	if v0 != nil {
		proj = v0.(*YTProject)
	} else {
		proj = &YTProject{}
	}
	proj.area = v2.(*token.Token).IDValue()
	// if v1 != nil {
	// 	opt.YTDoc = v1.(*YTDoc)
	// }
	return
}

// AppendProjectOption 附加选项
// ProjElements Doc Option
func AppendProjectOption(v0, v1, v2 interface{}) (proj *YTProject, _ error) {
	debug("AppendProjectOption", v0, v1, v2)
	if v0 != nil {
		proj = v0.(*YTProject)
	} else {
		proj = &YTProject{}
	}
	opt := v2.(*YTOption)
	if v1 != nil {
		opt.YTDoc = v1.(*YTDoc)
	}
	if last, ok := proj.check[proj.area]; ok {
		if find, tip := last.checkUnionOption(opt.Key); find {
			return nil, fmt.Errorf("project aread [%s] option repeated. %s %s", proj.area, opt.Key, tip)
		}
	} else {
		proj.check[proj.area] = &ytCheck{}
	}
	if proj.Conf[proj.area] == nil {
		proj.Conf[proj.area] = &YTOptions{}
	}
	proj.Conf[proj.area].Opts = append(proj.Conf[proj.area].Opts, opt)
	proj.check[proj.area].addUionOption(opt.Key, fmt.Sprintf(",last define on line:%d", opt.pos.Line))
	return
}

// NewMessageByFieldType 通过类型建消息
// MethodArgs: FiledType
// MethodRet: FiledType
func NewMessageByFieldType(v0 interface{}) (msg *YTMessage, _ error) {
	debug("NewMessageByFieldType", v0)
	ft := v0.(*YTFieldType)
	msg = &YTMessage{}
	field := &YTField{}
	// field.Name = "request"
	field.Type = ft
	msg.Fields = append(msg.Fields, field)
	msg.pos = field.pos
	return
}

// NewMessageByFeilds 新建message. 保证不能为nil
// MethodArgs: MethodArgFields
// MethodRet: MethodRetArgs
func NewMessageByFeilds(v0 interface{}) (msg *YTMessage, _ error) {
	debug("NewMessageByFeilds", v0)
	if v0 == nil {
		return nil, fmt.Errorf("参数为空,必须显示声明参数是void")
	}
	return v0.(*YTMessage), nil
}

// NewMethod 新建方法
// tok_identifier "(" MethodArgs ")" MethodRet MethodNo AddtionOption OptEnd
func NewMethod(v0, v2, v4, v5, v6 interface{}) (method *YTMethod, _ error) {
	method = &YTMethod{}
	method.Name = v0.(*token.Token).IDValue()
	method.pos = v0.(*token.Token).Pos
	if v2 != nil {
		if msg, ok := v2.(*YTMessage); ok {
			method.Request = msg
		} else if tok, ok := v2.(*token.Token); ok {
			method.Request = &YTMessage{
				Name:     tok.IDValue(),
				protobuf: true,
			}
		} else {
			return nil, fmt.Errorf("invalid service method [%s] request[%#v].  line:%d", method.Name, v2, method.pos.Line)
		}
	}
	if v4 != nil {
		if msg, ok := v4.(*YTMessage); ok {
			method.Reply = msg
		} else if tok, ok := v4.(*token.Token); ok {
			method.Reply = &YTMessage{
				Name:     tok.IDValue(),
				protobuf: true,
			}
		} else {
			return nil, fmt.Errorf("invalid service method [%s] response[%#v]. line:%d", method.Name, v4, method.pos.Line)
		}
	}
	if Flag.ServiceUseMethodID {
		if v5 != nil {
			method.No = v5.(*YTMethodNo)
		} else {
			return nil, fmt.Errorf("not set service method id. %s line:%d", method.Name, method.pos.Line)
		}
	}

	if v6 != nil {
		method.Opts = v6.([]*YTOption)
	}
	return
}

// NewMethodNo 新建
func NewMethodNo(num, str interface{}) (no *YTMethodNo, _ error) {
	debug("NewMe", num, str)
	no = &YTMethodNo{}

	if num != nil {
		val, err := num.(*token.Token).Int64Value()
		if err != nil {
			return nil, err
		}
		no.Value = &val
	} else if str != nil {
		// TODO 服务请求使用ID
		// TODO 服务请求ID使用枚举或者常量
		return nil, fmt.Errorf("now not support service method id set by enum or const value. %s", str.(*token.Token).IDValue())
		// no.Macro = &YTCustomType{
		// 	Name: str.(*token.Token).IDValue(),
		// }
	}

	return
}

// NewService 新建服务
// Service:	"service" tok_identifier "{" ServiceElements "}" OptEnd
func NewService(v1, v3 interface{}) (svr *YTService, _ error) {
	debug("NewService", v1, v3)
	if v3 != nil {
		svr = v3.(*YTService)
	} else {
		svr = &YTService{}
	}
	svr.Name = v1.(*token.Token).IDValue()
	svr.pos = v1.(*token.Token).Pos
	return
}

// AppendServiceOption 附加选项
// ServiceElements: ServiceElements Doc Option
func AppendServiceOption(v0, v1, v2 interface{}) (svr *YTService, _ error) {
	debug("AppendServiceOption", v0, v1, v2)
	if v0 != nil {
		svr = v0.(*YTService)
	} else {
		svr = &YTService{}
	}
	opt := v2.(*YTOption)
	if v1 != nil {
		opt.YTDoc = v1.(*YTDoc)
	}
	if find, tip := svr.checkUnionOption(opt.Key); find {
		return nil, fmt.Errorf("service option repeated. %s %s", opt.Key, tip)
	}
	svr.Opts = append(svr.Opts, opt)
	svr.addUionOption(opt.Key, fmt.Sprintf(",last define on line:%d", opt.pos.Line))
	return
}

// ChangeServiceMethodFlag 修改方法标记
// ServiceElements: ServiceElements Doc MethodFlag
func ChangeServiceMethodFlag(v0, v1, v2 interface{}) (svr *YTService, _ error) {
	debug("AppendServiceMethod", v0, v1, v2)
	if v0 != nil {
		svr = v0.(*YTService)
	} else {
		svr = &YTService{}
	}
	flag := v2.(MethodFlag)
	// if v1 != nil {
	// 	mtd.YTDoc = v1.(*YTDoc)
	// }
	svr.flag = flag
	return
}

// AppendServiceMethod 追加服务方法
// ServiceElements: ServiceElements Doc Method
func AppendServiceMethod(v0, v1, v2 interface{}) (svr *YTService, _ error) {
	debug("AppendServiceMethod", v0, v1, v2)
	if v0 != nil {
		svr = v0.(*YTService)
	} else {
		svr = &YTService{}
	}
	mtd := v2.(*YTMethod)
	if v1 != nil {
		mtd.YTDoc = v1.(*YTDoc)
	}
	mtd.Flag = svr.flag
	if find, tip := svr.checkUnionName(mtd.Name); find {
		return nil, fmt.Errorf("service method name repeated. %s %s", mtd.Name, tip)
	}
	// TODO 服务请求使用ID
	if Flag.ServiceUseMethodID {
		if mtd.No == nil {
			return nil, fmt.Errorf("service method id not set. %s $FILE:%d", mtd.Name, mtd.pos.Line)
		}
		// 解析宏或者引用
		if find, tip := svr.checkUnionNo(*mtd.No.Value); find {
			return nil, fmt.Errorf("service method id(%d) repeated. %s %s", *mtd.No.Value, mtd.Name, tip)
		}
	}
	svr.Methods = append(svr.Methods, mtd)
	svr.addUnionName(mtd.Name, fmt.Sprintf(",last define on $FILE:%d", mtd.pos.Line))
	// TODO 服务请求使用ID
	if Flag.ServiceUseMethodID {
	}
	return
}

// NewMessage 新建消息
// Message: "message" tok_identifier "{" MessageElements "}" OptEnd
func NewMessage(v1, v3 interface{}) (msg *YTMessage, _ error) {
	debug("NewMessage", v1, v3)
	if v3 != nil {
		msg = v3.(*YTMessage)
	} else {
		msg = &YTMessage{}
	}
	tok := v1.(*token.Token)
	msg.Name = tok.IDValue()
	msg.pos = tok.Pos
	return
}

// AppendMessageField 追加字段
// MessageElements: MessageElements Doc Field
// MessageElements: MethodArgFields Doc Field
func AppendMessageField(v0, v1, v2 interface{}) (msg *YTMessage, _ error) {
	debug("AppendMessageField", v0, v1, v2)
	if v0 != nil {
		msg = v0.(*YTMessage)
	} else {
		msg = &YTMessage{}
	}
	field := v2.(*YTField)
	if v1 != nil {
		field.YTDoc = v1.(*YTDoc)
	}
	if find, tip := msg.checkUnionName(field.Name); find {
		return nil, fmt.Errorf("message field name repeated. %s %s", field.Name, tip)
	}
	if find, tip := msg.checkUnionNo(int64(field.No)); find {
		return nil, fmt.Errorf("message field id(%d) repeated. %s %s", field.No, field.Name, tip)
	}
	msg.Fields = append(msg.Fields, field)
	msg.addUnionName(field.Name, fmt.Sprintf(",last define on $FILE:%d", field.pos.Line))
	msg.addUnionNo(int64(field.No), fmt.Sprintf(",last define on $FILE:%d", field.pos.Line))
	return
}

// AppendMessageOption 追加选项
// MessageElements:MessageElements Doc Option
func AppendMessageOption(v0, v1, v2 interface{}) (msg *YTMessage, _ error) {
	debug("AppendMessageOption", v0, v1, v2)
	if v0 != nil {
		msg = v0.(*YTMessage)
	} else {
		msg = &YTMessage{}
	}
	field := v2.(*YTOption)
	if v1 != nil {
		field.YTDoc = v1.(*YTDoc)
	}
	if find, tip := msg.checkUnionOption(field.Key); find {
		return nil, fmt.Errorf("message option repeated. %s %s", field.Key, tip)
	}
	msg.Opts = append(msg.Opts, field)
	msg.addUionOption(field.Key, fmt.Sprintf(",last define on $FILE:%d", field.pos.Line))
	return
}

// NewField 新建字段
// Field: tok_const_int ":" FiledType tok_identifier AddtionOption OptEnd
func NewField(v0, v2, v3, v4 interface{}) (filed *YTField, _ error) {
	debug("NewField", v0, v2, v3, v4)
	filed = &YTField{}
	filed.Type = v2.(*YTFieldType)
	filed.Name = v3.(*token.Token).IDValue()
	if v4 != nil {
		filed.Opts = v4.([]*YTOption)
	}
	val, err := v0.(*token.Token).Int32Value()
	if err != nil {
		return nil, err
	}
	if val > math.MaxUint8 || val < 0 {
		return nil, fmt.Errorf("字段[%s]序号超出限制(%d)", filed.Name, val)
	}
	filed.No = uint8(val)
	filed.pos = v0.(*token.Token).Pos
	return
}

// AppendFiledOptions 追加选项
// FieldOption:	FieldOption Doc Option
func AppendFiledOptions(v0, v1, v2 interface{}) (opts []*YTOption, _ error) {
	debug("AppendFiledOptions", v0, v1, v2)
	if v0 != nil {
		opts = v0.([]*YTOption)
	}
	opt := v2.(*YTOption)
	if v1 != nil {
		opt.YTDoc = v1.(*YTDoc)
	}
	// 重复检查
	for _, v := range opts {
		if v.Key == opt.Key {
			return nil, fmt.Errorf("field option repeated. %s last define on $FILE:%d", opt.Key, v.pos.Line)
		}
	}
	opts = append(opts, opt)
	return
}

// NewFieldTypeBase 新建基础类型
// ContainerElemType: BaseType
func NewFieldTypeBase(v0 interface{}) (*YTFieldType, error) {
	debug("NewFieldTypeBase", v0)
	return &YTFieldType{
		YTBaseType: v0.(*YTBaseType),
	}, nil
}

// NewFieldTypeCustom 新建自定类型
// CustomType:
// tok_identifier	<< ast.NewFieldTypeCustom($0) >>
// tok_option		<< ast.NewFieldTypeCustom($0) >>
func NewFieldTypeCustom(v0 interface{}) (*YTFieldType, error) {
	debug("NewFieldTypeCustom", v0)
	return &YTFieldType{
		YTCustomType: &YTCustomType{
			Name: v0.(*token.Token).IDValue(),
		},
	}, nil
}

// NewFieldTypeList 新建列表类型
// ListType: "[" "]" ContainerElemType
func NewFieldTypeList(v2 interface{}) (*YTFieldType, error) {
	debug("NewFieldTypeList", v2)
	elemType := v2.(*YTFieldType)
	return &YTFieldType{
		YTListType: &YTListType{
			YTBaseType:   elemType.YTBaseType,
			YTCustomType: elemType.YTCustomType,
		},
	}, nil
}

// NewFieldTypeMap 新建字典类型
// MapType:	"map" "[" BaseType "]" ContainerElemType
func NewFieldTypeMap(v2, v4 interface{}) (*YTFieldType, error) {
	debug("NewFieldTypeMap", v2, v4)
	elemType := v4.(*YTFieldType)
	return &YTFieldType{
		YTMapTypee: &YTMapTypee{
			Value: &YTListType{
				YTBaseType:   elemType.YTBaseType,
				YTCustomType: elemType.YTCustomType,
			},
			Key: v2.(*YTBaseType),
		},
	}, nil
}

// NewEnum 新建枚举
// Enum: "enum" tok_identifier "{" EnumItems "}" OptEnd
func NewEnum(v1, v3 interface{}) (enum *YTEnumDef, _ error) {
	if v3 != nil {
		enum = v3.(*YTEnumDef)
	} else {
		enum = &YTEnumDef{}
	}

	tok := v1.(*token.Token)
	enum.Name = tok.IDValue()
	enum.pos = tok.Pos
	return
}

// AppendEnumItem 追加枚举
func AppendEnumItem(val0, val1, val2 interface{}) (enum *YTEnumDef, _ error) {
	debug("AppendEnumItem", val0, val1, val2)
	if val0 != nil {
		enum = val0.(*YTEnumDef)
	} else {
		enum = &YTEnumDef{}
	}
	ev := val1.(*YTEnumValue)
	if val2 != nil {
		ev.YTDoc = val2.(*YTDoc)
	}
	// 填充枚举默认值
	if ev.Value == 0 && len(enum.Values) > 0 {
		ev.Value = enum.Values[len(enum.Values)-1].Value + 1
	}
	if find, tip := enum.checkUnionOption(ev.Name); find {
		return nil, fmt.Errorf("enum item name repeated. %s %s", ev.Name, tip)
	}
	if find, tip := enum.checkUnionNo(ev.Value); find {
		return nil, fmt.Errorf("enum item value repeated. %s value:%d %s", ev.Name, ev.Value, tip)
	}
	enum.Values = append(enum.Values, ev)
	enum.addUnionName(ev.Name, fmt.Sprintf(",last define on $FILE:%d", ev.pos.Line))
	enum.addUnionNo(ev.Value, fmt.Sprintf(",last define on $FILE:%d", ev.pos.Line))
	return
}

// AppendEnumOption 追加枚举
func AppendEnumOption(val0, val1, val2 interface{}) (enum *YTEnumDef, _ error) {
	debug("AppendEnumOption", val0, val1, val2)
	if val0 != nil {
		enum = val0.(*YTEnumDef)
	} else {
		enum = &YTEnumDef{}
	}
	ev := val1.(*YTOption)
	if val2 != nil {
		ev.YTDoc = val2.(*YTDoc)
	}
	if find, tip := enum.checkUnionOption(ev.Key); find {
		return nil, fmt.Errorf("enum option repeated. %s %s", ev.Key, tip)
	}
	enum.Opts = append(enum.Opts, ev)

	enum.addUionOption(ev.Key, fmt.Sprintf(",last define on $FILE:%d", ev.pos.Line))

	return
}

// NewEnumValue 新建枚举项
// EnumItem: tok_identifier EnumItemValue OptEnd
func NewEnumValue(name, val interface{}) (ev *YTEnumValue, _ error) {
	debug("NewEnumValue", name, val)
	ev = &YTEnumValue{}
	ev.Name = name.(*token.Token).IDValue()
	if val != nil {
		v, err := val.(*token.Token).Int64Value()
		if err != nil {
			return nil, err
		}
		ev.Value = v
	}
	ev.pos = name.(*token.Token).Pos
	return
}

// NewOptionVal 新建选项value
//OptionValue:
// 	empty
//	"=" tok_literal		<< ast.NewOptionVal($1,nil) >>
//  "=" tok_const_int 	<< ast.NewOptionVal(nil,$1) >>
func NewOptionVal(lit, num interface{}) (val *YTOptionValue, _ error) {
	debug("NewOptionVal", lit, num)
	val = &YTOptionValue{}
	if lit != nil {
		val.Value = strptr(lit.(*token.Token).StringValue())
	} else if num != nil {
		if n,ok := num.(int); ok {
			v := int64(n)
			val.IntVal = &v
			return
		}
		v, err := num.(*token.Token).Int64Value()
		if err != nil {
			return nil, err
		}
		val.IntVal = &v
	}
	return
}

// NewOption 新建选项
// Option: tok_option OptionValue
func NewOption(v0, v1 interface{}) (opt *YTOption, _ error) {
	debug("NewOption", v0, v1)
	opt = &YTOption{}
	opt.Key = v0.(*token.Token).IDValue()
	if v1 != nil {
		opt.Value = v1.(*YTOptionValue)
	}
	opt.pos = v0.(*token.Token).Pos
	return
}

// NewPackage 新建包
// Package:	"package" tok_identifier OptEnd
func NewPackage(v1 interface{}) (*YTPackage, error) {
	debug("NewPackage", v1)
	tok, ok := v1.(*token.Token)
	if !ok {
		return nil, errors.New("invalid package name")
	}
	return &YTPackage{Name: tok.IDValue()}, nil
}

// AppendDoc 附加文档
// Doc:	Doc tok_doc
func AppendDoc(v0, v1 interface{}) (doc *YTDoc, _ error) {
	debug("AppendDoc", v0, v1)
	if v0 != nil {
		doc = v0.(*YTDoc)
	} else {
		doc = &YTDoc{}
	}
	text := v1.(*token.Token).IDValue()
	text = strings.TrimSpace(text)
	doc.Doc = append(doc.Doc, text)
	return
}

// NewImport 新建导入
// Import: "import" AliasName tok_literal OptEnd
func NewImport(v1, v2 interface{}) (imp *YTImport, _ error) {
	debug("NewImport", v1, v2)
	tokName, ok := v2.(*token.Token)
	if !ok {
		return nil, errors.New("invalid import name")
	}
	debug("len", len(tokName.Lit), tokName.StringValue(), tokName.IDValue())
	imp = &YTImport{File: tokName.StringValue()}
	if v1 != nil {
		tokAlias, ok := v1.(*token.Token)
		if !ok {
			return nil, errors.New("invalid alias name")
		}
		imp.AliasName = tokAlias.IDValue()
		imp.pos = tokAlias.Pos
	} else {
		imp.pos = tokName.Pos
	}

	if gRegisterRecursionAnalyser == nil {
		return
	}

	// 解析依赖文件
	prog, err := gRegisterRecursionAnalyser.Analyse(imp.File)
	if err != nil {
		return nil, err
	}
	// 保存依赖
	imp.Prog = prog
	return
}

////////////////////////////////////////////////////////////
// NOTE: protobuf 支持
// NewSyntax protobuf 支持. 语法
// Syntax: "syntax" "=" tok_identifier ";"
func ProtoNewSyntax(v2 interface{}) (syntax *YTOption, _ error) {
	debug("NewImport", v2)
	nameTok, ok := v2.(*token.Token)
	if !ok {
		return nil, errors.New("invalid syntax")
	}
	name := nameTok.StringValue()
	sv := int64(0)
	switch name {
	case "proto3":
		sv = 3
	case "proto2":
		sv = 2
	default:
		return nil, errors.New("invalid protobuf syntax")
	}
	syntax = &YTOption{
		Key: ProtobufOptionSyntax,
		Value: &YTOptionValue{
			IntVal: &sv,
		},
		pos: nameTok.Pos,
	}
	return
}

// ProtoNewOption 新建选项
// Option: "option" tok_identifier tok_literal ";"
func ProtoNewOption(v1, v2 interface{}) (opt *YTOption, _ error) {
	debug("ProtoNewOption", v1, v2)
	opt = &YTOption{}
	opt.Key = v1.(*token.Token).IDValue()
	val := v2.(*token.Token).StringValue()
	opt.Value = &YTOptionValue{
		Value: &val,
	}
	// 修改成插件支持的选项名称
	switch opt.Key {
	case "go_package":
		opt.Key = "proto.gopkg"
	}
	opt.pos = v1.(*token.Token).Pos
	return
}

// ProtoNewMethod 新建方法
// "rpc" tok_identifier "(" tok_identifier ")" "returns" "(" tok_identifier ")" "{" "}"
func ProtoNewMethod(v1, v3, v7 interface{}) (method *YTMethod, _ error) {
	method = &YTMethod{}
	method.Name = v1.(*token.Token).IDValue()
	method.pos = v1.(*token.Token).Pos

	method.Request = &YTMessage{
		Name:     v3.(*token.Token).IDValue(),
		protobuf: true,
	}
	method.Reply = &YTMessage{
		Name:     v7.(*token.Token).IDValue(),
		protobuf: true,
	}
	return
}

// protobuf 消息解析设置选项
const (
	// ProtobufOptionSyntax syntax = "proto2" 选项
	ProtobufOptionSyntax = "proto.syntax"
)
