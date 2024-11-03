package stub

import (
	"github.com/artarts36/gomodfinder"
	"github.com/artarts36/gostub/internal/golang"
)

type Stub struct {
	Filename string
	Package  *gomodfinder.Package
	Imports  []golang.GoImport
	Types    []golang.Type

	GenMethods    bool
	GenTypes      bool
	MethodBodyTpl string
}
