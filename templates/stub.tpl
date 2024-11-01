{{ $methodBodyTpl := .Stub.MethodBodyTpl }}{{ $types := .Stub.Types }}package {{ .Stub.Package }}{{ if noEmpty .Stub.Imports }}

import ({{ $imports := .Stub.Imports }}{{ range $importIndex, $import := .Stub.Imports }}
    {{ raw $import.String }}{{ if (isLast $importIndex $imports) }}
{{ end }}{{ end }}){{ end }}{{ if .Stub.GenTypes }}
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
