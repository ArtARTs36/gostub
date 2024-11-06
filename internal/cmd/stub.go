package cmd

import (
	"github.com/artarts36/goimports"
	"github.com/artarts36/gostub/internal/golang"
)

type Stub struct {
	Filename string
	Package  string
	Imports  *goimports.ImportGroups
	Types    []golang.Type

	GenMethods    bool
	GenTypes      bool
	MethodBodyTpl string
}
