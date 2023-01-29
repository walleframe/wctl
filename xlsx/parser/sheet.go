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
	Flags   Flag         // 字段标记
}

func NewField(typ, name, commonts, filter string) (*ColumnType, error) {
	// #开头的类型,直接忽略本列
	if strings.HasPrefix(typ, "#") ||
		strings.HasPrefix(name, "#") ||
		(typ == "" && name == "" && filter == "") { // 除注释之外都是空的,也忽略
		return nil, nil
	}
	// 拆分可能的标记
	name, flags, err := SplitFlags(name)
	if err != nil {
		return nil, err
	}
	// pm 忽略
	if flags.Pm {
		return nil, nil
	}
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
		Flags:   flags,
	}, nil
}

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

// EnableExport 是否允许导出
func (filed *ColumnType) EnableExport(flag ExportFlag) bool {
	if filed.Flags.Client && flag == ExportServer {
		return false
	}
	if filed.Flags.Server && flag == ExportClient {
		return false
	}
	return true
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
	AllType []*ColumnType
	// 数据 索引分别是 行,列
	AllData [][]*XlsxCell
	// 单结构体标记
	KVFlag bool
	// sheet级标记
	Flag Flag
}

// XlsxCheckSheet lua数据检测
type XlsxCheckSheet struct {
	FromFile   string
	Sheet      string
	LuaScripts map[string]string
}
