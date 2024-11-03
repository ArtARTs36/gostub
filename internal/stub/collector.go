package stub

import (
	"fmt"
	"github.com/artarts36/gomodfinder"
	"strings"

	"github.com/fatih/camelcase"

	"github.com/artarts36/gostub/internal/golang"
	"github.com/artarts36/gostub/internal/renderer"
)

type Collector struct {
}

type CollectParams struct {
	GoInterfaces []*golang.GoInterface

	TypePerFile   bool
	MethodPerFile bool

	MethodBodyTpl string

	TargetPackage *gomodfinder.Package
}

func (c *Collector) Collect(params *CollectParams, nameGenerator *renderer.NameGenerator) ([]*Stub, error) {
	types, err := c.collectTypes(params, nameGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to collect types: %w", err)
	}

	if !params.TypePerFile && !params.MethodPerFile {
		filename, err := nameGenerator.GenerateCommonFilename()
		if err != nil {
			return nil, err
		}

		pkg := params.TargetPackage
		if pkg == nil {
			pkg = types[0].Package
		}

		return []*Stub{
			{
				Filename:      filename,
				Package:       pkg,
				Imports:       types[0].Imports,
				Types:         types,
				GenMethods:    true,
				GenTypes:      true,
				MethodBodyTpl: params.MethodBodyTpl,
			},
		}, nil
	}

	stubs := make([]*Stub, 0)

	if params.TypePerFile {
		for _, typ := range types {
			stubTypeFilename, err := nameGenerator.GenerateStubStructFilename(typ.Interface)
			if err != nil {
				return nil, err
			}

			imports := make([]golang.GoImport, 0)
			if !params.MethodPerFile && typ.Interface.UsedPackages.Valid() {
				importsMap := map[string]golang.GoImport{}
				for _, goImport := range typ.Imports {
					if goImport.Alias != "" {
						importsMap[goImport.Alias] = goImport
					}

					if goImport.ShortName != "" {
						importsMap[goImport.ShortName] = goImport
					}
				}

				for _, usedPackage := range typ.Interface.UsedPackages.List() {
					if imp, ok := importsMap[usedPackage]; ok {
						imports = append(imports, imp)
					}
				}
			}

			stub := &Stub{
				Filename: stubTypeFilename,
				Package:  typ.Package,
				Imports:  imports,
				Types: []golang.Type{
					typ,
				},
				GenTypes:      true,
				GenMethods:    !params.MethodPerFile,
				MethodBodyTpl: params.MethodBodyTpl,
			}

			stubs = append(stubs, stub)
		}
	}

	if params.MethodPerFile {
		methodStubs, err := c.collectMethodStubs(types, params, nameGenerator)
		if err != nil {
			return nil, fmt.Errorf("failed to collect method stubs: %w", err)
		}

		stubs = append(stubs, methodStubs...)
	}

	if !params.TypePerFile && params.MethodPerFile {
		filename, err := nameGenerator.GenerateCommonFilename()
		if err != nil {
			return nil, err
		}

		stub := &Stub{
			Filename:   filename,
			Package:    types[0].Package,
			Types:      types,
			GenMethods: false,
			GenTypes:   true,
		}

		stubs = append(stubs, stub)
	}

	return stubs, nil
}

func (c *Collector) collectTypes(
	params *CollectParams,
	nameGenerator *renderer.NameGenerator,
) ([]golang.Type, error) {
	types := make([]golang.Type, 0, len(params.GoInterfaces))

	for _, goInterface := range params.GoInterfaces {
		nameWords := camelcase.Split(goInterface.Name.Value)

		typeName, err := nameGenerator.GenerateTypeName(goInterface)
		if err != nil {
			return nil, fmt.Errorf("failed to generate type name for interface %q: %w", goInterface.Name, err)
		}

		pkg := goInterface.Package
		if params.TargetPackage != nil {
			pkg = params.TargetPackage
		}

		types = append(types, golang.Type{
			Name:      typeName,
			Imports:   goInterface.Imports,
			Package:   pkg,
			Receiver:  strings.ToLower(string(nameWords[len(nameWords)-1][0])),
			Methods:   goInterface.Methods,
			Interface: goInterface,
		})
	}

	return types, nil
}

func (c *Collector) collectMethodStubs(
	types []golang.Type,
	params *CollectParams,
	nameGenerator *renderer.NameGenerator,
) ([]*Stub, error) {
	stubs := make([]*Stub, 0)

	for _, typ := range types {
		for _, method := range typ.Methods {
			stubFilename, err := nameGenerator.GenerateStubPerMethodFilename(typ, method)
			if err != nil {
				return nil, err
			}

			cType := typ.Clone()
			cType.Methods = []*golang.GoMethod{
				method,
			}

			imports := make([]golang.GoImport, 0)
			if method.UsedPackages.Valid() {
				importsMap := map[string]golang.GoImport{}
				for _, goImport := range typ.Imports {
					if goImport.Alias != "" {
						importsMap[goImport.Alias] = goImport
					}

					if goImport.ShortName != "" {
						importsMap[goImport.ShortName] = goImport
					}
				}

				for _, usedPackage := range method.UsedPackages.List() {
					if imp, ok := importsMap[usedPackage]; ok {
						imports = append(imports, imp)
					}
				}
			}

			pkg := typ.Package
			if params.TargetPackage != nil {
				pkg = params.TargetPackage
				imports = append(imports, golang.GoImport{
					Path: fmt.Sprintf("%s/%s", typ.Interface.Package.Module.Module.Mod.Path, typ.Interface.Package.ModuleRelativePath),
				})
			}

			stub := &Stub{
				Filename: stubFilename,
				Package:  pkg,
				Imports:  imports,
				Types: []golang.Type{
					cType,
				},
				GenTypes:      false,
				GenMethods:    true,
				MethodBodyTpl: params.MethodBodyTpl,
			}

			stubs = append(stubs, stub)
		}
	}

	return stubs, nil
}
