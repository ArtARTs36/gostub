package golang

import (
	"fmt"
	"github.com/artarts36/gostub/internal/ds"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type GoInterface struct {
	Name         ds.String
	Imports      []GoImport
	Package      string
	Methods      []*GoMethod
	UsedPackages *ds.Set[string]
}

func ParseInterfacesFromSource(src []byte, needInterfaces []string) ([]*GoInterface, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "demo", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	pkg := file.Name.String()

	imports := make([]GoImport, 0)

	needInterfacesSet := map[string]bool{}
	for _, needInterface := range needInterfaces {
		needInterfacesSet[needInterface] = true
	}

	isNeed := func(interfaceName string) bool {
		if len(needInterfaces) == 0 {
			return true
		}

		return needInterfacesSet[interfaceName]
	}

	for _, spec := range file.Imports {
		imp := GoImport{
			Path: strings.Trim(spec.Path.Value, `"`),
		}

		pathParts := strings.Split(imp.Path, "/")
		imp.ShortName = pathParts[len(pathParts)-1]

		if spec.Name != nil {
			imp.Alias = spec.Name.Name
		}

		imports = append(imports, imp)
	}

	interfaces := make([]*GoInterface, 0)

	var inspectErr error

	ast.Inspect(file, func(x ast.Node) bool {
		spec, ok := x.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if !isNeed(spec.Name.Name) {
			return true
		}

		it, ok := spec.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}

		goInterface := &GoInterface{
			Name:         ds.NewString(spec.Name.Name),
			Imports:      imports,
			Package:      pkg,
			Methods:      make([]*GoMethod, 0),
			UsedPackages: ds.NewSet[string](),
		}

		for _, method := range it.Methods.List {
			goMethod, goMethodErr := ParseMethodFromField(method)
			if goMethodErr != nil {
				inspectErr = fmt.Errorf("failed to parse method for interface %q: %w", goInterface.Name, goMethodErr)
				return false
			}

			goInterface.Methods = append(goInterface.Methods, goMethod)
			goInterface.UsedPackages.Merge(goMethod.UsedPackages)
		}

		interfaces = append(interfaces, goInterface)

		return false
	})

	return interfaces, inspectErr
}
