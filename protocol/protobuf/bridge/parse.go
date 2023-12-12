package bridge

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/aggronmagi/wctl/protocol/ast"
	"github.com/aggronmagi/wctl/protocol/token"
	"github.com/aggronmagi/wctl/utils"
	"go.uber.org/multierr"
)

// protobuf 消息解析设置选项
const (
	// ProtobufOptionSyntax syntax = "proto2" 选项
	ProtobufOptionSyntax = "proto.syntax"
	ProtobufOptionPacked = "proto.packed"
	ProtobufOptionFixed  = "proto.fixed"
	ProtobufOptionSigned = "proto.signed"
)

func NewProtocol(c, pkg, imports, defines interface{}) (_ *ast.YTProgram, err error) {
	ctx := c.(*ast.Context)
	_ = ctx.Prog
	utils.Debugln("new protocol")
	return ctx.Prog, nil
}

// Syntax: "syntax" "=" tok_literal ";"			<< bridge.ProtoSyntax($Context, $2) >>
func ProtoSyntax(c, a0 interface{}) (opt *ast.YTOption, err error) {
	ctx := c.(*ast.Context)
	tokVersion := a0.(*token.Token)
	val := int64(0)
	switch tokVersion.StringValue() {
	case "proto3":
		val = 3
	case "proto2":
		val = 2
	default:
		return nil, ast.NewError(tokVersion, "protobuf syntax version invalid [%s] ", tokVersion.IDValue())
	}
	opt = &ast.YTOption{
		YTDoc:  ctx.PreDoc(tokVersion.Line),
		DefPos: tokVersion.Pos,
		Key:    ProtobufOptionSyntax,
		Value: &ast.YTOptionValue{
			IntVal: &val,
		},
	}
	ctx.Prog.Opts = append(ctx.Prog.Opts, opt)
	return
}

// Option: "option" tok_identifier "=" tok_literal ";"	<< ast.ProtoNewOption($Context, $1, $3) >>
func ProtoNewOption(c, a1, a3 interface{}) (opt *ast.YTOption, err error) {
	ctx := c.(*ast.Context)
	tokName := a1.(*token.Token)
	optName := tokName.IDValue()
	// 转换成wpb生成所使用的option name.
	if optName == "go_package" {
		optName = "proto.gopkg"
	}
	opt = &ast.YTOption{
		YTDoc:  ctx.PreDoc(tokName.Line),
		DefPos: tokName.Pos,
		Key:    optName,
		Value:  &ast.YTOptionValue{},
	}
	val := a3.(*token.Token)
	if strings.HasPrefix(val.IDValue(), `"`) {
		v := val.StringValue()
		opt.Value.Value = &v
	} else {
		return nil, ast.NewError(val, "protobuf option value not string literal")
	}
	ctx.Prog.Opts = append(ctx.Prog.Opts, opt)

	return
}

// Package: "package" tok_identifier OptEnd
func NewPackage(c, v1 interface{}) (pkg *ast.YTPackage, err error) {
	ctx := c.(*ast.Context)
	tok := v1.(*token.Token)
	pkg = &ast.YTPackage{
		DefPos: tok.Pos,
		Name:   tok.IDValue(),
		YTDoc:  ctx.PreDoc(tok.Line),
	}
	// 检测package名有效性
	err = checkNormalIdentifier(pkg.Name, "package name define")
	if err != nil {
		return nil, ast.NewError2(tok, err)
	}

	ctx.LastElement = pkg
	ctx.Prog.Pkg = pkg
	utils.Debugln("new package")
	return
}

// Import: "import" tok_identifier tok_literal OptEnd	<< bridge.NewImport($Context, $2, $1) >>
func NewImport(c, pkg, alias interface{}) (imp *ast.YTImport, err error) {
	ctx := c.(*ast.Context)
	tokPkg := pkg.(*token.Token)

	imp = &ast.YTImport{
		DefPos: tokPkg.Pos,
		YTDoc:  ctx.PreDoc(tokPkg.Line),
		File:   tokPkg.StringValue(),
	}
	// 检测import文件路径格式
	err = checkFilePath(imp.File, "import file path")
	if err != nil {
		return nil, ast.NewError2(tokPkg, err)
	}
	// import 别名
	if tokAlias, ok := alias.(*token.Token); ok {
		imp.AliasName = tokAlias.IDValue()
		err = checkNormalIdentifier(imp.AliasName, "import alias name")
		if err != nil {
			return nil, ast.NewError2(tokPkg, err)
		}
	}

	// 解析依赖文件
	if ast.RegisterRecursionAnalyser != nil {
		prog, err := ast.RegisterRecursionAnalyser.Analyse(imp.File)
		if err != nil {
			return nil, ast.NewError(tokPkg, "import file failed %+v", err)
		}
		// 保存依赖
		imp.Prog = prog
		//
		utils.Debugln("RecursionAnalyse true", prog.Pkg.Name)
	} else {
		utils.Debugln("RecursionAnalyse false")
	}

	ctx.LastElement = imp
	ctx.Prog.Imports = append(ctx.Prog.Imports, imp)
	utils.Debugln("new imports")
	return
}

// Enum: "enum" tok_identifier "{" Options "}" OptEnd	<< bridge.NewEnum($Context, $1, $3) >>
func NewEnum(c, en, evs interface{}) (def *ast.YTEnumDef, err error) {
	ctx := c.(*ast.Context)
	tokName := en.(*token.Token)
	opts := evs.(*ast.YTOptions)
	def = &ast.YTEnumDef{
		YTDoc:  ctx.PreDoc(tokName.Line),
		Name:   tokName.IDValue(),
		DefPos: tokName.Pos,
	}
	err = checkNormalIdentifier(def.Name, "enum name define")
	if err != nil {
		return nil, ast.NewError2(tokName, err)
	}

	// 将所有option拆分成enum定义和option选项定义
	enumValue := int64(-1)
	valCheck := make(map[int64][]string)
	for _, v := range opts.Opts {
		// 枚举值
		if checkNormalIdentifier(v.Key, "") == nil {
			if v.Value != nil {
				if v.Value.IntVal != nil {
					// 使用设置的枚举值
					enumValue = *v.Value.IntVal
				} else if v.Value.Value != nil {
					// 枚举值设置了string类型
					return nil, ast.NewErrorPos(v.DefPos, "enum value [%s.%s=%s] invalid", def.Name, v.Key, *v.Value.Value)
				} else {
					// 未设置任何值, 自增
					enumValue++
				}
			} else {
				// 未设置任何值, 自增
				enumValue++
			}
			def.Values = append(def.Values, &ast.YTEnumValue{
				YTDoc:  v.YTDoc,
				DefPos: v.DefPos,
				Name:   v.Key,
				Value:  enumValue,
			})
			valCheck[enumValue] = append(valCheck[enumValue], v.Key)
			continue
		} else if checkOptionName(v.Key, "") == nil {
			// 选项设置
			def.Opts = append(def.Opts, v)
		} else {
			// key 无效
			return nil, ast.NewErrorPos(v.DefPos, "enum define [%s.%s] name invalid, not option and not enum value name.", def.Name, v.Key)
		}
	}
	// 枚举值重复检测
	for k, v := range valCheck {
		if len(v) > 1 {
			err = multierr.Append(err, fmt.Errorf("enum value [%d] repeated. %+v", k, v))
		}
	}
	if err != nil {
		return nil, ast.NewError2(tokName, err)
	}

	ctx.LastElement = def
	ctx.Prog.EnumDefs = append(ctx.Prog.EnumDefs, def)

	utils.Debugln("    ---   ---  ", def.YTDoc)

	return
}

// Message: "message" tok_identifier "{" Fields "}" OptEnd  << bridge.NewMessage($Context, $1, $3) >>
func NewMessage(c, mn, fvs interface{}) (def *ast.YTMessage, err error) {
	ctx := c.(*ast.Context)
	tokName := mn.(*token.Token)
	def = fvs.(*ast.YTMessage)
	def.YTDoc = ctx.PreDoc(tokName.Line)
	def.Name = tokName.IDValue()
	def.DefPos = tokName.Pos
	err = checkNormalIdentifier(def.Name, "enum name define")
	if err != nil {
		return nil, ast.NewError2(tokName, err)
	}

	ctx.LastElement = def
	ctx.Prog.Messages = append(ctx.Prog.Messages, def)

	return
}

// Fields FieldExpr	<< bridge.FieldField($Context, $0, $1) >>
func FieldField(c, m, v interface{}) (msg *ast.YTMessage, err error) {
	ctx := c.(*ast.Context)
	msg = m.(*ast.YTMessage)
	field := v.(*ast.YTField)
	msg.Fields = append(msg.Fields, field)
	ctx.LastElement = field
	// utils.Debugln("filed field", field.DefPos.String())
	return
}

// Fields "repeated" FieldExpr	<< bridge.FieldArray($Context, $0, $1) >>
func FieldArray(c, m, v interface{}) (msg *ast.YTMessage, err error) {
	ctx := c.(*ast.Context)
	msg = m.(*ast.YTMessage)
	field := v.(*ast.YTField)

	msg.Fields = append(msg.Fields, field)
	// TODO: fix array type
	ctx.LastElement = field
	//utils.Debugln("filed array", field.DefPos.String())
	return
}

// Fields Message	<< bridge.FieldMessage($Context, $0, $1) >>
func FieldMessage(c, m, v interface{}) (msg *ast.YTMessage, err error) {
	ctx := c.(*ast.Context)
	msg = m.(*ast.YTMessage)
	sub := v.(*ast.YTMessage)
	msg.SubMsgs = append(msg.SubMsgs, sub)
	// 修复 NewMessage 导致的重复添加
	for k, v := range ctx.Prog.Messages {
		if v.Name == sub.Name && v.DefPos == sub.DefPos {
			ctx.Prog.Messages = append(ctx.Prog.Messages[:k], ctx.Prog.Messages[k+1:]...)
			break
		}
	}
	ctx.LastElement = sub
	return
}

// OptionExpr: tok_identifier OptionValue OptEnd << bridge.OptionExpr($Context, $0, $1) >>
func OptionExpr(c, n, v interface{}) (opt *ast.YTOption, err error) {
	ctx := c.(*ast.Context)
	tokName := n.(*token.Token)
	optName := tokName.IDValue()
	// e1 := checkNormalIdentifier(optName, "enum value name")
	// e2 := checkOptionName(optName, "option name")
	opt = &ast.YTOption{
		YTDoc:  ctx.PreDoc(tokName.Line),
		DefPos: tokName.Pos,
		Key:    optName,
		Value:  &ast.YTOptionValue{},
	}
	switch val := v.(type) {
	case bool:
		v := int64(0)
		if val {
			v = 1
		}
		opt.Value.IntVal = &v
	case int:
		v := int64(val)
		opt.Value.IntVal = &v
	case *token.Token:
		if strings.HasPrefix(val.IDValue(), `"`) {
			v := val.StringValue()
			opt.Value.Value = &v
		} else {
			v := strings.TrimPrefix(val.IDValue(), "+")
			// number
			if strings.HasPrefix(v, "0x") {
				num, err := strconv.ParseInt(strings.TrimPrefix(v, "0x"), 16, 64)
				if err != nil {
					return nil, ast.NewError(val, "hex value [%s] invalid.%+v", v, err)
				}
				opt.Value.IntVal = &num
			} else {
				num, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return nil, ast.NewError(val, "number value [%s] invalid.%+v", v, err)
				}
				opt.Value.IntVal = &num
			}
		}
	default:
		return nil, ast.NewError(tokName, "value invalid. %#v", val)
	}
	ctx.LastElement = opt
	return
}

// FieldExpr: FieldType tok_identifier "=" tok_num FieldOption OptEnd << bridge.NewField($Context,$0, $1, $3, $4) >>
func NewField(c, a0, a1, a3, a4 interface{}) (field *ast.YTField, err error) {
	ctx := c.(*ast.Context)
	field = a0.(*ast.YTField)
	tokName := a1.(*token.Token)
	tokNum := a3.(*token.Token)

	num, err := tokNum.Int32Value()
	if err != nil {
		return nil, ast.NewError2(tokNum, err)
	}

	if num > math.MaxUint8 || num < 0 {
		return nil, ast.NewError(tokNum, "field no must between 1 and 255")
	}

	field.YTDoc = ctx.PreDoc(tokName.Line)
	field.DefPos = tokName.Pos
	field.Name = tokName.IDValue()
	field.No = uint8(num)

	if a4 != nil {
		field.YTOptions.Opts = append(field.YTOptions.Opts, a4.(*ast.YTOptions).Opts...)
	}

	// field.Type = fieldType

	ctx.LastElement = field
	return
}

// Options OptionExpr						<< bridge.AppendOption($Context, $0, $1) >>
func AppendOption(c, a0, a1 interface{}) (val *ast.YTOptions, err error) {
	// ctx := c.(*ast.Context)
	if a0 == nil {
		val = &ast.YTOptions{}
	} else {
		val = a0.(*ast.YTOptions)
	}
	val.Opts = append(val.Opts, a1.(*ast.YTOption))
	return
}

// FieldType: "repeated" tok_identifier << bridge.ArrayType($Context, $1) >>
func ArrayType(c, a0 interface{}) (_ *ast.YTField, err error) {
	// ctx := c.(*ast.Context)
	tokType := a0.(*token.Token)
	ftyp, err := analyseType(tokType.IDValue(), "field array type")
	if err != nil {
		return nil, err
	}

	custom, basic := ftyp.Type.YTCustomType, ftyp.Type.YTBaseType

	ftyp.Type = &ast.YTFieldType{
		YTListType: &ast.YTListType{
			YTCustomType: custom,
			YTBaseType:   basic,
		},
	}

	return ftyp, nil
}

// FieldType: "map" "<" tok_identifier "," tok_identifier ">" << bridge.MapType($Context, $2, $4) >>
func MapType(c, a2, a4 interface{}) (_ *ast.YTField, err error) {
	// ctx := c.(*ast.Context)
	tokKey := a2.(*token.Token)
	tokValue := a4.(*token.Token)
	key, err := analyseType(tokKey.IDValue(), "field map key type")
	if err != nil {
		return nil, err
	}
	value, err := analyseType(tokValue.IDValue(), "field map value type")
	if err != nil {
		return nil, err
	}

	if key.Type.YTBaseType != nil {
		err = ast.NewError(tokKey, "field map key type must basic type")
		return
	}
	if value.Type.YTBaseType == nil && value.Type.YTCustomType == nil {
		err = ast.NewError(tokValue, "field map value type must basic or custom type")
		return
	}
	if len(key.Opts) > 0 {
		err = ast.NewError(tokKey, "field map key type not support signed/fixed type")
		return
	}

	if len(value.Opts) > 0 {
		err = ast.NewError(tokValue, "field map value type not support signed/fixed type")
		return
	}

	return &ast.YTField{
		Type: &ast.YTFieldType{
			YTMapTypee: &ast.YTMapTypee{
				Key: key.Type.YTBaseType,
				Value: &ast.YTListType{
					YTBaseType:   value.Type.YTBaseType,
					YTCustomType: value.Type.YTCustomType,
				},
			},
		},
	}, nil
}

// FieldType: tok_identifier 	<< bridge.BasicOrCustomType($Context, $0) >>
func BasicOrCustomType(c, a0 interface{}) (*ast.YTField, error) {
	// ctx := c.(*ast.Context)
	tokType := a0.(*token.Token)
	return analyseType(tokType.IDValue(), "field basic or custom type")
}

// Service: "service" tok_identifier  "{" ServiceElements "}" OptEnd << bridge.NewService($Context, $1, $3) >>
func NewService(c, a1, a3 interface{}) (_ *ast.YTService, err error) {
	ctx := c.(*ast.Context)
	tokName := a1.(*token.Token)
	svc := a3.(*ast.YTService)

	err = checkNormalIdentifier(tokName.IDValue(), "service name")
	if err != nil {
		return nil, ast.NewError2(tokName, err)
	}

	svc.DefPos = tokName.Pos
	svc.Name = tokName.IDValue()

	ctx.LastElement = svc
	return
}

// ServiceElements: ServiceElements ServiceMethod << bridge.ServiceMethod($Context, $0, $1) >>
func ServiceMethod(c, a0, a1 interface{}) (_ *ast.YTService, err error) {
	ctx := c.(*ast.Context)
	svc := a0.(*ast.YTService)
	method := a1.(*ast.YTMethod)
	method.Flag = svc.Flag
	svc.Methods = append(svc.Methods, method)
	// method.YTDoc = ctx.PreDoc()
	ctx.LastElement = method
	return svc, nil
}

// ServiceElements: ServiceElements OptionExpr    << bridge.ServiceOption($Context, $0, $1) >>
func ServiceOption(c, a0, a1 interface{}) (_ *ast.YTService, err error) {
	ctx := c.(*ast.Context)
	svc := a0.(*ast.YTService)
	opt := a1.(*ast.YTOption)
	svc.Opts = append(svc.Opts, opt)
	ctx.LastElement = opt
	return svc, nil
}

// ServiceElements: ServiceElements MethodFlag	 << bridge.ServiceFlag($Context, $0, $1) >>
func ServiceFlag(c, a0, a1 interface{}) (_ *ast.YTService, err error) {
	ctx := c.(*ast.Context)
	svc := a0.(*ast.YTService)
	flag := a1.(*token.Token)
	switch flag.IDValue() {
	case "call":
		svc.Flag = ast.Call
	case "notify":
		svc.Flag = ast.Notify
	default:
		return nil, ast.NewError(flag, "service flag invalid [%s]", flag.IDValue())
	}
	// 去除无用前置文档
	ctx.PreDoc(flag.Line)
	ctx.LastElement = flag
	return svc, nil
}

// ServiceMethod: tok_identifier "(" tok_identifier ")" tok_identifier MethodNo FieldOption OptEnd << bridge.NewMethod($Context, $0, $2, $4, $5, $6) >>
func NewMethod(c, a0, a2, a4, a5, a6 interface{}) (m *ast.YTMethod, err error) {
	ctx := c.(*ast.Context)
	tokFunc := a0.(*token.Token)
	tokRQ := a2.(*token.Token)
	tokRS := a4.(*token.Token)

	err = checkNormalIdentifier(tokFunc.IDValue(), "method name")
	if err != nil {
		err = ast.NewError2(tokFunc, err)
		return
	}

	rqField, err := analyseType(tokRQ.IDValue(), "method request body")
	if err != nil {
		err = ast.NewError2(tokRQ, err)
		return
	}
	if rqField.Type.YTCustomType == nil {
		err = ast.NewError(tokRQ, "method request body must be custom message type [%s]", tokRQ.IDValue())
		return
	}

	rsField, err := analyseType(tokRS.IDValue(), "method reply body")
	if err != nil {
		err = ast.NewError2(tokRS, err)
		return
	}
	if rsField.Type.YTCustomType == nil {
		err = ast.NewError(tokRS, "method reply body must be custom message type [%s]", tokRS.IDValue())
		return
	}

	m = &ast.YTMethod{
		DefPos: tokFunc.Pos,
		YTDoc:  ctx.PreDoc(tokFunc.Line),
		Name:   tokFunc.IDValue(),
		Request: &ast.YTMessage{
			Name:   tokRQ.IDValue(),
			DefPos: tokRQ.Pos,
			Fields: []*ast.YTField{
				{Name: "rq", Type: rqField.Type, No: 1},
			},
		},
		Reply: &ast.YTMessage{
			DefPos: tokRS.Pos,
			Name:   tokRS.IDValue(),
			Fields: []*ast.YTField{
				{Name: "rs", Type: rsField.Type, No: 1},
			},
		},
	}
	// method no.
	if a5 != nil {
		tokNo := a5.(*token.Token)
		v, err := tokNo.Int64Value()
		if err != nil {
			return nil, ast.NewError2(tokNo, err)
		}
		// 仅在配置开启情况下保存数据
		if ast.Flag.ServiceUseMethodID {
			m.No = &ast.YTMethodNo{
				DefPos: tokNo.Pos,
				Value:  &v,
			}
		}
	}
	// options
	if a6 != nil {
		m.Opts = a6.(*ast.YTOptions).Opts
	}

	ctx.LastElement = m
	return
}

// 项目定义
// Project:"project" tok_identifier "{" ProjElements "}" OptEnd	<< bridge.NewProject($Context, $1, $3) >>
func NewProject(c, a1, a3 interface{}) (proj *ast.YTProject, err error) {
	ctx := c.(*ast.Context)
	tokName := a1.(*token.Token)

	err = checkNormalIdentifier(tokName.IDValue(), "project name")
	if err != nil {
		err = ast.NewError2(tokName, err)
		return
	}

	proj = a3.(*ast.YTProject)
	proj.YTDoc = ctx.PreDoc(tokName.Line)
	proj.DefPos = tokName.Pos
	proj.Name = tokName.IDValue()

	ctx.Prog.Projects = append(ctx.Prog.Projects, proj)

	ctx.LastElement = proj
	return
}

// ProjElements ProjArea					<< bridge.ProjectArea($Context, $0, $1) >>
func ProjectArea(c, a0, a1 interface{}) (proj *ast.YTProject, err error) {
	ctx := c.(*ast.Context)
	proj = a1.(*ast.YTProject)
	area := a1.(*token.Token)

	err = checkNormalIdentifier(area.IDValue(), "project area name")
	if err != nil {
		err = ast.NewError2(area, err)
		return
	}

	proj.Area = area.IDValue()
	ctx.LastElement = area
	return
}

// ProjElements OptionExpr				  	<< bridge.ProjectOption($Context, $0, $1) >>
func ProjectOption(c, a0, a1 interface{}) (proj *ast.YTProject, err error) {
	ctx := c.(*ast.Context)
	proj = a1.(*ast.YTProject)
	opt := a1.(*ast.YTOption)

	// if proj.Area == "" {}

	if proj.Conf[proj.Area] == nil {
		proj.Conf[proj.Area] = &ast.YTOptions{}
	}
	proj.Conf[proj.Area].Opts = append(proj.Conf[proj.Area].Opts, opt)
	ctx.LastElement = opt
	return
}

// 检测是否是正常的标识符 a-z 0-9 _
func checkNormalIdentifier(def, tip string) error {
	for k, r := range def {
		if k == 0 {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				continue
			}
			return fmt.Errorf("%s invalid char '%c', identifier need lowercase and character first", tip, r)
		}
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r == '_') {
			continue
		}
		return fmt.Errorf("%s invalid char '%c', identifier need lowercase,number or '_'", tip, r)
	}
	return nil
}

// 检测是否是合法的option name定义
func checkOptionName(def, tip string) error {
	lastDot := -100
	for k, r := range def {
		if k == 0 {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				continue
			}
			return fmt.Errorf("%s invalid char '%c', option name need lowercase and character first", tip, r)
		}
		if r == '.' {
			// 连续的'.' 无效
			if lastDot+1 == k {
				return fmt.Errorf("%s consecutive '.' invalid", tip)
			}
			lastDot = k
			continue
		}
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r == '_') {
			continue
		}
		return fmt.Errorf("%s invalid char '%c', option need lowercase,number or '_'", tip, r)
	}
	return nil
}

// 检测是否是路径
func checkFilePath(def, tip string) error {
	return nil
}

// 检测是否是合法的type定义
func analyseType(def, tip string) (ftyp *ast.YTField, err error) {
	if def != strings.ToLower(def) {
		err = fmt.Errorf("%s need lowercase", tip)
		return
	}
	switch def {
	case "int32":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeInt32,
			},
		}
	case "int64":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeInt64,
			},
		}

	case "sint32":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeInt32,
			},
			YTOptions: ast.YTOptions{
				Opts: []*ast.YTOption{
					{
						DefPos: token.Pos{},
						Key:    ProtobufOptionSigned,
						Value: &ast.YTOptionValue{
							IntVal: intvar(1),
						},
					},
				},
			},
		}
	case "sint64":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeInt64,
			},
			YTOptions: ast.YTOptions{
				Opts: []*ast.YTOption{
					{
						DefPos: token.Pos{},
						Key:    ProtobufOptionSigned,
						Value: &ast.YTOptionValue{
							IntVal: intvar(1),
						},
					},
				},
			},
		}

	case "sfixed32":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeInt32,
			},
			YTOptions: ast.YTOptions{
				Opts: []*ast.YTOption{
					{
						DefPos: token.Pos{},
						Key:    ProtobufOptionSigned,
						Value: &ast.YTOptionValue{
							IntVal: intvar(1),
						},
					},
					{
						DefPos: token.Pos{},
						Key:    ProtobufOptionFixed,
						Value: &ast.YTOptionValue{
							IntVal: intvar(1),
						},
					},
				},
			},
		}
	case "sfixed64":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeInt64,
			},
			YTOptions: ast.YTOptions{
				Opts: []*ast.YTOption{
					{
						DefPos: token.Pos{},
						Key:    ProtobufOptionSigned,
						Value: &ast.YTOptionValue{
							IntVal: intvar(1),
						},
					},
					{
						DefPos: token.Pos{},
						Key:    ProtobufOptionFixed,
						Value: &ast.YTOptionValue{
							IntVal: intvar(1),
						},
					},
				},
			},
		}

	case "uint32":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeUint32,
			},
		}
	case "uint64":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeUint64,
			},
		}

	case "fixed32":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeUint32,
			},
			YTOptions: ast.YTOptions{
				Opts: []*ast.YTOption{
					{
						DefPos: token.Pos{},
						Key:    ProtobufOptionFixed,
						Value: &ast.YTOptionValue{
							IntVal: intvar(1),
						},
					},
				},
			},
		}
	case "fixed64":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeUint64,
			},
			YTOptions: ast.YTOptions{
				Opts: []*ast.YTOption{
					{
						DefPos: token.Pos{},
						Key:    ProtobufOptionFixed,
						Value: &ast.YTOptionValue{
							IntVal: intvar(1),
						},
					},
				},
			},
		}
	case "string":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeString,
			},
		}
	case "bytes":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeBinary,
			},
		}
	case "float":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeFloat32,
			},
		}
	case "double":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeFloat64,
			},
		}
	case "bool":
		ftyp = &ast.YTField{
			Type: &ast.YTFieldType{
				YTBaseType: ast.BaseTypeBool,
			},
		}

	default:
		// custom type check
		e1 := checkNormalIdentifier(def, "")
		e2 := checkOptionName(def, "")

		if e1 == nil || e2 == nil {
			ftyp = &ast.YTField{
				Type: &ast.YTFieldType{
					YTCustomType: &ast.YTCustomType{
						Name: def,
					},
				},
			}
		}
	}

	if ftyp != nil {
		return
	}

	// 无效的类型
	return nil, fmt.Errorf("%s invalid field type [%s]", tip, def)
}

func intvar(n int64) *int64 {
	return &n
}
