package jsondata

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aggronmagi/wctl/xlsx/gen"
	"github.com/aggronmagi/wctl/xlsx/parser"
)

var (
	// 本地配置
	cfg = struct {
	}{}
	// 语言配置
	language = gen.NewExportConfig("json",
		gen.WithCheckOptions(checkLanguageConfig),
		gen.WithExportData(exportLanguageData),
	)
)

// 注册语言函数
func Language() *gen.ExportSupportConfig {
	// 注册参数
	// language.StringVar(&cfg.Xxx, "xx", cfg.Xxx, "说明")
	// 返回语言
	return language
}

// 参数检测
func checkLanguageConfig() error {
	return nil
}

// 导出数据
func exportLanguageData(sheet *parser.XlsxSheet, opts *gen.ExportOption) (err error) {
	// log.Println("sheet cache export", sheet.SheetName, sheet.FromFile)
	var sheetData = sheet.AllData
	var sheetType = sheet.AllType
	var cache interface{}
	if sheet.KVFlag {
		row := make(map[string]interface{})
		for col, field := range sheetType {
			if !field.EnableExport(opts.ExportFlag) {
				continue
			}
			row[field.Name] = sheetData[0][col].Value
		}
		cache = row
	} else {
		slices := make([]interface{}, 0, len(sheetData))
		for _, rows := range sheetData {
			row := make(map[string]interface{})
			for col, field := range sheetType {
				if !field.EnableExport(opts.ExportFlag) {
					continue
				}
				row[field.Name] = rows[col].Value
			}
			slices = append(slices, row)
		}
		cache = map[string]interface{}{"data": slices}

	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json failed,%w", err)
	}
	err = gen.WriteFile(filepath.Join(opts.Outpath, strings.ToLower(sheet.StructName)+".json"), data)
	if err != nil {
		return fmt.Errorf("write json file failed,%w", err)
	}
	return
}
