package pb

const textTemplate = `
// generate by retool. DO NOT EDIT.
syntax = "proto3";
package {{.PackageName}};
option go_package = "{{.PackageName}}";

// {{.StructName}} generate from {{.SheetName}} in {{.FromFile}}
message {{.StructName}}
{ {{range $fi,$typ := $.AllType }}{{if EnableExport $typ }}
    {{Comment $typ}}
	{{TypeName $typ}} {{$typ.Name}} {{PBTag $fi}}; {{end}}{{end}}
}

{{- if not .KVFlag }}
// {{.StructName}} generate from {{.SheetName}} in {{.FromFile}}
message {{.StructName}}Container
{	
	repeated {{.StructName}} Data = {{1}}; // table 
}
{{ end -}}
`
