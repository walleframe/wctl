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
	Flags   ExportTags   // 字段标记
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
	if !flags.Client() && !flags.Server() && !flags.ID() {
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
func (filed *ColumnType) EnableExport(flag ExportTags) bool {
	return (filed.Flags & flag) > 0
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
	Flag ExportTags

	// 数据索引项
	cache struct {
		// id cache
		idTypes []*ColumnType

		// server side cache
		svcType []*ColumnType
		svcData [][]*XlsxCell
		// client cache
		clientType []*ColumnType
		clientData [][]*XlsxCell
	}
}

// EnableExport 是否允许导出
func (sheet *XlsxSheet) EnableExport(flag ExportTags) bool {
	return (sheet.Flag & flag) > 0
}

func (sheet *XlsxSheet) IDTypes() []*ColumnType {
	if sheet.cache.idTypes != nil {
		return sheet.cache.idTypes
	}
	sheet.cache.idTypes = make([]*ColumnType, 0, 2)

	for _, c := range sheet.AllType {
		if c.Flags.ID() {
			sheet.cache.idTypes = append(sheet.cache.idTypes, c)
		}
	}

	return sheet.cache.idTypes
}

func (sheet *XlsxSheet) ExportType(tag ExportTags) []*ColumnType {
	if tag.All() && sheet.Flag.All() {
		return sheet.AllType
	}
	if tag.Client() && sheet.Flag.Client() {
		return sheet.ClientType()
	}
	if tag.Server() && sheet.Flag.Server() {
		return sheet.ServerType()
	}
	return nil
}

func (sheet *XlsxSheet) ExportData(tag ExportTags) [][]*XlsxCell {
	if tag.All() && sheet.Flag.All() {
		return sheet.AllData
	}
	if tag.Client() && sheet.Flag.Client() {
		return sheet.ClientData()
	}
	if tag.Server() && sheet.Flag.Server() {
		return sheet.ServerData()
	}
	return nil
}

func (sheet *XlsxSheet) ServerType() []*ColumnType {
	if !sheet.Flag.Server() {
		return nil
	}

	if sheet.cache.svcType != nil {
		return sheet.cache.svcType
	}

	for _, c := range sheet.AllType {
		if c.Flags.Server() {
			sheet.cache.svcType = append(sheet.cache.svcType, c)
		}
	}

	return sheet.cache.svcType
}

func (sheet *XlsxSheet) ServerData() [][]*XlsxCell {
	if !sheet.Flag.Server() {
		return nil
	}

	if sheet.cache.svcData != nil {
		return sheet.cache.svcData
	}

	for row, rows := range sheet.AllData {
		sheet.cache.svcData = append(sheet.cache.svcData, make([]*XlsxCell, 0, len(sheet.cache.svcType)))
		for col, typ := range sheet.AllType {
			cell := rows[col]
			if typ.Flags.Server() {
				sheet.cache.svcData[row] = append(sheet.cache.svcData[row], cell)
			}
		}
	}

	return sheet.cache.svcData
}

func (sheet *XlsxSheet) ClientType() []*ColumnType {
	if !sheet.Flag.Client() {
		return nil
	}

	if sheet.cache.clientType != nil {
		return sheet.cache.clientType
	}

	for _, c := range sheet.AllType {
		if c.Flags.Client() {
			sheet.cache.clientType = append(sheet.cache.clientType, c)
		}
	}

	return sheet.cache.clientType
}

func (sheet *XlsxSheet) ClientData() [][]*XlsxCell {
	if !sheet.Flag.Client() {
		return nil
	}

	if sheet.cache.clientData != nil {
		return sheet.cache.clientData
	}

	for row, rows := range sheet.AllData {
		sheet.cache.clientData = append(sheet.cache.clientData, make([]*XlsxCell, 0, len(sheet.cache.clientType)))
		for col, typ := range sheet.AllType {
			cell := rows[col]
			if typ.Flags.Client() {
				sheet.cache.clientData[row] = append(sheet.cache.clientData[row], cell)
			}
		}
	}

	return sheet.cache.clientData
}

// XlsxCheckSheet lua数据检测
type XlsxCheckSheet struct {
	FromFile   string
	Sheet      string
	LuaScripts map[string]string
}
