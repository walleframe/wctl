package protobuf

import (
	"encoding/json"
	"testing"

	"github.com/aggronmagi/wctl/builder/buildpb"
	"github.com/aggronmagi/wctl/protocol/ast"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	datas := []struct {
		name string
		data []byte
		err  bool
		prog *ast.YTProgram
		dst  *buildpb.FileDesc
	}{
		{
			name: "basic",
			data: []byte(`syntax = "proto3";

// package doc
package test // package tail doc


// import doc 1
import "abc" // import tail doc
// import doc 2
import abc2 "abc2";


option go_package = "xxx.xx/xx/xx";

// enum doc
enum e1 // enum tail doc 
{
    // enum value doc 
    v1 = 1; // enum value tail doc

    v2 = 0x03 // hex enum value
}

// xx
// xxx

// message doc
message m1 // message tail doc
{
    // field doc
    int32 f1 = 1; // field tail doc 

    // field 2 doc 
    int64 f2 = 2;
	repeated int32 f3 = 4;
}

message m2 {
	m1 f1 = 1;
	message m3 {
		int32 f1 = 1;
	}
	abc.abc f2 = 2;
	m3 f3 = 3;
}

// xxx xlxlxj
`),
			dst: &buildpb.FileDesc{
				Pkg: &buildpb.PackageDesc{
					Package: "test",
					Doc: &buildpb.DocDesc{
						Doc:     []string{"// package doc\n"},
						TailDoc: "// package tail doc\n",
					},
				},
				Options: &buildpb.OptionDesc{
					Options: map[string]*buildpb.OptionValue{
						"proto.gopkg": {
							Value: "xxx.xx/xx/xx",
						},
						"proto.syntax": {
							IntValue: 3,
						},
					},
				},
				Imports: []*buildpb.ImportDesc{
					{
						Doc: &buildpb.DocDesc{
							Doc:     []string{"// import doc 1\n"},
							TailDoc: "// import tail doc\n",
						},
						Alias: "",
						File:  "abc",
					},
					{
						Doc: &buildpb.DocDesc{
							Doc: []string{"// import doc 2\n"},
						},
						Alias: "abc2",
						File:  "abc2",
					},
				},
				Enums: []*buildpb.EnumDesc{
					{
						Name: "e1",
						Doc: &buildpb.DocDesc{
							Doc:     []string{"// enum doc\n"},
							TailDoc: "// enum tail doc \n",
						},
						Options: &buildpb.OptionDesc{},
						Values: []*buildpb.EnumValue{
							{
								Name: "v1",
								Doc: &buildpb.DocDesc{
									Doc:     []string{"// enum value doc \n"},
									TailDoc: "// enum value tail doc\n",
								},
								Value: 1,
							},
							{
								Name: "v2",
								Doc: &buildpb.DocDesc{
									TailDoc: "// hex enum value\n",
								},
								Value: 3,
							},
						},
					},
				},
				Msgs: []*buildpb.MsgDesc{
					{
						Name: "m1",
						Doc: &buildpb.DocDesc{
							Doc:     []string{"// message doc\n"},
							TailDoc: "// message tail doc\n",
						},
						Options: &buildpb.OptionDesc{},
						Fields: []*buildpb.Field{
							{
								Name: "f1",
								Doc: &buildpb.DocDesc{
									Doc:     []string{"// field doc\n"},
									TailDoc: "// field tail doc \n",
								},
								Options: &buildpb.OptionDesc{},
								No:      1,
								Type: &buildpb.TypeDesc{
									Type:    buildpb.FieldType_BaseType,
									Key:     "int32",
									KeyBase: buildpb.BaseTypeDesc_Int32,
								},
							},
							{
								Name: "f2",
								Doc: &buildpb.DocDesc{
									Doc: []string{"// field 2 doc \n"},
								},
								Options: &buildpb.OptionDesc{},
								No: 2,
								Type: &buildpb.TypeDesc{
									Type:    buildpb.FieldType_BaseType,
									Key:     "int64",
									KeyBase: buildpb.BaseTypeDesc_Int64,
								},
							},
							{
								Name: "f3",
								No:   4,
								Type: &buildpb.TypeDesc{
									Type:    buildpb.FieldType_ListType,
									Key:     "int32",
									KeyBase: buildpb.BaseTypeDesc_Int32,
								},
								Options: &buildpb.OptionDesc{},
							},
						},
						SubMsgs: []*buildpb.MsgDesc{},
					},
					{
						Name:    "m2",
						Options: &buildpb.OptionDesc{},
						Fields: []*buildpb.Field{
							{
								Name:    "f1",
								Options: &buildpb.OptionDesc{},
								No:      1,
								Type: &buildpb.TypeDesc{
									Type:       buildpb.FieldType_CustomType,
									Key:        "m1",
									ElemCustom: true,
								},
							},
							{
								Name:    "f2",
								Options: &buildpb.OptionDesc{},
								No:      2,
								Type: &buildpb.TypeDesc{
									Type:       buildpb.FieldType_CustomType,
									Key:        "abc.abc",
									ElemCustom: true,
								},
							},
							{
								Name:    "f3",
								Options: &buildpb.OptionDesc{},
								No:      3,
								Type: &buildpb.TypeDesc{
									Type:       buildpb.FieldType_CustomType,
									Key:        "m3",
									ElemCustom: true,
								},
							},
						},
						SubMsgs: []*buildpb.MsgDesc{
							{
								Name:    "m3",
								Options: &buildpb.OptionDesc{},
								Fields: []*buildpb.Field{
									{
										Name:    "f1",
										Options: &buildpb.OptionDesc{},
										No:      1,
										Type: &buildpb.TypeDesc{
											Type:    buildpb.FieldType_BaseType,
											Key:     "int32",
											KeyBase: buildpb.BaseTypeDesc_Int32,
										},
									},
								},
								SubMsgs: []*buildpb.MsgDesc{},
							},
						},
					},
				},
			},
		},
	}

	for _, data := range datas {
		t.Run(data.name, func(t *testing.T) {
			ret, err := Parse("test.wproto", data.data)
			if err != nil && !data.err {
				t.Fatal(err)
			}
			if !data.err && data.prog != nil {
				assert.EqualValues(t, data.prog, ret, "program")
			}
			if data.err {
				assert.NotNil(t, err, "error except")
				return
			}
			real, err := json.MarshalIndent(ret.GetFileDesc(), "", "  ")
			assert.Nil(t, err, "marshal rezult")
			except, err := json.MarshalIndent(data.dst, "", "  ")
			assert.Nil(t, err, "marshal except")
			assert.EqualValues(t, string(except), string(real), "rezult")
			// utils.Dump(ret.Pkg)
			// utils.Dump(ret.Imports)
			// utils.Dump(ret)
		})
	}
}
