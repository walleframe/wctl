/*
Copyright © 2020 aggronmagi <czy463@163.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/walleframe/wctl/protocol/ast"
	"github.com/walleframe/wctl/protocol/wproto/lexer"
	"github.com/walleframe/wctl/protocol/wproto/parser"

	"github.com/stretchr/testify/assert"
	"github.com/walleframe/wctl/utils"
)

func parseWProtoSrc(src string) (*ast.YTProgram, error) {
	s := lexer.NewLexer([]byte(src))
	p := parser.NewParser()
	a, err := p.Parse(s)
	if err != nil {
		return nil, err
	}
	if val, ok := a.(*ast.YTProgram); ok {
		return val, nil
	}
	return nil, errors.New("invalid type")
}

func optInt(v int64) *ast.YTOptionValue {
	return &ast.YTOptionValue{
		IntVal: &v,
	}
}

func optString(v string) *ast.YTOptionValue {
	return &ast.YTOptionValue{
		Value: &v,
	}
}

func doc(v ...string) *ast.YTDoc {
	return &ast.YTDoc{
		Doc: v,
	}
}

func TestWProtoParse(t *testing.T) {

	if 1 == 0 {
		utils.Flag.ShowDetail = true
	}
	ast.RegisterRecursionAnalyser = ast.RecursionAnalyseFunc(func(file string) (prog *ast.YTProgram, err error) {
		prog = &ast.YTProgram{
			Pkg: &ast.YTPackage{
				Name: path.Base(file),
			},
		}
		return
	})
	testOnly := ""
	data := []struct {
		name string
		src  string
		data *ast.YTProgram
		err  bool
	}{
		{
			name: "only package",
			src: `// package doc
package x1`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					YTDoc: &ast.YTDoc{
						Doc: []string{"// package doc"},
					},
					Name: "x1",
				},
			},
		},
		{
			name: "no package(must error)",
			src: `
test.opt1
// opt comment
test.opt2 = 1
test.opt3 = "xx"
`,
			err: true,
		},
		{
			name: "package and options",
			src: `
// package doc
package x1

test.opt1
// opt comment
test.opt2 = 1
test.opt3 = "xx"
`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					YTDoc: doc("// package doc"),
					Name:  "x1",
				},
				YTOptions: ast.YTOptions{
					Opts: []*ast.YTOption{
						{
							Key: "test.opt1",
						},
						{
							YTDoc: doc("// opt comment"),
							Key:   "test.opt2",
							Value: optInt(1),
						},
						{
							Key:   "test.opt3",
							Value: optString("xx"),
						},
					},
				},
			},
		},
		{
			name: "import",
			src:  `package x1;import x2 "xxx/xx"; import "xx"`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				Imports: []*ast.YTImport{
					{
						AliasName: "x2",
						File:      "xxx/xx",
					},
					{
						File: "xx",
					},
				},
			},
		},
		{
			name: "enum",
			src:  `package x1;enum e1 {x=2;y;z=6}; enum e2{a;b;c;opt.xx}`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				EnumDefs: []*ast.YTEnumDef{
					{
						Name: "e1",
						Values: []*ast.YTEnumValue{
							{
								Name:  "x",
								Value: 2,
							},
							{
								Name:  "y",
								Value: 3,
							},
							{
								Name:  "z",
								Value: 6,
							},
						},
					},
					{
						Name: "e2",
						YTOptions: ast.YTOptions{
							Opts: []*ast.YTOption{
								{
									Key: "opt.xx",
								},
							},
						},
						Values: []*ast.YTEnumValue{
							{
								Name:  "a",
								Value: 0,
							},
							{
								Name:  "b",
								Value: 1,
							},
							{
								Name:  "c",
								Value: 2,
							},
						},
					},
				},
			},
		},
		{
			name: "message",
			src: `
// package doc
package x1

// comment for m1
message m1 {
  test.opt1
  int32 f1 = 1;
  int64 f2 = 2
  // comment f3
  string f3 = 3 {
     fopt.v1
     // comment fopt.v2
     fopt.v2 = "xx"
  }
  test.opt2 = 1
  float f4 = 4;
  double f5 = 5;
}

`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					YTDoc: doc("// package doc"),
					Name:  "x1",
				},
				Messages: []*ast.YTMessage{
					{
						YTDoc: doc("// comment for m1"),
						Name:  "m1",
						Fields: []*ast.YTField{
							{
								No:   1,
								Name: "f1",
								Type: &ast.YTFieldType{
									YTBaseType: ast.BaseTypeInt32,
								},
							},
							{
								No:   2,
								Name: "f2",
								Type: &ast.YTFieldType{
									YTBaseType: ast.BaseTypeInt64,
								},
							},
							{
								No:   3,
								Name: "f3",
								Type: &ast.YTFieldType{
									YTBaseType: ast.BaseTypeString,
								},
								YTDoc: doc("// comment f3"),
								YTOptions: ast.YTOptions{
									Opts: []*ast.YTOption{
										{
											Key: "fopt.v1",
										},
										{
											Key:   "fopt.v2",
											Value: optString("xx"),
											YTDoc: doc("// comment fopt.v2"),
										},
									},
								},
							},
							{
								No:   4,
								Name: "f4",
								Type: &ast.YTFieldType{
									YTBaseType: ast.BaseTypeFloat32,
								},
							},
							{
								No:   5,
								Name: "f5",
								Type: &ast.YTFieldType{
									YTBaseType: ast.BaseTypeFloat64,
								},
							},
						},
						YTOptions: ast.YTOptions{
							Opts: []*ast.YTOption{
								{
									Key: "test.opt1",
								},
								{
									Key:   "test.opt2",
									Value: optInt(1),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "message empty",
			src:  `package x1;message m1 {};message m2 {};`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				Messages: []*ast.YTMessage{
					{
						Name: "m1",
					},
					{
						Name: "m2",
					},
				},
			},
		},
		{
			name: "message only options",
			src:  `package x1;message m1 {test.opt1}`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				Messages: []*ast.YTMessage{
					{
						Name: "m1",
						YTOptions: ast.YTOptions{
							Opts: []*ast.YTOption{
								{
									Key: "test.opt1",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "message only options 2",
			src:  `package x1;message m1 {test.opt1}; message m2{}`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				Messages: []*ast.YTMessage{
					{
						Name: "m1",
						YTOptions: ast.YTOptions{
							Opts: []*ast.YTOption{
								{
									Key: "test.opt1",
								},
							},
						},
					},
					{
						Name: "m2",
					},
				},
			},
		},
		{
			name: "message only field",
			src:  `package x1;message m1 {int32 f1 = 1;}; message m2{}`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				Messages: []*ast.YTMessage{
					{
						Name: "m1",
						Fields: []*ast.YTField{
							{
								No:   1,
								Name: "f1",
								Type: &ast.YTFieldType{
									YTBaseType: ast.BaseTypeInt32,
								},
							},
						},
					},
					{
						Name: "m2",
					},
				},
			},
		},
		{
			name: "service",
			src: `package x1;message m1 {int32 f1 = 1;}; message m2{};
service s1
{
  set(m1);
  get() m1;
  mul(m1) m2{
    opt.m1
    opt.m2 = 2
  };
  opt.xx = "1"
}
`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				Messages: []*ast.YTMessage{
					{
						Name: "m1",
						Fields: []*ast.YTField{
							{
								No:   1,
								Name: "f1",
								Type: &ast.YTFieldType{
									YTBaseType: ast.BaseTypeInt32,
								},
							},
						},
					},
					{
						Name: "m2",
					},
				},
				Services: []*ast.YTService{
					{
						Name: "s1",
						YTOptions: ast.YTOptions{
							Opts: []*ast.YTOption{
								{
									Key:   "opt.xx",
									Value: optString("1"),
								},
							},
						},
						Methods: []*ast.YTMethod{
							{
								Name: "set",
								Request: &ast.YTMessage{
									Name: "m1",
									Fields: []*ast.YTField{
										{
											No:   1,
											Name: "f1",
											Type: &ast.YTFieldType{
												YTBaseType: ast.BaseTypeInt32,
											},
										},
									},
								},
							},
							{
								Name: "get",
								Reply: &ast.YTMessage{
									Name: "m1",
									Fields: []*ast.YTField{
										{
											No:   1,
											Name: "f1",
											Type: &ast.YTFieldType{
												YTBaseType: ast.BaseTypeInt32,
											},
										},
									},
								},
							},
							{
								Name: "mul",
								Request: &ast.YTMessage{
									Name: "m1",
									Fields: []*ast.YTField{
										{
											No:   1,
											Name: "f1",
											Type: &ast.YTFieldType{
												YTBaseType: ast.BaseTypeInt32,
											},
										},
									},
								},
								Reply: &ast.YTMessage{
									Name: "m2",
								},
								YTOptions: ast.YTOptions{
									Opts: []*ast.YTOption{
										{
											Key: "opt.m1",
										},
										{
											Key:   "opt.m2",
											Value: optInt(2),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "project",
			src: `package x1;
project m1 {
test.opt1
xx:
  vcf.ff = "xx"
}`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				Projects: []*ast.YTProject{
					{
						Name: "m1",
						Conf: map[string]*ast.YTOptions{
							"": {
								Opts: []*ast.YTOption{
									{
										Key: "test.opt1",
									},
								},
							},
							"xx": {
								Opts: []*ast.YTOption{
									{
										Key:   "vcf.ff",
										Value: optString("xx"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "protobuf fields",
			src:  `package x1;message m1{repeated int32 v1 = 1; map<int32,string> v2 = 2;}`,
			data: &ast.YTProgram{
				Pkg: &ast.YTPackage{
					Name: "x1",
				},
				Messages: []*ast.YTMessage{
					{
						Name: "m1",
						Fields: []*ast.YTField{
							{
								Type: &ast.YTFieldType{
									YTListType: &ast.YTListType{
										YTBaseType: ast.BaseTypeInt32,
									},
								},
								No:   1,
								Name: "v1",
							},
							{
								Type: &ast.YTFieldType{
									YTMapTypee: &ast.YTMapTypee{
										Key: ast.BaseTypeInt32,
										Value: &ast.YTListType{
											YTBaseType: ast.BaseTypeString,
										},
									},
								},
								No:   2,
								Name: "v2",
							},
						},
					},
				},
			},
		},
	}

	for _, v := range data {
		if len(testOnly) > 0 && v.name != testOnly {
			continue
		}
		t.Run(fmt.Sprintf("parse %s", v.name), func(t *testing.T) {
			prog, err := parseWProtoSrc(v.src)
			if v.err {
				t.Logf("parse %s error: %v\n", v.name, err)
				assert.NotNil(t, err, "it should be error")
				return
			}
			assert.Nil(t, err, "unexcepted parser error")
			if prog == nil {
				utils.Dump(err)
				return
			}
			if v.data.Pkg != nil || prog.Pkg != nil {
				assert.Equal(t, v.data.Pkg, prog.Pkg, "compare package")
			}
			if len(v.data.YTOptions.Opts) > 0 || len(prog.YTOptions.Opts) > 0 {
				assertValue(t, v.data.YTOptions.Opts, prog.YTOptions.Opts, "compare file options")
			}
			if len(v.data.Imports) > 0 || len(prog.Imports) > 0 {
				// 测试不手写import包,太复杂
				for _, v := range prog.Imports {
					v.Prog = nil
				}
				assertValue(t, v.data.Imports, prog.Imports, "compare import")
			}
			if len(v.data.EnumDefs) > 0 || len(prog.EnumDefs) > 0 {
				assertValue(t, v.data.EnumDefs, prog.EnumDefs, "compare enum")
			}
			if len(v.data.Messages) > 0 || len(prog.Messages) > 0 {
				assertValue(t, v.data.Messages, prog.Messages, "compare messages")
			}
			if len(v.data.Services) > 0 || len(prog.Services) > 0 {
				assertValue(t, v.data.Services, prog.Services, "compare service")
			}
			if len(v.data.Projects) > 0 || len(prog.Projects) > 0 {
				assertValue(t, v.data.Projects, prog.Projects, "compare service")
			}
			return
		})
	}
}

func assertValue(t testing.TB, except, value interface{}, msgAndArgs ...interface{}) {
	j1, _ := json.Marshal(except)
	j2, _ := json.Marshal(value)
	assert.JSONEq(t, string(j1), string(j2), msgAndArgs...)
}
