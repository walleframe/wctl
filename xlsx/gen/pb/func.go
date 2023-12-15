package pb

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/walleframe/wctl/xlsx/parser"
)

var UseFuncMap = template.FuncMap{}

func init() {
	UseFuncMap["PBTag"] = func(fieldIndex int) string {
		var sb strings.Builder
		fmt.Fprintf(&sb, "= %d", fieldIndex+1)
		return sb.String()
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
}
