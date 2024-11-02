package golang

import (
	"errors"
	"fmt"
	"github.com/artarts36/gostub/internal/ds"
	"go/ast"
	"html/template"
)

type GoParameters struct {
	List                  []GoParameter
	HasValueThroughAnyArg bool
	UsedPackages          *ds.Set[string]
}

type GoParameter struct {
	Name string
	Type GoParameterType
}

type GoParameterType struct {
	Pointer bool
	Name    string

	UsedPackages *ds.Set[string]

	ValueThroughNil bool
	ValueThroughVar bool

	Value template.HTML
}

func (t *GoParameterType) IsNumeric() bool {
	return t.Name == "int" || t.Name == "int16" || t.Name == "int32" ||
		t.Name == "uint8" || t.Name == "uint16" || t.Name == "uint32" || t.Name == "uint64" ||
		t.Name == "float32" || t.Name == "float64"
}

func (t *GoParameterType) String() string {
	return t.Name
}

func (t *GoParameterType) calcStubInstantiateExpr(methodName string) {
	calc := func(methodName string) template.HTML {
		if t.ValueThroughNil {
			return "nil"
		}

		if !t.UsedPackages.Valid() {
			if t.Name == "error" {
				return template.HTML(fmt.Sprintf(`fmt.Errorf("is not real method %s")`, methodName))
			} else if t.Name == "any" || t.Name == "interface" {
				return "nil"
			} else if t.Name == "string" {
				return `""`
			} else if t.Name == "bool" {
				return "false"
			} else if t.IsNumeric() {
				return "0"
			}
		}

		t.ValueThroughVar = true

		return template.HTML(fmt.Sprintf(
			"anyArg.(%s)",
			t.String(),
		))
	}

	t.Value = calc(methodName)
}

func parseParameterType(ptNode ast.Node) (GoParameterType, error) {
	result := GoParameterType{
		Name:         "",
		UsedPackages: ds.NewSet[string](),
	}

	if _, ok := ptNode.(*ast.StarExpr); ok {
		result.Pointer = true
		result.ValueThroughNil = true
	} else if at, ok := ptNode.(*ast.ArrayType); ok {
		if at.Len == nil {
			result.ValueThroughNil = true
		}
	} else if _, ok := ptNode.(*ast.InterfaceType); ok {
		result.ValueThroughNil = true
	}

	var parse func(node ast.Node) (string, error)

	parse = func(node ast.Node) (string, error) {
		switch pt := node.(type) {
		case *ast.Ident:
			return pt.Name, nil
		case *ast.SelectorExpr:
			packageNameIdent, ok := pt.X.(*ast.Ident)
			if ok {
				result.UsedPackages.Add(packageNameIdent.Name)

				return fmt.Sprintf("%s.%s", packageNameIdent.Name, pt.Sel.Name), nil
			}
			return pt.Sel.Name, nil
		case *ast.StarExpr:
			n, err := parse(pt.X)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("*%s", n), nil
		case *ast.ArrayType:
			el, err := parse(pt.Elt)
			if err != nil {
				return "", err
			}

			length := ""
			if pt.Len != nil {
				length, err = parse(pt.Len)
				if err != nil {
					return "", fmt.Errorf("failed to parse array length: %w", err)
				}
			}

			return fmt.Sprintf("[%s]%s", length, el), nil
		case *ast.MapType:
			key, err := parse(pt.Key)
			if err != nil {
				return "", fmt.Errorf("failed to parse map key: %w", err)
			}

			value, err := parse(pt.Value)
			if err != nil {
				return "", fmt.Errorf("failed to parse map value: %w", err)
			}

			return fmt.Sprintf("map[%s]%s", key, value), nil
		case *ast.BasicLit:
			return pt.Value, nil
		case *ast.FuncType:
			return "", errors.New("func in parameters/results not supported. You can use aliases")
		case *ast.InterfaceType:
			return "interface{}", nil
		}

		return "", fmt.Errorf("unknown node %T", node)
	}

	var err error
	result.Name, err = parse(ptNode)
	if err != nil {
		return result, err
	}

	return result, nil
}
