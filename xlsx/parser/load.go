package parser

import (
	"fmt"
	"log"
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
				if errs == nil {
					errs = err
				}
				log.Println("parse ", fname, "sheet", sheet.Name, "failed")
				for _, v := range multierr.Errors(err) {
					log.Println("\t", v)
				}
				continue
			}
			datas = append(datas, value)
		}
		// key/value 单结构体
		if strings.HasSuffix(sheet.Name, "_vert") {
			value, err := ParseKVXlsx(fname, sheet)
			if err != nil {
				if errs == nil {
					errs = err
				}
				log.Println("parse ", fname, "sheet", sheet.Name, "failed")
				for _, v := range multierr.Errors(err) {
					log.Println("\t", v)
				}
				continue
			}
			datas = append(datas, value)
		}
		// lua check
		if strings.HasSuffix(sheet.Name, "_check") {
			value, err := ParseCheckXlsx(fname, sheet)
			if err != nil {
				if errs == nil {
					errs = err
				}
				log.Println("parse ", fname, "sheet", sheet.Name, "failed")
				for _, v := range multierr.Errors(err) {
					log.Println("\t", v)
				}
				continue
			}
			check = append(check, value)
		}
	}
	return
}

// getCellValue 获取xlsx单元格数据
func getCellValue(sheet *xlsx.Sheet, row, col int, typ Type) (ret string) {
	// 兼容空行,空单元格
	if row >= len(sheet.Rows) {
		return ret
	}
	// 兼容空行,空单元格
	if col >= len(sheet.Rows[row].Cells) {
		return ret
	}
	c := sheet.Rows[row].Cells[col]
	// 浮点数单元格按原样输出
	if typ != nil && strings.Contains(typ.Name(), "float") {
		ret, _ = c.GeneralNumeric()
		ret = strings.TrimSpace(ret)
	} else {
		// 取列头所在列和当前行交叉的单元格
		ret = strings.TrimSpace(c.Value)
	}
	return
}

func emptyRow(row *xlsx.Row) bool {
	for _, cell := range row.Cells {
		if cell == nil {
			continue
		}
		if len(cell.Value) == 0 {
			continue
		}
		if strings.TrimSpace(cell.Value) == "" {
			continue
		}
		return false
	}
	return true
}

// ParseDataXlsx 解析数据表格 数组或者map
func ParseDataXlsx(fromFile string, sheet *xlsx.Sheet) (data *XlsxSheet, errs error) {
	data = &XlsxSheet{
		SheetName:  sheet.Name,
		StructName: strings.TrimSuffix(sheet.Name, "_config"),
		FromFile:   fromFile,
		allType:    []*ColumnType{},
		allData:    [][]*XlsxCell{},
		KVFlag:     false,
	}
	if sheet.MaxRow < 4 {
		return nil, fmt.Errorf("sheet format invliad. row less then 4(fieldName,type,options,comment)")
	}

	// 解析文件头,分析类型
	columns := make([]int, 0, sheet.MaxCol)
	for col := 0; col < sheet.MaxCol; col++ {
		fieldComment := getCellValue(sheet, 0, col, nil)
		fieldName := getCellValue(sheet, 1, col, nil)
		fieldType := getCellValue(sheet, 2, col, nil)
		filedOptions := getCellValue(sheet, 3, col, nil)
		// 忽略空列
		if fieldComment == "" && fieldName == "" && fieldType == "" && filedOptions == "" {
			continue
		}
		typ, err := NewField(fieldType, fieldName, fieldComment, filedOptions)
		if err != nil {
			// 合并错误,一次性检测
			errs = multierr.Append(errs, fmt.Errorf("%w. column %s",
				err, fieldName,
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
		if emptyRow(sheet.Rows[row]) {
			continue
		}
		data.allData = append(data.allData, make([]*XlsxCell, 0, len(columns)))

		for _, col := range columns {
			typ := data.allType[col]
			cell, err := typ.Parse(getCellValue(sheet, row, col, typ.Type))
			if err != nil { // 合并错误,一次性检测
				errs = multierr.Append(errs, fmt.Errorf("%w. column %s, row %d [%s]",
					err,
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
		StructName: strings.TrimSuffix(sheet.Name, "_vert"),
		FromFile:   fromFile,
		allType:    []*ColumnType{},
		allData:    [][]*XlsxCell{},
		KVFlag:     true,
	}
	if sheet.MaxCol < 5 {
		return nil, fmt.Errorf("sheet format invliad.  columns count less then 5(fieldName,type,options,comment,value)")
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
			errs = multierr.Append(errs, fmt.Errorf("%w. row %s",
				err,
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
		Sheet:      sheet.Name,
		LuaScripts: map[string]string{},
	}
	if sheet.MaxRow < 2 || sheet.MaxCol < 2 {
		return nil, fmt.Errorf("sheet format invliad. columns count less then 2(flag,lua_script)")
	}
	// 解析数据
	for row := 0; row < sheet.MaxRow; row++ {
		luaFlag := getCellValue(sheet, row, 0, nil)
		luaScript := getCellValue(sheet, row, 1, nil)
		if luaFlag == "" && luaScript == "" {
			continue
		}
		if luaFlag == "" || luaScript == "" {
			errs = multierr.Append(errs, fmt.Errorf("row %d invalid, tag:[%s] script:[%s]", row, luaFlag, luaScript))
		}
		if _, ok := data.LuaScripts[luaFlag]; ok {
			errs = multierr.Append(errs, fmt.Errorf("lua tag repeated. tag [%s] row %d repeated", luaFlag, row))
			return
		}
		data.LuaScripts[luaFlag] = luaScript
	}
	if errs != nil {
		return
	}
	return
}
