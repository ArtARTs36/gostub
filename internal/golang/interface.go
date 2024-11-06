package golang

import (
	"fmt"
	"github.com/artarts36/gds"
	"github.com/artarts36/goimports"
	"github.com/artarts36/gomodfinder"
	"github.com/artarts36/gostub/internal/ds"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

type GoInterface struct {
	Name    ds.String
	Imports *goimports.ImportGroups
	Package *gomodfinder.Package
	Methods []*GoMethod
}

type ParseInterfacesParams struct {
	Source         []byte
	SourcePath     string
	FilterNames    []string
	SourceGoModule *gomodfinder.ModFile
}

func ParseInterfacesFromSource(params ParseInterfacesParams) (*File, error) {
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "demo", params.Source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parsedFile: %w", err)
	}

	pkg := params.SourceGoModule.CalcPackageFromAbsPathWithName(parsedFile.Name.String(), filepath.Dir(params.SourcePath))

	imports := goimports.NewImportGroupsFromAstImportSpecs(parsedFile.Imports, params.SourceGoModule.File.Module.Mod.Path)

	needInterfacesSet := map[string]bool{}
	for _, needInterface := range params.FilterNames {
		needInterfacesSet[needInterface] = true
	}

	isNeed := func(interfaceName string) bool {
		if len(params.FilterNames) == 0 {
			return true
		}

		return needInterfacesSet[interfaceName]
	}

	file := &File{
		Imports:    imports,
		Interfaces: make([]*GoInterface, 0),
	}

	importsShortnameMap := gds.NewMap[string, goimports.GoImport]()
	for _, goImports := range imports.SortedImports() {
		for _, goImport := range goImports {
			if goImport.Alias != "" {
				importsShortnameMap.Set(goImport.Alias, goImport)
			} else {
				importsShortnameMap.Set(goImport.Package.LastName, goImport)
			}
		}
	}

	var inspectErr error

	ast.Inspect(parsedFile, func(x ast.Node) bool {
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
			Name:    ds.NewString(spec.Name.Name),
			Imports: goimports.NewImportGroups(params.SourceGoModule.Module.Mod.Path),
			Package: pkg,
			Methods: make([]*GoMethod, 0),
		}

		for _, method := range it.Methods.List {
			goMethod, goMethodErr := ParseMethodFromField(method, pkg, importsShortnameMap, params.SourceGoModule.Module.Mod.Path)
			if goMethodErr != nil {
				inspectErr = fmt.Errorf("failed to parse method for interface %q: %w", goInterface.Name, goMethodErr)
				return false
			}

			goInterface.Methods = append(goInterface.Methods, goMethod)
		}

		for _, method := range goInterface.Methods {
			for _, impGroups := range method.Imports.SortedImports() {
				for _, imp := range impGroups {
					goInterface.Imports.Add(imp.Alias, imp.Package.Path)
				}
			}
		}

		file.Interfaces = append(file.Interfaces, goInterface)

		return false
	})

	return file, inspectErr
}
