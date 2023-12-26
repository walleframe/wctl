package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/walleframe/wctl/xlsx/gen"
	"github.com/walleframe/wctl/xlsx/parser"
)

var (
	// 本地配置
	cfg = struct {
		// 模板配置
		Template string
		tpl      *template.Template
		// 导入的配置管理包名称
		ImportMgr string
		// PB包名称
		ImportPB string
	}{
		ImportMgr: "github.com/walleframe/svc_xlsx",
		ImportPB:  "",
	}
	// 语言配置
	language = gen.NewExportConfig("go",
		gen.WithExportDefine(exportGolangDefine),
		gen.WithCheckOptions(checkOptionConfig),
	)
)

// 注册语言函数
func Language() *gen.ExportSupportConfig {
	// 注册参数
	language.StringVar(&cfg.Template, "template-file", cfg.Template, "模板规则文件")
	language.StringVar(&cfg.ImportMgr, "import-mgr", cfg.ImportMgr, "导入配置管理器包")
	language.StringVar(&cfg.ImportPB, "import-pb", cfg.ImportPB, "导入的pb管理包")
	// 返回语言
	return language
}

func checkOptionConfig() (err error) {
	if cfg.Template != "" {
		cfg.tpl, err = template.New("go").Funcs(UseFuncMap).ParseFiles(cfg.Template)
		if err != nil {
			return fmt.Errorf("parse template failed [%s] %+v", cfg.Template, err)
		}
	} else {
		cfg.tpl, err = template.New("go").Funcs(UseFuncMap).Parse(textTemplate)
		if err != nil {
			return fmt.Errorf("parse inner template failed [%s] %+v", cfg.Template, err)
		}
	}
	cfg.ImportMgr = filepath.ToSlash(filepath.Clean(cfg.ImportMgr))
	if cfg.ImportMgr == "" {
		return fmt.Errorf("go-import-mgr empty")
	}
	cfg.ImportPB = filepath.ToSlash(filepath.Clean(cfg.ImportPB))
	if cfg.ImportPB == "" {
		return fmt.Errorf("go-import-pb empty")
	}
	return nil
}

func exportGolangDefine(sheet *parser.XlsxSheet, opts *gen.ExportOption) (err error) {
	// 检测配置
	columns := sheet.ExportType(opts.ExportFlag)
	if columns == nil {
		// not generate
		return nil
	}

	if err != nil {
		log.Println("go template parse err", err)
		return
	}

	pkgName := strings.ToLower(sheet.StructName)

	pkgName = fixPkgName(pkgName)

	//log.Println(sheet.FromFile)
	fileDir := strings.TrimSuffix(sheet.FromFile, filepath.Base(sheet.FromFile))
	//log.Println(fileDir)
	fileDir = strings.TrimPrefix(fileDir, strings.Replace(filepath.Dir(filepath.Dir(fileDir)), "\\", "/", -1))
	//log.Println(fileDir, filepath.Dir(filepath.Dir(fileDir)))
	fileDir = strings.TrimPrefix(fileDir, "/")
	fileDir = strings.TrimSuffix(fileDir, "/")
	fileDir = strings.TrimPrefix(fileDir, "\\")
	fileDir = strings.TrimSuffix(fileDir, "\\")

	// 如果有特殊模板规则
	bf := &bytes.Buffer{}
	err = cfg.tpl.Execute(bf, &struct {
		*parser.XlsxSheet
		PkgName     string
		ImportMgr   string
		ImportPB    string
		MgrPkg      string
		ProtoPkg    string
		ColumnTypes []*parser.ColumnType
		IDTypes     []*parser.ColumnType
		IDCnt       int
		FileDir     string
	}{
		XlsxSheet:   sheet,
		PkgName:     pkgName,
		ImportMgr:   cfg.ImportMgr,
		ImportPB:    cfg.ImportPB,
		MgrPkg:      filepath.Base(cfg.ImportMgr),
		ProtoPkg:    filepath.Base(cfg.ImportPB),
		ColumnTypes: columns,
		IDTypes:     sheet.IDTypes(),
		IDCnt:       len(sheet.IDTypes()),
		FileDir:     fileDir,
	})
	if err != nil {
		log.Println("go template execute err", err)
		return
	}
	data, err := format.Source(bf.Bytes())
	if err != nil {
		log.Println("format go code failed", err)
		data = bf.Bytes()
	}
	err = gen.WriteFile(path.Join(opts.Outpath, fmt.Sprintf("%s/%s.go", pkgName, strings.ToLower(sheet.StructName))), data)
	if err != nil {
		log.Println("go write file err", err)
		return
	}
	return
}

func fixPkgName(pkgName string) string {
	// 包名转小写后,可能会导致生成冲突
	switch pkgName {
	case "map":
		return pkgName + "_cfg"
	}
	return pkgName
}
