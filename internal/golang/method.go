package golang

import (
	"fmt"
	"go/ast"

	"github.com/artarts36/gostub/internal/ds"
)

type GoMethod struct {
	Name         ds.String
	Parameters   *GoParameters
	Results      *GoParameters
	UsedPackages *ds.Set[string]
}

func ParseMethodFromField(method *ast.Field) (*GoMethod, error) {
	goMethod := &GoMethod{
		Name:         ds.NewString(method.Names[0].Name),
		UsedPackages: ds.NewSet[string](),
	}

	mFunc, mFuncOk := method.Type.(*ast.FuncType)
	if !mFuncOk {
		return nil, fmt.Errorf("got invalid type for method %q", goMethod.Name.String())
	}

	goMethod.Parameters = &GoParameters{
		List:         make([]GoParameter, 0, len(mFunc.Params.List)),
		UsedPackages: ds.NewSet[string](),
	}

	goMethod.Results = &GoParameters{
		List:         make([]GoParameter, 0, len(mFunc.Results.List)),
		UsedPackages: ds.NewSet[string](),
	}

	for _, param := range mFunc.Params.List {
		goParam := GoParameter{
			Name: param.Names[0].Name,
		}

		goParamType, paramErr := parseParameterType(param.Type)
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
		goMethod.Parameters.UsedPackages.Merge(goParam.Type.UsedPackages)
	}

	if mFunc.Results != nil {
		for i, resultNode := range mFunc.Results.List {
			result := GoParameter{}

			if resultNode.Names != nil {
				result.Name = resultNode.Names[0].Name
			}

			paramType, paramTypeErr := parseParameterType(resultNode.Type)
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
			goMethod.Results.UsedPackages.Merge(paramType.UsedPackages)
		}
	}

	goMethod.UsedPackages.Merge(goMethod.Parameters.UsedPackages)
	goMethod.UsedPackages.Merge(goMethod.Results.UsedPackages)

	return goMethod, nil
}
