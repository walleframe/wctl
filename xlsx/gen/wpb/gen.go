package wpb

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"text/template"

	"github.com/walleframe/wctl/xlsx/gen"
	"github.com/walleframe/wctl/xlsx/parser"
)

var (
	// 本地配置
	cfg = struct {
	}{}
	// 语言配置
	language = gen.NewExportConfig("wpb",
		gen.WithExportDefine(exportDefine),
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
	var temp = template.New("pb")
	temp.Funcs(UseFuncMap)
	t, err := temp.Parse(textTemplate)
	if err != nil {
		log.Println("pb template parse err", err)
		return
	}
	// 如果有特殊模板规则
	bf := &bytes.Buffer{}
	err = t.Execute(bf, &struct {
		*parser.XlsxSheet
		Columns     []*parser.ColumnType
		PackageName string
	}{
		XlsxSheet:   sheet,
		Columns:     sheet.ExportType(opts.ExportFlag),
		PackageName: strings.ToLower(filepath.Base(opts.Outpath)),
	})
	if err != nil {
		log.Println("pb template execute err", err)
		return
	}

	err = gen.WriteFile(filepath.Join(opts.Outpath, strings.ToLower(sheet.StructName)+".wproto"), bf.Bytes())
	if err != nil {
		return fmt.Errorf("write wpb file failed,%w", err)
	}
	return
}

const textTemplate = `
// generate by wctl xlsx. DO NOT EDIT.
package {{.PackageName}};

// {{.StructName}} generate from {{.SheetName}} in {{.FromFile}}
message {{ToSnake .StructName}}
{ {{range $fi,$typ := $.Columns }}
    {{Comment $typ}}
	{{TypeName $typ}} {{ToSnake $typ.Name}} {{PBTag $fi}}; {{end}}
}

{{- if not .KVFlag }}
// {{.StructName}} generate from {{.SheetName}} in {{.FromFile}}
message {{ToSnake .StructName}}_container
{	
	repeated {{ToSnake .StructName}} data = {{1}}; // table 
}
{{ end -}}
`
