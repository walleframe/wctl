package examples

import (
	"github.com/walleframe/wctl/xlsx/gen"
	"github.com/walleframe/wctl/xlsx/parser"
)

var (
	// 本地配置
	cfg = struct {
		ScriptPaths []string
	}{}
	// 语言配置
	language = gen.NewExportConfig("examples",
		gen.WithExportDefine(exportDefine),
		gen.WithExportMergeDefine(exportMergeDefine),
		gen.WithExportData(exportData),
		gen.WithExportMergeData(exportMergeData),
		gen.WithCheckOptions(checkOptionConfig),
	)
)

// 注册语言函数
func Language() *gen.ExportSupportConfig {
	// 返回语言
	return language
}

func checkOptionConfig() error {
	return nil
}

func exportDefine(sheet *parser.XlsxSheet, opts *gen.ExportOption) (err error) {
	return
}

// 生成代码
func exportMergeDefine(sheets []*parser.XlsxSheet, opts *gen.ExportOption) (err error) {

	return
}

func exportData(sheet *parser.XlsxSheet, opts *gen.ExportOption) (err error) {
	return
}

func exportMergeData(sheets []*parser.XlsxSheet, opts *gen.ExportOption) (err error) {
	return
}
