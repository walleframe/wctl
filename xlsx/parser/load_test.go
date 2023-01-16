package parser

import "testing"

func TestLoad(t *testing.T) {
	RegisterDefaultChecker()
	RegisterDefaultType()
	// ~/
	datas, check, err := LoadXlsx("/Users/aggron/Downloads/Food_Fail.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(datas)
	t.Log(check)
	
}
