package golang

import (
	"github.com/artarts36/goimports"
	"github.com/artarts36/gomodfinder"
)

type Type struct {
	Name     string
	Imports  *goimports.ImportGroups
	Package  *gomodfinder.Package
	Receiver string
	Methods  []*GoMethod

	Interface *GoInterface
}

func (t *Type) Clone() Type {
	return Type{
		Name:     t.Name,
		Imports:  t.Imports,
		Package:  t.Package,
		Receiver: t.Receiver,
		Methods:  t.Methods,

		Interface: t.Interface,
	}
}
