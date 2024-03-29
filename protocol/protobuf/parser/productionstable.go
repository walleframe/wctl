// Code generated by gocc; DO NOT EDIT.

package parser

import (
	"github.com/walleframe/wctl/protocol/protobuf/bridge"
	"github.com/walleframe/wctl/protocol/ast"
)

type (
	ProdTab      [numProductions]ProdTabEntry
	ProdTabEntry struct {
		String     string
		Id         string
		NTType     int
		Index      int
		NumSymbols int
		ReduceFunc func([]Attrib, interface{}) (Attrib, error)
	}
	Attrib interface {
	}
)

var productionsTable = ProdTab{
	ProdTabEntry{
		String: `S' : ProtocolDefine	<<  >>`,
		Id:         "S'",
		NTType:     0,
		Index:      0,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `ProtocolDefine : Syntax Package Imports Defines	<< bridge.NewProtocol(C, X[0], X[1], X[2]) >>`,
		Id:         "ProtocolDefine",
		NTType:     1,
		Index:      1,
		NumSymbols: 4,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewProtocol(C, X[0], X[1], X[2])
		},
	},
	ProdTabEntry{
		String: `OptEnd : empty	<<  >>`,
		Id:         "OptEnd",
		NTType:     2,
		Index:      2,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String: `OptEnd : ";"	<<  >>`,
		Id:         "OptEnd",
		NTType:     2,
		Index:      3,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Syntax : "syntax" "=" tok_literal ";"	<< bridge.ProtoSyntax(C, X[2]) >>`,
		Id:         "Syntax",
		NTType:     3,
		Index:      4,
		NumSymbols: 4,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.ProtoSyntax(C, X[2])
		},
	},
	ProdTabEntry{
		String: `Package : "package" tok_identifier OptEnd	<< bridge.NewPackage(C, X[1]) >>`,
		Id:         "Package",
		NTType:     4,
		Index:      5,
		NumSymbols: 3,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewPackage(C, X[1])
		},
	},
	ProdTabEntry{
		String: `Imports : empty	<<  >>`,
		Id:         "Imports",
		NTType:     5,
		Index:      6,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String: `Imports : Imports Import	<<  >>`,
		Id:         "Imports",
		NTType:     5,
		Index:      7,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Import : "import" tok_literal OptEnd	<< bridge.NewImport(C, X[1], "") >>`,
		Id:         "Import",
		NTType:     6,
		Index:      8,
		NumSymbols: 3,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewImport(C, X[1], "")
		},
	},
	ProdTabEntry{
		String: `Import : "import" tok_identifier tok_literal OptEnd	<< bridge.NewImport(C, X[2], X[1]) >>`,
		Id:         "Import",
		NTType:     6,
		Index:      9,
		NumSymbols: 4,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewImport(C, X[2], X[1])
		},
	},
	ProdTabEntry{
		String: `Defines : empty	<<  >>`,
		Id:         "Defines",
		NTType:     7,
		Index:      10,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String: `Defines : Defines Define	<<  >>`,
		Id:         "Defines",
		NTType:     7,
		Index:      11,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Define : Enum	<<  >>`,
		Id:         "Define",
		NTType:     8,
		Index:      12,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Define : Message	<<  >>`,
		Id:         "Define",
		NTType:     8,
		Index:      13,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Define : Option	<<  >>`,
		Id:         "Define",
		NTType:     8,
		Index:      14,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Define : Service	<<  >>`,
		Id:         "Define",
		NTType:     8,
		Index:      15,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Enum : "enum" tok_identifier "{" EnumValues "}" OptEnd	<< bridge.NewEnum(C, X[1], X[3]) >>`,
		Id:         "Enum",
		NTType:     9,
		Index:      16,
		NumSymbols: 6,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewEnum(C, X[1], X[3])
		},
	},
	ProdTabEntry{
		String: `EnumValues : empty	<<  >>`,
		Id:         "EnumValues",
		NTType:     10,
		Index:      17,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String: `EnumValues : EnumValues EnumValue	<< bridge.AppendOption(C, X[0], X[1]) >>`,
		Id:         "EnumValues",
		NTType:     10,
		Index:      18,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.AppendOption(C, X[0], X[1])
		},
	},
	ProdTabEntry{
		String: `EnumValue : tok_identifier OptionValue OptEnd	<< bridge.OptionExpr(C, X[0], X[1]) >>`,
		Id:         "EnumValue",
		NTType:     11,
		Index:      19,
		NumSymbols: 3,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.OptionExpr(C, X[0], X[1])
		},
	},
	ProdTabEntry{
		String: `OptionValue : empty	<<  >>`,
		Id:         "OptionValue",
		NTType:     12,
		Index:      20,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String: `OptionValue : "=" tok_num	<< X[1], nil >>`,
		Id:         "OptionValue",
		NTType:     12,
		Index:      21,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[1], nil
		},
	},
	ProdTabEntry{
		String: `Option : "option" tok_identifier "=" tok_literal ";"	<< bridge.ProtoNewOption(C, X[1], X[3]) >>`,
		Id:         "Option",
		NTType:     13,
		Index:      22,
		NumSymbols: 5,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.ProtoNewOption(C, X[1], X[3])
		},
	},
	ProdTabEntry{
		String: `Message : "message" tok_identifier "{" Fields "}" OptEnd	<< bridge.NewMessage(C, X[1], X[3]) >>`,
		Id:         "Message",
		NTType:     14,
		Index:      23,
		NumSymbols: 6,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewMessage(C, X[1], X[3])
		},
	},
	ProdTabEntry{
		String: `Fields : empty	<< &ast.YTMessage{},nil >>`,
		Id:         "Fields",
		NTType:     15,
		Index:      24,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return &ast.YTMessage{},nil
		},
	},
	ProdTabEntry{
		String: `Fields : Fields FieldExpr	<< bridge.FieldField(C, X[0], X[1]) >>`,
		Id:         "Fields",
		NTType:     15,
		Index:      25,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.FieldField(C, X[0], X[1])
		},
	},
	ProdTabEntry{
		String: `Fields : Fields Message	<< bridge.FieldMessage(C, X[0], X[1]) >>`,
		Id:         "Fields",
		NTType:     15,
		Index:      26,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.FieldMessage(C, X[0], X[1])
		},
	},
	ProdTabEntry{
		String: `FieldExpr : FieldType tok_identifier "=" tok_num OptEnd	<< bridge.NewField(C,X[0], X[1], X[3], nil) >>`,
		Id:         "FieldExpr",
		NTType:     16,
		Index:      27,
		NumSymbols: 5,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewField(C,X[0], X[1], X[3], nil)
		},
	},
	ProdTabEntry{
		String: `FieldType : "map" "<" tok_identifier "," tok_identifier ">"	<< bridge.MapType(C, X[2], X[4]) >>`,
		Id:         "FieldType",
		NTType:     17,
		Index:      28,
		NumSymbols: 6,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.MapType(C, X[2], X[4])
		},
	},
	ProdTabEntry{
		String: `FieldType : "repeated" tok_identifier	<< bridge.ArrayType(C, X[1]) >>`,
		Id:         "FieldType",
		NTType:     17,
		Index:      29,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.ArrayType(C, X[1])
		},
	},
	ProdTabEntry{
		String: `FieldType : tok_identifier	<< bridge.BasicOrCustomType(C, X[0]) >>`,
		Id:         "FieldType",
		NTType:     17,
		Index:      30,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.BasicOrCustomType(C, X[0])
		},
	},
	ProdTabEntry{
		String: `Service : "service" tok_identifier "{" ServiceElements "}"	<< bridge.NewService(C, X[1], X[3]) >>`,
		Id:         "Service",
		NTType:     18,
		Index:      31,
		NumSymbols: 5,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewService(C, X[1], X[3])
		},
	},
	ProdTabEntry{
		String: `ServiceElements : empty	<<  >>`,
		Id:         "ServiceElements",
		NTType:     19,
		Index:      32,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String: `ServiceElements : ServiceElements Method	<< bridge.ServiceMethod(C, X[0], X[1]) >>`,
		Id:         "ServiceElements",
		NTType:     19,
		Index:      33,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.ServiceMethod(C, X[0], X[1])
		},
	},
	ProdTabEntry{
		String: `Method : "rpc" tok_identifier "(" tok_identifier ")" "returns" "(" tok_identifier ")" "{" "}"	<< bridge.NewMethod(C, X[1], X[3], X[7], nil, nil) >>`,
		Id:         "Method",
		NTType:     20,
		Index:      34,
		NumSymbols: 11,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return bridge.NewMethod(C, X[1], X[3], X[7], nil, nil)
		},
	},
}
