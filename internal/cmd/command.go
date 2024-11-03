package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/artarts36/gomodfinder"
	"github.com/artarts36/gostub/internal/golang"
	"github.com/artarts36/gostub/internal/renderer"
	st "github.com/artarts36/gostub/internal/stub"
	"log/slog"
	"os"
	"path/filepath"
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

	Interfaces     []string
	SourceGoModule *gomodfinder.ModFile
	TargetGoModule *gomodfinder.ModFile
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

	stubs, err := c.collectStubs(src, params, nameGenerator)
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

func (c *Command) collectStubs(src []byte, params *Params, nameGenerator *renderer.NameGenerator) ([]*st.Stub, error) {
	methodBodyTpl := "method_body_nil_returns.tpl"
	if params.MethodBody == "panic" {
		methodBodyTpl = "method_body_panic.tpl"
	}

	sourceAbsPath, err := filepath.Abs(params.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to get absoulte source path: %w", sourceAbsPath)
	}

	interfaces, err := golang.ParseInterfacesFromSource(golang.ParseInterfacesParams{
		Source:         src,
		SourcePath:     sourceAbsPath,
		FilterNames:    params.Interfaces,
		SourceGoModule: params.SourceGoModule,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse go interfaces: %w", err)
	}

	var targetPkg *gomodfinder.Package
	if params.Package != "" {
		targetPkg = params.TargetGoModule.Package(params.Package)
	} else {
		targetPkg = interfaces[0].Package
	}

	return c.stubCollector.Collect(&st.CollectParams{
		GoInterfaces: interfaces,

		TypePerFile:   params.TypePerFile,
		MethodPerFile: params.MethodPerFile,

		MethodBodyTpl: methodBodyTpl,

		TargetPackage: targetPkg,
	}, nameGenerator)
}

func (c *Command) genFilename(tmpl *template.Template, params map[string]interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), err
}
