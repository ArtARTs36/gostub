package golang

import (
	"fmt"
	"github.com/artarts36/gds"
	"github.com/artarts36/goimports"
	"github.com/artarts36/gomodfinder"
	"go/ast"

	"github.com/artarts36/gostub/internal/ds"
)

type GoMethod struct {
	Name         ds.String
	Parameters   *GoParameters
	Results      *GoParameters
	UsedPackages *ds.Set[string]
	Imports      *goimports.ImportGroups
}

func ParseMethodFromField(
	method *ast.Field,
	pkg *gomodfinder.Package,
	imports *gds.Map[string, goimports.GoImport],
	goModule string,
) (*GoMethod, error) {
	goMethod := &GoMethod{
		Name:         ds.NewString(method.Names[0].Name),
		UsedPackages: ds.NewSet[string](),
		Imports:      goimports.NewImportGroups(goModule),
		Parameters:   &GoParameters{List: make([]GoParameter, 0)},
		Results:      &GoParameters{List: make([]GoParameter, 0)},
	}

	mFunc, mFuncOk := method.Type.(*ast.FuncType)
	if !mFuncOk {
		return nil, fmt.Errorf("got invalid type for method %q", goMethod.Name.String())
	}

	if mFunc.Params != nil {
		goMethod.Parameters = &GoParameters{
			List: make([]GoParameter, 0, len(mFunc.Params.List)),
		}

		for _, param := range mFunc.Params.List {
			goParam := GoParameter{
				Name: param.Names[0].Name,
			}

			goParamType, paramErr := parseParameterType(param.Type, pkg)
			if paramErr != nil {
				return nil, fmt.Errorf(
					"failed to get type name for %s.%s: %w",
					goMethod.Name.String(),
					goParam.Name,
					paramErr,
				)
			}

			if !goParamType.ValueThroughNil {
				goMethod.Parameters.HasValueThroughAnyArg = true
			}

			goParam.Type = goParamType

			goMethod.Parameters.List = append(goMethod.Parameters.List, goParam)

			for _, pkgName := range goParam.Type.UsedPackages.List() {
				imp, ok := imports.Get(pkgName)
				if ok {
					goMethod.Imports.Add(imp.Alias, imp.Package.Path)
				}
			}
		}
	}

	if mFunc.Results != nil {
		goMethod.Results = &GoParameters{
			List: make([]GoParameter, 0, len(mFunc.Results.List)),
		}

		for i, resultNode := range mFunc.Results.List {
			result := GoParameter{}

			if resultNode.Names != nil {
				result.Name = resultNode.Names[0].Name
			}

			paramType, paramTypeErr := parseParameterType(resultNode.Type, pkg)
			if paramTypeErr != nil {
				return nil, fmt.Errorf(
					"failed to parse result[%d] type for method %q: %w",
					i,
					goMethod.Name.String(),
					paramTypeErr,
				)
			}

			paramType.calcStubInstantiateExpr(goMethod.Name.String())

			if paramType.ValueThroughVar {
				goMethod.Results.HasValueThroughAnyArg = true
			}

			result.Type = paramType

			goMethod.Results.List = append(goMethod.Results.List, result)

			for _, pkgName := range paramType.UsedPackages.List() {
				imp, ok := imports.Get(pkgName)
				if ok {
					goMethod.Imports.Add(imp.Alias, imp.Package.Path)
				}
			}
		}
	}

	return goMethod, nil
}
