{{ $methodBodyTpl := .Stub.MethodBodyTpl }}{{ $types := .Stub.Types }}package {{ .Stub.Package.Name }}{{ if noEmpty .Stub.Imports }}

import ({{ $imports := .Stub.Imports.SortedImports }}{{ range $importGroupIndex, $importGroup := $imports }}{{ range $importIndex, $import := $importGroup }}
    {{ raw $import.GoString }}{{ if (isLast $importIndex $importGroup) }}
{{ end }}{{ end }}{{ end }}){{ end }}{{ if .Stub.GenTypes }}
{{ range $typIndex, $typ := $types }}
type {{ .Name }} struct {

}{{ if (hasNext $typIndex $types) }}
{{ end }}{{ end }}
{{ range $typIndex, $typ := .Stub.Types }}
func New{{ .Name }}() *{{ .Name }} {
    return &{{ .Name }}{}
}{{ if hasNext $typIndex $types }}
{{ end }}{{ end }}{{ end }}{{ if .Stub.GenMethods }}
{{ range $typIndex, $typ := .Stub.Types }}{{ $methods := $typ.Methods }}{{ range $index, $method := $methods }}
{{ include "method.tpl" "Type" $typ "Method" $method "MethodBodyTpl" $methodBodyTpl }}{{ if hasNext $index $methods }}
{{ end }}{{ end }}{{ if hasNext $typIndex $types }}
{{ end }}{{ end }}{{ end }}
