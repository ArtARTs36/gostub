package golang

type Type struct {
	Name     string
	Imports  []GoImport
	Package  string
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
