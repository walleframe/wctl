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
	"strconv"
	"strings"

	"github.com/aggronmagi/wctl/builder/buildpb"
	"github.com/aggronmagi/wctl/protocol/token"
	"github.com/aggronmagi/wctl/utils"
)

// YTProgram 文件语法树
type YTProgram struct {
	ytCheck                           // 命名检查
	YTOptions                         // 包 选项
	impMap    map[string][]*YTProgram // 依赖映射
	msgMap    map[string]*YTMessage   // 消息映射
	Pkg       *YTPackage              // 包定义
	Imports   []*YTImport             // 导入文件
	EnumDefs  []*YTEnumDef            // 枚举定义
	Messages  []*YTMessage            // 消息定义
	Services  []*YTService            // 服务定义
	Projects  []*YTProject            // 项目定义
	// File 文件名 - 只有整个文件解析成功才会赋值
	File string
	// 解析阶段不使用. 仅用于生成阶段. 放在这做缓存
	desc *buildpb.FileDesc
}

// ApplyCmdOptions 应用命令行参数添加的全局选项
// 全局Options. 格式为 "xx.xxx=66" "xx.x1" "xx.xx2=xxx"
func (prog *YTProgram) ApplyCmdOptions(opts ...string) {
	if len(opts) < 1 {
		return
	}
	var index, check int
	var key, value string
	for _, opt := range opts {
		opt = strings.TrimSpace(opt)
		index = strings.IndexByte(opt, '=')
		// 无value
		if index < 0 {
			// 有效性检测
			check = strings.IndexByte(opt, '.')
			if check < 0 {
				fmt.Println("Error:", opt, "is not valid options. Ignore!!!")
				continue
			}
			// 保存
			prog.Opts = append(prog.Opts, &YTOption{
				Key: opt,
			})
			continue
		}
		key = opt[:index]
		key = strings.TrimSpace(key)
		// 有效性检测
		check = strings.IndexByte(key, '.')
		if check < 0 {
			fmt.Println("Error:", opt, "is not valid options. Ignore!!!")
			continue
		}
		value = opt[index+1:]
		value = strings.TrimSpace(value)
		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			// 保存
			prog.Opts = append(prog.Opts, &YTOption{
				Key: opt,
				Value: &YTOptionValue{
					Value: &value,
				},
			})
			continue
		}
		prog.Opts = append(prog.Opts, &YTOption{
			Key: opt,
			Value: &YTOptionValue{
				IntVal: &num,
			},
		})
	}
}

// YTDoc 文档,注释
type YTDoc struct {
	Doc []string
}

// YTPackage 文件包定义.
type YTPackage struct {
	*YTDoc // 文件注释
	Name   string
}

// YTImport 导入依赖
type YTImport struct {
	*YTDoc
	pos       token.Pos
	Prog      *YTProgram
	File      string
	AliasName string
}

// YTOption 定义选项节点
type YTOption struct {
	*YTDoc
	pos   token.Pos
	Key   string
	Value *YTOptionValue
}

// YTOptionValue 选项值
type YTOptionValue struct {
	Value  *string
	IntVal *int64
}

func (val *YTOptionValue) String() (desc string) {
	if val == nil {
		return
	}
	if val.Value != nil {
		return *val.Value
	}
	return strconv.FormatInt(*val.IntVal, 10)
}

// YTEnumDef 枚举定义
type YTEnumDef struct {
	ytCheck
	*YTDoc
	YTOptions
	Name   string
	Values []*YTEnumValue
}

// YTEnumValue 枚举值
type YTEnumValue struct {
	*YTDoc
	pos   token.Pos
	Name  string
	Value int64
}

// YTMessage 消息定义
type YTMessage struct {
	ytCheck
	*YTDoc
	YTOptions
	Name     string
	Fields   []*YTField
	protobuf bool
}

// YTField 字段定义
type YTField struct {
	*YTDoc
	YTOptions
	pos  token.Pos
	Type *YTFieldType
	No   uint8
	Name string
}

// YTFieldType 字段类型
type YTFieldType struct {
	*YTBaseType
	*YTListType
	*YTMapTypee
	*YTCustomType
}

// YTCustomType 自定义类型
type YTCustomType struct {
	Name string
	Msg  *YTMessage
}

// YTListType 列表类型
type YTListType struct {
	*YTBaseType
	*YTCustomType
}

// YTMapTypee 映射类型
type YTMapTypee struct {
	Key   *YTBaseType
	Value *YTListType
}

// YTBaseType 基本类型
type YTBaseType int

var (
	// BaseTypeInt8 Int8
	BaseTypeInt8 *YTBaseType
	// BaseTypeUint8 Uint8
	BaseTypeUint8 *YTBaseType
	// BaseTypeInt16 Int16
	BaseTypeInt16 *YTBaseType
	// BaseTypeUint16 Uint16
	BaseTypeUint16 *YTBaseType
	// BaseTypeInt32 Int32
	BaseTypeInt32 *YTBaseType
	// BaseTypeUint32 Uint32
	BaseTypeUint32 *YTBaseType
	// BaseTypeInt64 Int64
	BaseTypeInt64 *YTBaseType
	// BaseTypeUint64 Uint64
	BaseTypeUint64 *YTBaseType
	// BaseTypeString String
	BaseTypeString *YTBaseType
	// BaseTypeBinary Binary
	BaseTypeBinary *YTBaseType
	// BaseTypeBool Bool
	BaseTypeBool *YTBaseType
	// BaseTypeFloat32 float32 float
	BaseTypeFloat32 *YTBaseType
	// BaseTypeFloat64 float64 double
	BaseTypeFloat64 *YTBaseType
)

func init() {
	set := func(v int) *YTBaseType {
		vt := YTBaseType(v)
		return &vt
	}
	BaseTypeInt8 = set(0)
	BaseTypeUint8 = set(1)
	BaseTypeInt16 = set(2)
	BaseTypeUint16 = set(3)
	BaseTypeInt32 = set(4)
	BaseTypeUint32 = set(5)
	BaseTypeInt64 = set(6)
	BaseTypeUint64 = set(7)
	BaseTypeString = set(8)
	BaseTypeBinary = set(9)
	BaseTypeBool = set(10)
	BaseTypeFloat32 = set(11)
	BaseTypeFloat64 = set(12)
}

func (typ *YTBaseType) String() string {
	switch typ {
	case BaseTypeInt8:
		return "int8"
	case BaseTypeUint8:
		return "uint8"
	case BaseTypeInt16:
		return "int16"
	case BaseTypeUint16:
		return "uint16"
	case BaseTypeInt32:
		return "int32"
	case BaseTypeUint32:
		return "uint32"
	case BaseTypeInt64:
		return "int64"
	case BaseTypeUint64:
		return "uint64"
	case BaseTypeString:
		return "string"
	case BaseTypeBinary:
		return "binary"
	case BaseTypeBool:
		return "bool"
	case BaseTypeFloat32:
		return "float32"
	case BaseTypeFloat64:
		return "float64"
	default:
		return "unkown"
	}
}

// YTService 服务定义
type YTService struct {
	ytCheck
	*YTDoc
	YTOptions
	flag    MethodFlag
	Name    string
	Methods []*YTMethod
}

// MethodFlag 方法标记
type MethodFlag int8

// 方法标记
const (
	Call MethodFlag = iota
	Notify
)

//  String 文字描述
func (flag MethodFlag) String() string {
	switch flag {
	case Call:
		return "Call"
	case Notify:
		return "Notify"
	default:
		return "unsupport"
	}
}

// YTMethod 函数,方法定义
type YTMethod struct {
	*YTDoc
	YTOptions
	pos     token.Pos
	Flag    MethodFlag
	Name    string
	Request *YTMessage
	Reply   *YTMessage
	No      *YTMethodNo
}

// YTMethodNo 方法ID
type YTMethodNo struct {
	pos   token.Pos
	Macro *YTCustomType
	Value *int64
}

// YTProject 项目定义
type YTProject struct {
	*YTDoc
	area  string
	Name  string
	Conf  map[string]*YTOptions
	check map[string]*ytCheck
}

// YTOptions 选项
type YTOptions struct {
	Opts []*YTOption
	opm  map[string]*YTOption
}

func (opts *YTOptions) check() {
	if len(opts.Opts) == len(opts.opm) {
		return
	}
	opts.opm = make(map[string]*YTOption, len(opts.Opts))
	for _, v := range opts.Opts {
		opts.opm[v.Key] = v
	}
}

// HasOption 是否有选项
func (opts *YTOptions) HasOption(key string) bool {
	opts.check()
	if _, ok := opts.opm[key]; ok {
		return true
	}
	return false
}

// GetOption 获取选项
func (opts *YTOptions) GetOption(key string) *YTOption {
	opts.check()
	if opt, ok := opts.opm[key]; ok {
		return opt
	}
	return nil
}

// GetOptionValue 获取选项值
func (opts *YTOptions) GetOptionValue(key string) *YTOptionValue {
	opts.check()
	if opt, ok := opts.opm[key]; ok {
		return opt.Value
	}
	return nil
}

// GetOptionString 获取选项值
func (opts *YTOptions) GetOptionString(key string) (val string) {
	if utils.ShowDetail() {
		fmt.Println("get option", key)
	}
	opts.check()
	if opt, ok := opts.opm[key]; ok {
		if opt.Value != nil && opt.Value.Value != nil {
			return *opt.Value.Value
		}
	}
	return
}
