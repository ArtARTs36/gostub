package golang

import "github.com/artarts36/goimports"

type File struct {
	Imports    *goimports.ImportGroups
	Interfaces []*GoInterface
}
