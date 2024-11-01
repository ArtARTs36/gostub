package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/artarts36/gostub/internal/golang"
	"github.com/artarts36/gostub/internal/renderer"
	st "github.com/artarts36/gostub/internal/stub"
	"github.com/fatih/camelcase"
	"log/slog"
	"os"
	"strings"
	"text/template"
)

type Command struct {
	renderer      *renderer.Renderer
	stubCollector *st.Collector
}

type Params struct {
	Source string

	MethodBody string
	Package    string

	Filename string

	MethodPerFile     bool
	PerMethodFilename string

	TypePerFile     bool
	PerTypeFilename string

	TypeName string

	Out        string
	SkipExists bool

	Interfaces []string
}

func NewCommand(renderer *renderer.Renderer) *Command {
	return &Command{
		renderer: renderer,
	}
}

func (c *Command) Run(ctx context.Context, params *Params) error {
	slog.
		With(slog.Any("params", params)).
		InfoContext(ctx, "[command] running")

	nameGenerator, err := renderer.NewNameGenerator(
		params.Filename,
		params.PerMethodFilename,
		params.PerTypeFilename,
		params.TypeName,
	)
	if err != nil {
		return fmt.Errorf("failed to create name generator: %w", err)
	}

	src, err := os.ReadFile(params.Source)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", params.Source, err)
	}

	stubs, err := c.collectStubs2(src, params, nameGenerator)
	if err != nil {
		return fmt.Errorf("failed to collect stubs: %w", stubs)
	}

	return c.generate(ctx, stubs, params)
}

func (c *Command) generate(
	ctx context.Context,
	stubs []*st.Stub,
	params *Params,
) error {
	for _, stub := range stubs {
		filename := stub.Filename
		if params.Out != "" {
			filename = fmt.Sprintf("%s%s%s", params.Out, string(os.PathSeparator), filename)
		}

		if params.SkipExists {
			if _, err := os.Stat(filename); err == nil {
				continue
			}
		}

		slog.InfoContext(ctx, "[command] generating file", slog.String("file", filename))

		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("failed to open file %q: %w", stub.Filename, err)
		}

		err = c.renderer.Render(file, "stub.tpl", map[string]interface{}{
			"Stub": stub,
		})
		if err != nil {
			return fmt.Errorf("failed to render stub: %w", err)
		}
	}

	return nil
}

func (c *Command) collectStubs2(src []byte, params *Params, nameGenerator *renderer.NameGenerator) ([]*st.Stub, error) {
	methodBodyTpl := "method_body_nil_returns.tpl"
	if params.MethodBody == "panic" {
		methodBodyTpl = "method_body_panic.tpl"
	}

	interfaces, err := golang.ParseInterfacesFromSource(src, params.Interfaces)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go interfaces: %w", err)
	}

	return c.stubCollector.Collect(&st.CollectParams{
		GoInterfaces: interfaces,

		TypePerFile:   params.TypePerFile,
		MethodPerFile: params.MethodPerFile,

		MethodBodyTpl: methodBodyTpl,
	}, nameGenerator)
}

func (c *Command) collectStubs(src []byte, params *Params, nameGenerator *renderer.NameGenerator) ([]*Stub, error) {
	methodBodyTpl := "method_body_nil_returns.tpl"
	if params.MethodBody == "panic" {
		methodBodyTpl = "method_body_panic.tpl"
	}

	interfaces, err := golang.ParseInterfacesFromSource(src, params.Interfaces)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go interfaces: %w", err)
	}

	types, err := c.collectTypes(interfaces, nameGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to collect types: %w", err)
	}

	if !params.TypePerFile && !params.MethodPerFile {
		filename, err := nameGenerator.GenerateCommonFilename()
		if err != nil {
			return nil, err
		}

		return []*Stub{
			{
				Filename:      filename,
				Package:       types[0].Package,
				Imports:       types[0].Imports,
				Types:         types,
				GenMethods:    true,
				GenTypes:      true,
				MethodBodyTpl: methodBodyTpl,
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

			stub := &Stub{
				Filename: stubTypeFilename,
				Package:  typ.Package,
				Imports:  typ.Imports,
				Types: []golang.Type{
					typ,
				},
				GenTypes:      true,
				GenMethods:    !params.MethodPerFile,
				MethodBodyTpl: methodBodyTpl,
			}

			stubs = append(stubs, stub)
		}
	}

	if params.MethodPerFile {
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

				stub := &Stub{
					Filename: stubFilename,
					Package:  typ.Package,
					Imports:  imports,
					Types: []golang.Type{
						cType,
					},
					GenTypes:      false,
					GenMethods:    true,
					MethodBodyTpl: methodBodyTpl,
				}

				stubs = append(stubs, stub)
			}
		}
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

func (c *Command) collectTypes(
	interfaces []*golang.GoInterface,
	nameGenerator *renderer.NameGenerator,
) ([]golang.Type, error) {
	types := make([]golang.Type, 0, len(interfaces))

	for _, goInterface := range interfaces {
		nameWords := camelcase.Split(goInterface.Name.Value)

		typeName, err := nameGenerator.GenerateTypeName(goInterface)
		if err != nil {
			return nil, fmt.Errorf("failed to generate type name for interface %q: %w", goInterface.Name, err)
		}

		types = append(types, golang.Type{
			Name:      typeName,
			Imports:   goInterface.Imports,
			Package:   goInterface.Package,
			Receiver:  strings.ToLower(string(nameWords[len(nameWords)-1][0])),
			Methods:   goInterface.Methods,
			Interface: goInterface,
		})
	}

	return types, nil
}

func (c *Command) genFilename(tmpl *template.Template, params map[string]interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), err
}
