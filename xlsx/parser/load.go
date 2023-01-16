package parser

import (
	"fmt"
	"strings"

	"github.com/tealeg/xlsx"
	"go.uber.org/multierr"
)

// LoadXlsx 加载数据
func LoadXlsx(fname string) (datas []*XlsxSheet, check []*XlsxCheckSheet, errs error) {
	file, err := xlsx.OpenFile(fname)
	if err != nil {
		return nil, nil, fmt.Errorf("open %s failed %w", fname, err)
	}
	for _, sheet := range file.Sheets {
		// 数据配置
		if strings.HasSuffix(sheet.Name, "_config") {
			value, err := ParseDataXlsx(fname, sheet)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			datas = append(datas, value)
		}
		// key/value 单结构体
		if strings.HasSuffix(sheet.Name, "_vert") {
			value, err := ParseKVXlsx(fname, sheet)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			datas = append(datas, value)
		}
		// lua check
		if strings.HasSuffix(sheet.Name, "_check") {
			value, err := ParseCheckXlsx(fname, sheet)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			check = append(check, value)
		}
	}
	return
}

// getCellValue 获取xlsx单元格数据
func getCellValue(sheet *xlsx.Sheet, row, col int, typ Type) (ret string) {
	c := sheet.Rows[row].Cells[col]
	// 浮点数单元格按原样输出
	if strings.Contains(typ.Name(), "float") {
		ret, _ = c.GeneralNumeric()
		ret = strings.TrimSpace(ret)
	} else {
		// 取列头所在列和当前行交叉的单元格
		ret = strings.TrimSpace(c.Value)
	}
	return
}

// ParseDataXlsx 解析数据表格 数组或者map
func ParseDataXlsx(fromFile string, sheet *xlsx.Sheet) (data *XlsxSheet, errs error) {
	data = &XlsxSheet{
		SheetName:  sheet.Name,
		StructName: "",
		FromFile:   fromFile,
		allType:       []*ColumnType{},
		allData:       [][]*XlsxCell{},
		KVFlag:     false,
	}
	if sheet.MaxRow < 4 {
		return nil, fmt.Errorf("sheet format invliad. from %s sheet %s, row less then 4(fieldName,type,options,comment)", fromFile, sheet.Name)
	}
	// 解析文件头,分析类型
	columns := make([]int, 0, sheet.MaxCol)
	for col := 0; col < sheet.MaxCol; col++ {
		fieldName := strings.TrimSpace(sheet.Rows[0].Cells[col].Value)
		fieldType := strings.TrimSpace(sheet.Rows[1].Cells[col].Value)
		filedOptions := strings.TrimSpace(sheet.Rows[2].Cells[col].Value)
		fieldComment := strings.TrimSpace(sheet.Rows[3].Cells[col].Value)
		typ, err := NewField(fieldType, fieldName, fieldComment, filedOptions)
		if err != nil {
			// 合并错误,一次性检测
			errs = multierr.Append(errs, fmt.Errorf("%w. from %s sheet %s column %s",
				err, fromFile, sheet.Name,
				fieldName,
			))
			continue
		}
		data.allType = append(data.allType, typ)
		columns = append(columns, col)
	}
	if errs != nil {
		return
	}
	// 解析文件数据
	for row := 4; row < sheet.MaxRow; row++ {
		data.allData = append(data.allData, make([]*XlsxCell, 0, len(columns)))
		for _, col := range columns {
			typ := data.allType[col]
			cell, err := typ.Parse(getCellValue(sheet, row, col, typ.Type))
			if err != nil { // 合并错误,一次性检测
				errs = multierr.Append(errs, fmt.Errorf("%w. from %s sheet %s column %s, row %d [%s]",
					err,
					fromFile, sheet.Name,
					typ.Name, row+1, sheet.Rows[row].Cells[col].Value,
				))
				data.allData[row-4] = append(data.allData[row-4], nil)
				continue
			}
			data.allData[row-4] = append(data.allData[row-4], cell)
		}
	}
	return
}

// ParseKVXlsx 解析结构体表格
func ParseKVXlsx(fromFile string, sheet *xlsx.Sheet) (data *XlsxSheet, errs error) {
	data = &XlsxSheet{
		SheetName:  sheet.Name,
		StructName: "",
		FromFile:   fromFile,
		allType:       []*ColumnType{},
		allData:       [][]*XlsxCell{},
		KVFlag:     true,
	}
	if sheet.MaxCol < 5 {
		return nil, fmt.Errorf("sheet format invliad. from %s sheet %s, columns count less then 5(fieldName,type,options,comment,value)", fromFile, sheet.Name)
	}
	// 解析文件头,分析类型
	rows := make([]int, 0, sheet.MaxRow)
	for row := 0; row < sheet.MaxRow; row++ {
		fieldName := strings.TrimSpace(sheet.Rows[row].Cells[0].Value)
		fieldType := strings.TrimSpace(sheet.Rows[row].Cells[1].Value)
		filedOptions := strings.TrimSpace(sheet.Rows[row].Cells[2].Value)
		fieldComment := strings.TrimSpace(sheet.Rows[row].Cells[3].Value)
		typ, err := NewField(fieldType, fieldName, fieldComment, filedOptions)
		if err != nil {
			// 合并错误,一次性检测
			errs = multierr.Append(errs, fmt.Errorf("%w. from %s sheet %s row %s",
				err,
				fromFile, sheet.Name,
				fieldName,
			))
			continue
		}
		data.allType = append(data.allType, typ)
		rows = append(rows, row)
	}
	if errs != nil {
		return
	}
	// 解析文件数据
	data.allData = append(data.allData, make([]*XlsxCell, 0, len(rows)))
	for _, row := range rows {
		typ := data.allType[row]
		cell, err := typ.Parse(getCellValue(sheet, row, 4, typ.Type))
		if err != nil { // 合并错误,一次性检测
			errs = multierr.Append(errs, fmt.Errorf("%w. from %s sheet %s row %s, value [%s]",
				err,
				fromFile, sheet.Name,
				typ.Name, sheet.Rows[row].Cells[4].Value,
			))
			data.allData[0] = append(data.allData[0], nil)
			continue
		}
		data.allData[0] = append(data.allData[0], cell)
	}
	return
}

// ParseCheckXlsx 解析数据检测脚本表格(lua)
func ParseCheckXlsx(fromFile string, sheet *xlsx.Sheet) (data *XlsxCheckSheet, errs error) {
	data = &XlsxCheckSheet{
		FromFile:   fromFile,
		Sheet:      data.Sheet,
		LuaScripts: map[string]string{},
	}
	if sheet.MaxRow < 2 {
		return nil, fmt.Errorf("sheet format invliad. from %s sheet %s, columns count less then 2(flag,lua_script)", fromFile, sheet.Name)
	}
	// 解析数据
	for row := 0; row < sheet.MaxRow; row++ {
		luaFlag := strings.TrimSpace(sheet.Rows[row].Cells[0].Value)
		luaScript := strings.TrimSpace(sheet.Rows[row].Cells[1].Value)
		if _, ok := data.LuaScripts[luaFlag]; ok {
			errs = multierr.Append(errs, fmt.Errorf("lua flag repeated. from %s sheet %s flag [%s] row %d repeated", fromFile, sheet.Name, luaFlag, row))
		}
		data.LuaScripts[luaFlag] = luaScript
	}
	if errs != nil {
		return
	}
	return
}
