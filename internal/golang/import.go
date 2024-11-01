package golang

import "fmt"

type GoImport struct {
	Alias string
	Path  string

	ShortName string
}

func (i *GoImport) String() string {
	if i.Alias == "" {
		return fmt.Sprintf("%q", i.Path)
	}

	return fmt.Sprintf("%s %q", i.Alias, i.Path)
}
