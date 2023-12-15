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

	"github.com/walleframe/wctl/protocol/errors"
	"github.com/walleframe/wctl/protocol/token"
	"github.com/walleframe/wctl/utils"
)

func NewError(tok *token.Token, format string, args ...interface{}) error {
	return &errors.Error{
		Err:        fmt.Errorf(format, args...),
		ErrorToken: tok,
	}
}

func NewError2(tok *token.Token, err error) error {
	return &errors.Error{
		Err:        err,
		ErrorToken: tok,
	}
}

func NewErrorPos(pos token.Pos, format string, args ...interface{}) error {
	return &errors.Error{
		Err: fmt.Errorf(format, args...),
		ErrorToken: &token.Token{
			Pos: pos,
		},
	}
}

func NewErrorPos2(pos token.Pos, err error) error {
	return &errors.Error{
		Err: err,
		ErrorToken: &token.Token{
			Pos: pos,
		},
	}
}

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

// //////////////////////////////////////////////////////////
// NOTE: protobuf 支持
// NewSyntax protobuf 支持. 语法
// Syntax: "syntax" "=" tok_identifier ";"
func ProtoNewSyntax(v2 interface{}) (syntax *YTOption, _ error) {
	debug("NewImport", v2)
	nameTok := v2.(*token.Token)

	name := nameTok.StringValue()
	sv := int64(0)
	switch name {
	case "proto3":
		sv = 3
	case "proto2":
		sv = 2

	}
	syntax = &YTOption{
		Key: ProtobufOptionSyntax,
		Value: &YTOptionValue{
			IntVal: &sv,
		},
		DefPos: nameTok.Pos,
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
	opt.DefPos = v1.(*token.Token).Pos
	return
}

// ProtoNewMethod 新建方法
// "rpc" tok_identifier "(" tok_identifier ")" "returns" "(" tok_identifier ")" "{" "}"
func ProtoNewMethod(v1, v3, v7 interface{}) (method *YTMethod, _ error) {
	method = &YTMethod{}
	method.Name = v1.(*token.Token).IDValue()
	method.DefPos = v1.(*token.Token).Pos

	method.Request = &YTMessage{
		Name:         v3.(*token.Token).IDValue(),
		ProtobufFlag: true,
	}
	method.Reply = &YTMessage{
		Name:         v7.(*token.Token).IDValue(),
		ProtobufFlag: true,
	}
	return
}

// protobuf 消息解析设置选项
const (
	// ProtobufOptionSyntax syntax = "proto2" 选项
	ProtobufOptionSyntax = "proto.syntax"
)
