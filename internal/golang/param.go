package golang

import (
	"errors"
	"fmt"
	"github.com/artarts36/gomodfinder"
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
	Name         string
	ExternalName string

	UsedPackages *ds.Set[string]

	ValueThroughNil bool
	ValueThroughVar bool

	Value template.HTML

	Package *gomodfinder.Package
}

func (t *GoParameterType) Call(pkg *gomodfinder.Package) string {
	if t.Package.Equal(pkg) {
		return t.Name
	}

	return t.ExternalName
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
			t.UsedPackages.Add("errors")

			switch t.Name {
			case TypeError:
				return template.HTML(fmt.Sprintf(`errors.New("is not real method %s")`, methodName))
			case TypeAny, "interface":
				return "nil"
			case TypeString:
				return `""`
			case TypeBool:
				return "false"
			}

			if IsNumericType(t.Name) {
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

func calcPtVtn(result GoParameterType, ptNode ast.Node) {
	if _, ok := ptNode.(*ast.StarExpr); ok {
		result.ValueThroughNil = true
	} else if at, atOk := ptNode.(*ast.ArrayType); atOk {
		if at.Len == nil {
			result.ValueThroughNil = true
		}
	} else if _, itOk := ptNode.(*ast.InterfaceType); itOk {
		result.ValueThroughNil = true
	}
}

func parseParameterType(ptNode ast.Node, pkg *gomodfinder.Package) (GoParameterType, error) { //nolint:funlen,lll // not need
	result := GoParameterType{
		Name:         "",
		UsedPackages: ds.NewSet[string](),
		Package:      pkg,
	}

	calcPtVtn(result, ptNode)

	var parse func(node ast.Node) (string, string, error)

	parse = func(node ast.Node) (string, string, error) {
		switch pt := node.(type) {
		case *ast.Ident:
			if IsStdType(pt.Name) {
				return pt.Name, pt.Name, nil
			}

			return pt.Name, fmt.Sprintf("%s.%s", pkg.Name, pt.Name), nil
		case *ast.SelectorExpr:
			packageNameIdent, ok := pt.X.(*ast.Ident)
			if ok {
				result.UsedPackages.Add(packageNameIdent.Name)

				return fmt.Sprintf("%s.%s", packageNameIdent.Name, pt.Sel.Name),
					fmt.Sprintf("%s.%s", packageNameIdent.Name, pt.Sel.Name),
					nil
			}

			if IsStdType(pt.Sel.Name) {
				return pt.Sel.Name, pt.Sel.Name, nil
			}

			return pt.Sel.Name, fmt.Sprintf("%s.%s", pkg.Name, pt.Sel.Name), nil
		case *ast.StarExpr:
			n, extN, err := parse(pt.X)
			if err != nil {
				return "", "", err
			}

			return fmt.Sprintf("*%s", n), fmt.Sprintf("*%s", extN), nil
		case *ast.ArrayType:
			el, extEl, err := parse(pt.Elt)
			if err != nil {
				return "", "", err
			}

			length := ""
			if pt.Len != nil {
				length, _, err = parse(pt.Len)
				if err != nil {
					return "", "", fmt.Errorf("failed to parse array length: %w", err)
				}
			}

			return fmt.Sprintf("[%s]%s", length, el), fmt.Sprintf("[%s]%s", length, extEl), nil
		case *ast.MapType:
			key, extKey, err := parse(pt.Key)
			if err != nil {
				return "", "", fmt.Errorf("failed to parse map key: %w", err)
			}

			value, extVal, err := parse(pt.Value)
			if err != nil {
				return "", "", fmt.Errorf("failed to parse map value: %w", err)
			}

			return fmt.Sprintf("map[%s]%s", key, value), fmt.Sprintf("map[%s]%s", extKey, extVal), nil
		case *ast.BasicLit:
			return pt.Value, pt.Value, nil
		case *ast.FuncType:
			return "", "", errors.New("func in parameters/results not supported. You can use aliases")
		case *ast.InterfaceType:
			return "interface{}", "interface{}", nil
		}

		return "", "", fmt.Errorf("unknown node %T", node)
	}

	var err error
	result.Name, result.ExternalName, err = parse(ptNode)
	if err != nil {
		return result, err
	}

	return result, nil
}
