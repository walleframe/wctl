package luatpl

import (
	"log"
	"path/filepath"

	"github.com/walleframe/wctl/xlsx/gen"
	"github.com/walleframe/wctl/xlsx/parser"
	lua "github.com/yuin/gopher-lua"

	_ "embed"
)

//go:embed template.lua
var templateLuaScript string

var (
	// 本地配置
	cfg = struct {
		ScriptPaths []string
	}{}
	// 语言配置
	language = gen.NewExportConfig("luatpl",
		gen.WithExportMergeDefine(exportMergeDefine),
		gen.WithExportMergeData(exportMergeData),
		gen.WithCheckOptions(checkOptionConfig),
	)
)

// 注册语言函数
func Language() *gen.ExportSupportConfig {
	language.StringSliceVar(&cfg.ScriptPaths, "script-paths", cfg.ScriptPaths, "生成脚本路径")
	// 返回语言
	return language
}

func checkOptionConfig() error {
	return nil
}

// 生成代码
func exportMergeDefine(sheets []*parser.XlsxSheet, opts *gen.ExportOption) (err error) {
	for _, path := range cfg.ScriptPaths {
		l := lua.NewState(lua.Options{
			CallStackSize:       0,
			RegistrySize:        0,
			RegistryMaxSize:     0,
			RegistryGrowStep:    0,
			SkipOpenLibs:        false,
			IncludeGoStackTrace: false,
			MinimizeStackMemory: false,
		})
		// load template.lua
		err = l.DoString(templateLuaScript)
		if err != nil {
			log.Println("load embed template script failed", err)
			return err
		}
		// set sheet scripts
		setLuaState(l, sheets)
		// run
		err = l.DoFile(filepath.Join(path, "init.lua"))
		if err != nil {
			log.Println("do file failed", path, err)
			return err
		}
	}

	return
}

func exportMergeData(sheets []*parser.XlsxSheet, opts *gen.ExportOption) (err error) {
	return
}

func setLuaState(l *lua.LState, sheets []*parser.XlsxSheet, opts *gen.ExportOption) {
	tbl := l.NewTable()
	for _, sheet := range sheets {
		ud := l.NewUserData()
		ud.Value = sheet
		st := l.NewTable()
		l.SetField(st, "__index", l.NewClosure(sheetIndex, ud))
		tbl.Append(st)
	}
	l.SetGlobal("sheets", tbl)
}

func sheetIndex(l *lua.LState) int {
	return 0
}
