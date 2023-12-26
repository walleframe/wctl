package wpb

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/walleframe/wctl/xlsx/parser"
)

var UseFuncMap = template.FuncMap{}

func init() {
	UseFuncMap["PBTag"] = func(fieldIndex int) string {
		return fmt.Sprintf("= %d", fieldIndex+1)
	}
	UseFuncMap["TypeName"] = func(p *parser.ColumnType) string {
		pbSupport := p.Type.(parser.PBSupport)
		name, err := pbSupport.PBTypeName()
		if err != nil {
			log.Println("get pb type failed,", p)
		}

		return name
	}
	UseFuncMap["Comment"] = func(filed *parser.ColumnType) string {
		comment := strings.Replace(filed.Comment, "\r", "", -1)
		comment = strings.Replace(comment, "\n", " ", -1)
		return fmt.Sprintf("// %s %s", filed.Name, comment)
	}
	UseFuncMap["ToSnake"] = strcase.ToSnake
}
