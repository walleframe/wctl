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
		ClientMode bool
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
	language.BoolVar(&cfg.ClientMode, "client-mode", cfg.ClientMode, "生成客户端数据,默认生成服务器数据")
	// 返回语言
	return language
}

// 参数检测
func checkLanguageConfig() error {
	return nil
}

// 导出数据
func exportLanguageData(sheet *parser.XlsxSheet, outpath string) (err error) {
	// log.Println("sheet cache export", sheet.SheetName, sheet.FromFile)
	var sheetData [][]*parser.XlsxCell
	var sheetType []*parser.ColumnType
	if cfg.ClientMode {
		sheetData = sheet.ClientData()
		sheetType = sheet.ClientType()
	} else {
		sheetData = sheet.ServerData()
		sheetType = sheet.ServerType()
	}
	var cache interface{}
	if sheet.KVFlag {
		row := make(map[string]interface{})
		for col, field := range sheetType {
			row[field.Name] = sheetData[0][col].Value
		}
		cache = row
	} else {
		slices := make([]interface{}, 0, len(sheetData))
		for _, rows := range sheetData {
			row := make(map[string]interface{})
			for col, field := range sheetType {
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
	err = gen.WriteFile(filepath.Join(outpath, strings.ToLower(sheet.StructName)+".json"), data)
	if err != nil {
		return fmt.Errorf("write json file failed,%w", err)
	}
	return
}
