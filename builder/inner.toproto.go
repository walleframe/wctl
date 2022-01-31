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
	"log"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/aggronmagi/wctl/builder/gen"
	"github.com/aggronmagi/wctl/protocol/ast"
)

func fileName(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name)) + ".proto"
}

type toProtoGenerater struct {
	// tpl *template.Template
}

// Generate 生成代码接口
func (gen *toProtoGenerater) Generate(prog *ast.YTProgram) (outs []*Output, err error) {
	data, err := transProtobuf(prog)
	outs = append(outs, &Output{
		File: fileName(prog.File),
		Data: data,
	})
	return
}

// Union 唯一标识符 用于标识不同插件
func (gen *toProtoGenerater) Union() string {
	return "toproto"
}

func init() {
	//tpl := template.New("")
	//tpl.Funcs(gTplFunc)
	//tpl, err := tpl.Parse(gToProtoTemplate)
	//utils.PanicIf(err, "内置生成器: 转换protobuf")
	RegisterGenerater(&toProtoGenerater{
		//	tpl: tpl,
	})
}

var gTplFunc = template.FuncMap{
	"normal": gTplNormalize,
}

func gTplNormalize(name string) (final string, err error) {
	buf := &strings.Builder{}
	buf.Grow(len(name))
	index := strings.IndexByte(name, '.')
	if index >= 0 {
		buf.WriteString(name[:index+1])
		name = name[index+1:]
	}

	change := false
	for k, v := range name {
		if k == 0 {
			v = unicode.ToTitle(v)
			buf.WriteRune(v)
			continue
		}
		if v == '_' {
			change = true
			continue
		}
		if change {
			v = unicode.ToTitle(v)
		}
		buf.WriteRune(v)
	}
	return buf.String(), nil
}

func printDoc(pb *gen.Generator, docs *ast.YTDoc) {
	if docs == nil {
		return
	}
	for _, v := range docs.Doc {
		lists := strings.Split(v, "\n")
		for _, lv := range lists {
			if len(lv) < 1 {
				continue
			}
			pb.P(lv)
		}
	}
}

func baseTypeName(typ *ast.YTBaseType) string {
	switch typ {
	case ast.BaseTypeInt8, ast.BaseTypeInt16, ast.BaseTypeInt32:
		return "int32"
	case ast.BaseTypeUint8, ast.BaseTypeUint16, ast.BaseTypeUint32:
		return "uint32"
	case ast.BaseTypeInt64:
		return "int64"
	case ast.BaseTypeUint64:
		return "uint64"
	case ast.BaseTypeString:
		return "string"
	case ast.BaseTypeBinary:
		return "string"
	case ast.BaseTypeBool:
		return "bool"
	case ast.BaseTypeFloat32:
		return "float"
	case ast.BaseTypeFloat64:
		return "double"
	default:
		fmt.Println(typ)
		panic("unkown type")
	}
}

func getTypeName(typ *ast.YTFieldType) string {
	switch {
	case typ.YTBaseType != nil:
		return baseTypeName(typ.YTBaseType)
	case typ.YTCustomType != nil:
		return typ.YTCustomType.Name
	case typ.YTListType != nil:
		if typ.YTListType.YTBaseType != nil {
			return "repeated " + baseTypeName(typ.YTListType.YTBaseType)
		} else if typ.YTListType.YTCustomType != nil {
			return "repeated " + typ.YTListType.YTCustomType.Name
		}
	case typ.YTMapTypee != nil:

		if typ.YTMapTypee.Value.YTBaseType != nil {
			return "map<" + baseTypeName(typ.YTMapTypee.Key) + "," +
				baseTypeName(typ.YTMapTypee.Value.YTBaseType) + ">"
		} else if typ.YTMapTypee.Value.YTCustomType != nil {
			return "map<" + baseTypeName(typ.YTMapTypee.Key) + "," +
				typ.YTMapTypee.Value.YTCustomType.Name + ">"
		}
	}
	fmt.Printf("%#v\n", typ)
	panic("invalid type")
}

const (
	OptionGoPKG = "proto.gopkg"
)

func transProtobuf(prog *ast.YTProgram) (data []byte, err error) {
	if !prog.HasOption(OptionGoPKG) {
		//err = fmt.Errorf("文件[%s]不包含 proto.gopkg 选项.无法转换protobuf协议", prog.File)
		//return
		log.Println("WARN: 文件 [", prog.File, "] 不包含 ", OptionGoPKG)
	}
	pb := gen.New(gen.WithIndent("  "))
	pb.P("// Generate by ytctl. DO NOT EDIT.")
	pb.P(`syntax = "proto3";`)
	pb.P(`// source file: `, prog.File)
	printDoc(pb, prog.Pkg.YTDoc)
	pb.P(`package `, prog.Pkg.Name, `;`)
	if prog.HasOption(OptionGoPKG) {
		pb.P(`option go_package = "`, prog.GetOptionString(OptionGoPKG), `";`)
	}
	pb.P()
	// Import
	for _, v := range prog.Imports {
		printDoc(pb, v.YTDoc)
		pb.P(`import "`, fileName(v.File), `";`)
	}
	pb.P()
	// 枚举定义
	for _, v := range prog.EnumDefs {
		printDoc(pb, v.YTDoc)
		pb.P(`enum `, v.Name, " {")
		pb.In()

		for _, ev := range v.Values {
			printDoc(pb, ev.YTDoc)
			pb.P(ev.Name, " = ", ev.Value, ";")
		}

		pb.Out()
		pb.P(`}`)
	}
	// Message
	for _, v := range prog.Messages {
		printDoc(pb, v.YTDoc)
		pb.P(`message `, v.Name, " {")
		pb.In()

		for _, fv := range v.Fields {
			printDoc(pb, fv.YTDoc)
			pb.P(getTypeName(fv.Type), " ", fv.Name, " = ", fv.No, ";")
		}

		pb.Out()
		pb.P(`}`)
	}
	data, _ = pb.Bytes()
	return
}

// var gToProtoTemplate = `
// {{define "ProtoMessage"}}
// syntax = "proto3";
// package {{.Pkg.Name}};
// option go_package = "{{.GetOptionString "proto.gopkg" }}";

// {{range .Messages -}}
// {{ template "doc" . }}message {{normal .Name}}{
// {{ range .Fields }}
// {{template "doc" .}}
// 	{{template "type" .Type}} {{normal .Name}} = {{.No}};
// {{ end }}
// }

// {{ end -}}
// {{end}}

// {{/* 类型名称 */}}
// {{define "type" -}}
// {{- if .YTBaseType -}}
//     {{.YTBaseType}}
// {{- else if .YTListType -}}
//     {{- if .YTListType.YTBaseType -}}
//         {{.YTListType.YTBaseType}}
//     {{- else -}}
//         {{normal .YTListType.YTCustomType}}
//     {{- end -}}
// {{- else if .YTMapTypee -}}
//     map[.YTBaseType]{{- if .YTMapTypee.Value.YTBaseType -}}
//         {{.YTMapTypee.Value.YTBaseType}}
//     {{- else -}}
//         {{normal .YTMapTypee.Value.YTCustomType.Name}}
//     {{- end -}}
// {{- else if .YTCustomType -}}
//     {{normal .YTCustomType.Name}}
// {{- else -}}
//     unkown
// {{- end -}}
// {{- end}}

// {{/* 文档注释 */}}
// {{define "doc" -}}
// // {{normal .Name}} generate by ytctl.DO NOT EDIT
// {{with .YTDoc -}}
// {{range .Doc}}{{.}}
// {{- end}}{{- end}}{{- end}}
// `
