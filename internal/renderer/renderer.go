package renderer

import (
	"bytes"
	"fmt"
	"github.com/artarts36/gostub/templates"
	"html/template"
	"io"
	"reflect"
)

type Renderer struct {
	templates *template.Template
}

func NewRenderer() (*Renderer, error) {
	rend := &Renderer{}

	tmpl, err := template.New("*.tpl").Funcs(template.FuncMap{
		"include": func(tplName string, kv ...any) (template.HTML, error) {
			params := map[string]interface{}{}

			for i, keyOrVal := range kv {
				if i%2 == 0 {
					continue
				}

				key, ok := kv[i-1].(string)
				if !ok {
					return "", fmt.Errorf("invalid key %q at index %d", keyOrVal, i)
				}

				params[key] = keyOrVal
			}

			buf := bytes.Buffer{}

			err := rend.Render(&buf, tplName, params)
			if err != nil {
				return "", err
			}

			return template.HTML(buf.String()), nil
		},
		"isLast": func(index int, arr interface{}) bool {
			return index == reflect.ValueOf(arr).Len()-1
		},
		"hasNext": func(index int, arr interface{}) bool {
			return index < reflect.ValueOf(arr).Len()-1
		},
		"noEmpty": func(item interface{}) bool {
			switch v := item.(type) {
			case interface{ Len() int }:
				return v.Len() > 0
			}

			return reflect.ValueOf(item).Len() > 0
		},
		"raw": func(val string) template.HTML {
			return template.HTML(val)
		},
		"isOnce": func(arr interface{}) bool {
			return reflect.ValueOf(arr).Len() == 1
		},
		"isMany": func(arr interface{}) bool {
			return reflect.ValueOf(arr).Len() > 1
		},
	}).ParseFS(templates.FS, "*.tpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	rend.templates = tmpl

	return rend, nil
}

func (r *Renderer) Render(w io.Writer, tplName string, params map[string]interface{}) error {
	err := r.templates.ExecuteTemplate(w, tplName, params)
	if err != nil {
		return fmt.Errorf("failed to render %q: %w", tplName, err)
	}
	return nil
}
