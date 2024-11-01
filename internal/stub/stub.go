package stub

import "github.com/artarts36/gostub/internal/golang"

type Stub struct {
	Filename string
	Package  string
	Imports  []golang.GoImport
	Types    []golang.Type

	GenMethods    bool
	GenTypes      bool
	MethodBodyTpl string
}
