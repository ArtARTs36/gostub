package stub

import (
	"github.com/artarts36/goimports"
	"github.com/artarts36/gomodfinder"
	"github.com/artarts36/gostub/internal/golang"
)

type Stub struct {
	Filename string
	Package  *gomodfinder.Package
	Imports  *goimports.ImportGroups
	Types    []golang.Type

	GenMethods    bool
	GenTypes      bool
	MethodBodyTpl string
}
