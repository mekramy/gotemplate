package gotemplate

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
)

// viewPipe registers a custom "view" function for rendering child
// inside layout template. this function returns error if child render
// fail or view called from non-layout view.
func viewPipe(engine *template.Template, data []byte) {
	engine.Funcs(map[string]any{
		"view": func() (template.HTML, error) {
			if data == nil {
				return "", errors.New("layout template called without view")
			} else {
				return template.HTML(data), nil
			}
		},
	})
}

// existsPipe registers a custom "exists" function to the template engine.
// The "exists" function checks if a template with the given name or path exists.
func existsPipe(engine *template.Template) {
	engine.Funcs(map[string]any{
		"exists": func(name string) bool {
			return engine.Lookup(name) != nil
		},
	})
}

// includePipe registers a custom "include" function to the template engine.
// The "include" function includes and executes a template with the given name or path,
// and do nothing if the template does not exist.
func includePipe(engine *template.Template) {
	engine.Funcs(map[string]any{
		"include": func(name string, data ...any) (template.HTML, error) {
			tpl := engine.Lookup(name)
			if tpl == nil {
				return "", nil
			}

			var v any = nil
			if (len(data)) > 0 {
				v = data[0]
			}

			var buf bytes.Buffer
			err := tpl.Execute(&buf, underlyingValue(v))
			if err != nil {
				return "", err
			}

			return template.HTML(buf.String()), nil
		},
	})
}

// requirePipe registers a custom "require" function to the template engine.
// The "require" function includes and executes a template with the given name or path,
// returning an error if the template does not exist.
func requirePipe(engine *template.Template) {
	engine.Funcs(map[string]any{
		"require": func(name string, data ...any) (template.HTML, error) {
			tpl := engine.Lookup(name)
			if tpl == nil {
				return "", fmt.Errorf("template %s does not exist", name)
			}

			var v any = nil
			if (len(data)) > 0 {
				v = data[0]
			}

			var buf bytes.Buffer
			err := tpl.Execute(&buf, underlyingValue(v))
			if err != nil {
				return "", err
			}

			return template.HTML(buf.String()), nil
		},
	})
}
