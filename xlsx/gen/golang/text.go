package golang

const textTemplate = `// Generate by wctl xlsx. DO NOT EDIT.
package {{.PkgName}}

import (
	"fmt"
	"sync/atomic"
	
    "{{.ImportMgr}}"
    "{{.ImportPB}}"
)

const (
	ConfigName = "{{ToLower .StructName}}"
)

{{ if .IDTypes }}
{{- if eq .IDCnt 1 }}
type KeyType = {{TypeNameIndex .IDTypes 0}}
{{else}}
type KeyType struct { {{range $fi,$typ := $.IDTypes }}
    {{Comment $typ}}
	{{Title $typ.Name}} {{TypeName $typ}}  {{end}}
}
{{ end }}
{{ end }}

var (
	globalData    atomic.Pointer[tableData]
	checkHandlers []func(*{{.ProtoPkg}}.{{Title .StructName}}{{- if not .KVFlag -}}Container{{end}}) error
)

// Combine struct from {{.FromFile}} {{.SheetName}}
type tableData struct {
    // table {{.SheetName}}
	{{Title .StructName}} *{{.ProtoPkg}}.{{Title .StructName}}{{- if not .KVFlag -}}Container{{end}}{{ if .IDTypes }}
    MapData map[KeyType]*{{.ProtoPkg}}.{{Title .StructName}}{{- end }}
}

func init() {
	globalData.Store(new(tableData))
	{{.MgrPkg}}.RegAutoConfig("{{ToLower .StructName}}","{{.FromFile}}","{{.SheetName}}", nil,  &loader{})
}

type loader struct{}

// NewContainer 新建container 返回数据指针
func (l *loader) NewContainer() interface{} {
	return new({{.ProtoPkg}}.{{Title .StructName}}{{- if not .KVFlag -}}Container{{end}})
}

// Check 检查数据
func (l *loader) Check(new interface{}) error {
	if new == nil {
		return fmt.Errorf("ptr is nil")
	}
	t, ok := new.(*{{.ProtoPkg}}.{{Title .StructName}}{{- if not .KVFlag -}}Container{{end}})
	if !ok {
		return fmt.Errorf("error type")
	}

	for _, c := range checkHandlers {
		if err := c(t); err != nil {
			return err
		}
	}
	return nil
}

// Swap 交换内存地址
func (l *loader) Swap(new interface{}) {
	newT, ok := new.(*{{.ProtoPkg}}.{{Title .StructName}}{{- if not .KVFlag -}}Container{{end}})
	if !ok {
		return
	}
	newL := &tableData{
		{{Title .StructName}}: newT,
{{- if .IDTypes -}}
        MapData: make(map[KeyType]*{{.ProtoPkg}}.{{Title .StructName}}, len(newT.Data)),{{- end }}
	}
{{ if .IDTypes }}
{{- if eq .IDCnt 1 }}{{ $typ := index .IDTypes 0}}
    for _, item := range newT.Data {
        newL.MapData[item.{{Title  $typ.Name}}] = item
    }
{{else}}
    for _, item := range newT.Data {
        id := KeyType{ {{range $fi,$typ := $.IDTypes }}
            {{Title  $typ.Name}}: item.{{Title  $typ.Name}},{{end}}
        }
        newL.MapData[id] = item
    } 
{{ end }}
{{ end }}
	globalData.Store(newL)
}

// RegisterCheckEntry 注册校验回调(用于更新数据前校验)
func RegisterCheckEntry(h func(*{{.ProtoPkg}}.{{Title .StructName}}{{- if not .KVFlag -}}Container{{end}}) error) {

	if h == nil {
		panic("empty preload handler")
	}

	checkHandlers = append(checkHandlers, h)
}

{{- if .KVFlag }}

// Get() 获取数据 from {{.FromFile}} {{.SheetName}}
func Get() *{{.ProtoPkg}}.{{Title .StructName}} {
    data := globalData.Load()
	if data == nil || data.{{Title .StructName}} == nil {
		return nil
	}
	return data.{{Title .StructName}}
}

{{- else }}

// Get 获取全部数据 {{.FromFile}} {{.SheetName}}
func Get() []*{{.ProtoPkg}}.{{Title .StructName}} {
    data := globalData.Load()
	if data == nil || data.{{Title .StructName}} == nil {
		return nil
	}
	return data.{{Title .StructName}}.Data
}

{{ if .IDTypes }}
// GetByID 根据索引获取数据
func GetByID(id KeyType) *{{.ProtoPkg}}.{{Title .StructName}} {
    data := globalData.Load()
	if data == nil || data.MapData == nil {
		return nil
	}
	v,ok := data.MapData[id]
	if !ok {
		return nil
	}
	return v
}

{{- if gt .IDCnt 1 }}
type KeyType = {{TypeNameIndex .IDTypes 0}}
// GetByKey 根据索引获取数据
func GetByKey({ {{range $fi,$typ := $.IDTypes }}{{Title $typ.Name}} {{TypeName $typ}},  {{end}}) *{{.ProtoPkg}}.{{Title .StructName}} {
    data := globalData.Load()
	if data == nil || data.MapData == nil {
		return nil
	}
	v,ok := data.MapData[KeyType{ {{range $fi,$typ := $.IDTypes }}
		{{Title $typ.Name}}:{{Title $typ.Name}},  {{end}}
	}]
	if !ok {
		return nil
	}
	return v
}

{{ end }}
{{ end }}

// Count 获取配置总个数
func Count() int {
	data := globalData.Load()
	if data == nil || data.{{Title .StructName}} == nil {
		return 0
	}
	return len(data.{{Title .StructName}}.Data)
}

// Range 遍历
func Range(filter func(index int, val *{{.ProtoPkg}}.{{Title .StructName}}) bool) {
    data := globalData.Load()
	if data == nil || data.{{Title .StructName}} == nil {
		return
	}
	for index, v := range data.{{Title .StructName}}.Data {
		if !filter(index, v) {
			return
		}
	}
}

// GetByIndex 根据下标获取数据
func GetByIndex(index int) *{{.ProtoPkg}}.{{Title .StructName}} {
    data := globalData.Load()
	if data == nil || data.{{Title .StructName}} == nil {
		return nil
	}
	if index < 0 || index > len(data.{{Title .StructName}}.Data) {
		return nil
	}
	return data.{{Title .StructName}}.Data[index]
}

// GetByFilter 根据过滤器获取批量数据
func GetByFilter(filter func(val *{{.ProtoPkg}}.{{Title .StructName}}) bool) (ret []*{{.ProtoPkg}}.{{Title .StructName}}) {
	for _, v := range Get() {
		if filter(v) {
			ret = append(ret, v)
		}
	}
	return ret
}

// GetOneByFilter 根据过滤器获取单个数据
func GetOneByFilter(filter func(val *{{.ProtoPkg}}.{{Title .StructName}}) bool) *{{.ProtoPkg}}.{{Title .StructName}} {
	for _, v := range Get() {
		if filter(v) {
			return v
		}
	}
	return nil
}

{{end}}
`
