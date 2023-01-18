package parser

import (
	"errors"
	"log"

	"go.uber.org/multierr"
)

func LimitCheck(sheets []*XlsxSheet) (errs error) {
	sheetName := make(map[string]*XlsxSheet, len(sheets))
	for _, sheet := range sheets {
		// 同sheet名检测
		if last, ok := sheetName[sheet.SheetName]; ok {
			log.Println(sheet.SheetName, "sheet name repeat!", last.FromFile, " vs ", sheet.FromFile)
			errs = multierr.Append(errs, errors.New("sheet name repeated"))
			continue
		}
		sheetName[sheet.SheetName] = sheet
		// 同字段名检测
		fieldName := make(map[string]struct{}, len(sheet.allType))
		for _, v := range sheet.allType {
			if _, ok := fieldName[v.Name]; ok {
				log.Println(v.Name, " field name repeat, from ", sheet.FromFile, sheet.SheetName)
				errs = multierr.Append(errs, errors.New("filed name repeated"))
				continue
			}
			fieldName[v.Name] = struct{}{}
		}
	}
	return
}
