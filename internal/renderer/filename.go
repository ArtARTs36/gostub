package renderer

import (
	"bytes"
	"fmt"
	"github.com/artarts36/gostub/internal/golang"
	"html/template"
)

type NameGenerator struct {
	commonFilenameTpl    *template.Template
	perMethodFilenameTpl *template.Template
	perTypeFilenameTpl   *template.Template

	typeNameTpl *template.Template
}

func NewNameGenerator(
	commonFilenameTpl string,
	perMethodFileTpl string,
	perTypeFilenameTpl string,
	typeNameTpl string,
) (*NameGenerator, error) {
	var err error

	generator := &NameGenerator{}

	generator.commonFilenameTpl, err = template.New("stub-common-filename-template").Parse(commonFilenameTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to compile common filename template: %w", err)
	}

	generator.perMethodFilenameTpl, err = template.New("stub-per-method-filename-template").Parse(perMethodFileTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to compile stub per method filename template: %w", err)
	}

	generator.typeNameTpl, err = template.New("stub-struct-name-template").Parse(typeNameTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to compile struct name template: %w", err)
	}

	generator.perTypeFilenameTpl, err = template.New("stub-per-struct-name-filename-template").Parse(perTypeFilenameTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to compile struct name template: %w", err)
	}

	return generator, nil
}

func (g *NameGenerator) GenerateCommonFilename() (string, error) {
	return g.genName(g.commonFilenameTpl, map[string]interface{}{})
}

func (g *NameGenerator) GenerateStubStructFilename(goInterface *golang.GoInterface) (string, error) {
	return g.genName(g.perTypeFilenameTpl, map[string]interface{}{
		"Interface": goInterface,
	})
}

func (g *NameGenerator) GenerateStubPerMethodFilename(typ golang.Type, method *golang.GoMethod) (string, error) {
	return g.genName(g.perMethodFilenameTpl, map[string]interface{}{
		"Type":      typ,
		"Method":    method,
		"Interface": typ.Interface,
	})
}

func (g *NameGenerator) GenerateTypeName(goInterface *golang.GoInterface) (string, error) {
	return g.genName(g.typeNameTpl, map[string]interface{}{
		"Interface": goInterface,
	})
}

func (g *NameGenerator) genName(tmpl *template.Template, params map[string]interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), err
}
