package parser

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrIgnoreColumn = errors.New("ingore this column")
)

type ColumnType struct {
	Type    Type         // 字段类型
	Name    string       // 字段名
	Comment string       // 注释信息
	chekers []ValueCheck // 格式及有效校验
	// flags   map[string]struct{} // 字段标记
}

func NewField(typ, name, commonts, filter string) (*ColumnType, error) {
	// #开头的类型,直接忽略本列
	if strings.HasPrefix(typ, "#") {
		return nil, ErrIgnoreColumn
	}
	// // 拆分可能的标记
	// typ, flags, err := SplitFlags(typ)
	// if err != nil {
	// 	return nil, err
	// }
	// // pm 忽略
	// if _, ok := flags["pm"]; ok {
	// 	return nil, ErrIgnoreColumn
	// }
	// 解析类型
	typDef, err := ParseType(typ)
	if err != nil {
		return nil, fmt.Errorf("parse type [%s] failed,%w", typ, err)
	}
	// 数据检测
	checkers, err := ParseCheker(filter)
	if err != nil {
		return nil, fmt.Errorf("parse checkers [%s] failed,%w", filter, err)
	}
	// 追加类型限定
	checkers = append(checkers, typDef.Checkers()...)

	return &ColumnType{
		Type:    typDef,
		Name:    name,
		Comment: commonts,
		chekers: checkers,
		//flags:   flags,
	}, nil
}

// func (field *ColumnType) HasFlag(flag string) bool {
// 	switch flag {

// 	}
// 	_, ok := field.flags[flag]
// 	return ok
// }

// Parse 解析单元格
func (field *ColumnType) Parse(cell string) (*XlsxCell, error) {
	replace, val, err := field.Type.Parse(cell)
	if err != nil {
		return nil, err
	}

	// 检测
	for _, checker := range field.chekers {
		err = checker(val)
		if err != nil {
			return nil, err
		}
	}
	return &XlsxCell{
		Raw:   replace,
		Value: val,
	}, nil
}

type XlsxCell struct {
	// 原始数据
	Raw string
	//
	Value interface{}
}

type XlsxSheet struct {
	// sheet 名
	SheetName string
	// 生成结构体名称
	StructName string
	// 来源文件
	FromFile string
	// 字段及类型
	allType []*ColumnType
	// 数据 索引分别是 行,列
	allData [][]*XlsxCell
	// 单结构体标记
	KVFlag bool

	// 缓存lua脚本错误
	errs []error
}

// func (sheet *XlsxSheet) Rows() int {
// 	return len(sheet.Data)
// }

// func (sheet *XlsxSheet) Columns() int {
// 	return len(sheet.Type)
// }

// ServerType 服务器导出类型
func (sheet *XlsxSheet) ServerType() []*ColumnType {
	return sheet.allType
}

// ServerData 服务器导出数据
func (sheet *XlsxSheet) ServerData() [][]*XlsxCell {
	return sheet.allData
}

// ClientType 客户端导出类型
func (sheet *XlsxSheet) ClientType() []*ColumnType {
	return sheet.allType
}

// ClientData 客户端导出数据
func (sheet *XlsxSheet) ClientData() [][]*XlsxCell {
	return sheet.allData
}

// // Language 导出语言支持
// type Language interface {
// 	// ExportType 是否支持导出类型声明文件 (lua这种动态语言,不需要类型生成)
// 	ExportType() bool
// 	// FieldType 对应语言的字段类型
// 	FieldType(Type) string
// 	// StructName 生成结构体名字
// 	StructName(sheet string) string
// 	// ValueDesc 值描述格式
// 	ValueDesc(cell *XlsxCell) string
// }

// XlsxCheckSheet lua数据检测
type XlsxCheckSheet struct {
	FromFile   string
	Sheet      string
	LuaScripts map[string]string
}
