package golang

type Package struct {
	Name                string
	FullName            string
	ProjectRelativePath string

	Module string
}

type UsedPackage struct {
	Alias   string
	Package Package
}
