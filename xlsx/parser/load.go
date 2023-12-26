package parser

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/tealeg/xlsx"
	"go.uber.org/multierr"
)

// LoadXlsx 加载数据
func LoadXlsx(fname string) (datas []*XlsxSheet, check []*XlsxCheckSheet, errs error) {
	fname = strings.Replace(filepath.Clean(fname), "\\", "/", -1)
	file, err := xlsx.OpenFile(fname)
	if err != nil {
		return nil, nil, fmt.Errorf("open %s failed %w", fname, err)
	}
	for _, sheet := range file.Sheets {
		sheetName, flags, err := SplitFlags(sheet.Name)
		if err != nil {
			if errs != nil {
				err = nil
			}
			log.Println("parse ", fname, "sheet", sheet.Name, " flag failed")
			log.Println("\t", err)
			continue
		}
		if flags.ID() {
			log.Println("parse ", fname, "sheet", sheet.Name, " flag failed")
			log.Println("\tSheet Name Flag Not Support @id")
		}
		// 数据配置
		if strings.HasSuffix(sheetName, "_cfgs") {
			value, err := ParseDataXlsx(fname, sheet, sheetName, flags)
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
		if strings.HasSuffix(sheetName, "_st") {
			value, err := ParseKVXlsx(fname, sheet, sheetName, flags)
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
		if strings.HasSuffix(sheetName, "_lua") {
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
func ParseDataXlsx(fromFile string, sheet *xlsx.Sheet, sheetName string, flags ExportTags) (data *XlsxSheet, errs error) {
	data = &XlsxSheet{
		SheetName:  sheetName,
		StructName: strings.TrimSuffix(sheetName, "_cfgs"),
		FromFile:   fromFile,
		AllType:    []*ColumnType{},
		AllData:    [][]*XlsxCell{},
		KVFlag:     false,
		Flag:       flags,
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
		data.AllType = append(data.AllType, typ)
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
		data.AllData = append(data.AllData, make([]*XlsxCell, 0, len(columns)))

		for _, col := range columns {
			typ := data.AllType[col]
			cell, err := typ.Parse(getCellValue(sheet, row, col, typ.Type))
			if err != nil { // 合并错误,一次性检测
				errs = multierr.Append(errs, fmt.Errorf("%w. column %s, row %d [%s]",
					err,
					typ.Name, row+1, sheet.Rows[row].Cells[col].Value,
				))
				data.AllData[row-4] = append(data.AllData[row-4], nil)
				continue
			}
			data.AllData[row-4] = append(data.AllData[row-4], cell)
		}
	}
	return
}

// ParseKVXlsx 解析结构体表格
func ParseKVXlsx(fromFile string, sheet *xlsx.Sheet, sheetName string, flags ExportTags) (data *XlsxSheet, errs error) {
	data = &XlsxSheet{
		SheetName:  sheetName,
		StructName: strings.TrimSuffix(sheetName, "_st"),
		FromFile:   fromFile,
		AllType:    []*ColumnType{},
		AllData:    [][]*XlsxCell{},
		KVFlag:     true,
		Flag:       flags,
	}
	if sheet.MaxCol < 5 {
		return nil, fmt.Errorf("sheet format invliad.  columns count less then 5(fieldName,type,options,comment,value)")
	}
	// 解析文件头,分析类型
	rows := make([]int, 0, sheet.MaxRow)
	for row := 0; row < sheet.MaxRow; row++ {
		fieldComment := getCellValue(sheet, row, 0, nil)
		fieldName := getCellValue(sheet, row, 1, nil)
		fieldType := getCellValue(sheet, row, 2, nil)
		filedOptions := getCellValue(sheet, row, 3, nil)
		typ, err := NewField(fieldType, fieldName, fieldComment, filedOptions)
		if err != nil {
			// 合并错误,一次性检测
			errs = multierr.Append(errs, fmt.Errorf("%w. row %s",
				err,
				fieldName,
			))
			continue
		}
		data.AllType = append(data.AllType, typ)
		rows = append(rows, row)
	}
	if errs != nil {
		return
	}
	// 解析文件数据
	data.AllData = append(data.AllData, make([]*XlsxCell, 0, len(rows)))
	for _, row := range rows {
		typ := data.AllType[row]
		cell, err := typ.Parse(getCellValue(sheet, row, 4, typ.Type))
		if err != nil { // 合并错误,一次性检测
			errs = multierr.Append(errs, fmt.Errorf("%w. from %s sheet %s row %s, value [%s]",
				err,
				fromFile, sheet.Name,
				typ.Name, sheet.Rows[row].Cells[4].Value,
			))
			data.AllData[0] = append(data.AllData[0], nil)
			continue
		}
		data.AllData[0] = append(data.AllData[0], cell)
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
