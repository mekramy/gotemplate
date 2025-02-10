package gotemplate

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"regexp"
	"sync"

	"github.com/mekramy/gofs"
)

type tplEngine struct {
	option Option
	fs     gofs.FlexibleFS

	base      *template.Template
	templates map[string]*template.Template
	partialRx *regexp.Regexp
	mutex     sync.RWMutex
}

func (engine *tplEngine) Load() error {
	var err error

	// Safe race condition
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	// Initialize
	engine.templates = make(map[string]*template.Template)
	engine.base = template.New("").
		Delims(engine.option.leftDelim, engine.option.rightDelim).
		Funcs(engine.option.Pipes)

	// Add built-in pipes
	viewPipe(engine.base, nil)
	existsPipe(engine.base)
	includePipe(engine.base)
	requirePipe(engine.base)

	// Generate partial pattern
	if engine.option.partials != "" {
		engine.partialRx, err = regexp.Compile(extPattern(
			engine.option.partials,
			engine.option.extension,
		))
		if err != nil {
			return err
		}
	}

	// Read files from fs
	files, err := engine.fs.Lookup(
		engine.option.root,
		extPattern("", engine.option.extension),
	)
	if err != nil {
		return err
	}

	// Load partials
	if engine.option.partials != "" {
		for _, file := range files {
			// Skip non partials
			if !engine.partialRx.MatchString(file) {
				continue
			}

			// Generate friendly name
			name := toName(file, engine.option.partials, engine.option.extension)
			name = "@partials/" + name

			// Read file
			content, err := engine.fs.ReadFile(file)
			if err != nil {
				return err
			}

			_, err = engine.base.New(name).Parse(string(content))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (engine *tplEngine) Render(w io.Writer, name string, data interface{}, layouts ...string) error {
	var err error

	// Reload on development mode
	if engine.option.Dev {
		if err := engine.Load(); err != nil {
			return err
		}
	}

	// Resolve and normalize view
	view := toPath(name, engine.option.root, engine.option.extension)
	viewId := toName(view, engine.option.root, engine.option.extension)

	// Resolve and normalize layout and partials
	layout := ""
	layoutId := ""
	partials := make([]string, 0)
	partialsId := make([]string, 0)
	if len(layouts) > 0 {
		for i := range layouts {
			if i == 0 {
				layout = toPath(layouts[0], engine.option.root, engine.option.extension)
				layoutId = toName(layout, engine.option.root, engine.option.extension)
			} else if layouts[i] != "" {
				name := toPath(layouts[i], engine.option.root, engine.option.extension)
				id := toName(name, engine.option.root, engine.option.extension)
				partials = append(partials, name)
				partialsId = append(partialsId, id)
			}
		}

	}

	// Generate key
	key := toKey(append([]string{viewId, layoutId}, partialsId...)...)

	// Check partials render
	if engine.partialRx != nil && engine.partialRx.MatchString(view) {
		return fmt.Errorf("%s partial cannot render directly", view)
	}
	if layout != "" && engine.partialRx != nil && engine.partialRx.MatchString(layout) {
		return fmt.Errorf("%s partial cannot render directly", layout)
	}
	for _, partial := range partials {
		if engine.partialRx != nil && engine.partialRx.MatchString(partial) {
			return fmt.Errorf("%s partial already loaded globally", layout)
		}
	}

	// Safe race condition
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()

	// Resolve Template
	tpl, ok := engine.templates[key]
	if !ok {
		// Clone from base engine
		tpl, err = engine.base.Clone()
		if err != nil {
			return err
		}

		// Read and parse view
		if raw, err := engine.fs.ReadFile(view); os.IsNotExist(err) {
			return fmt.Errorf("%s template not found", view)
		} else if err != nil {
			return err
		} else {
			_, err := tpl.New("view::" + viewId).Parse(string(raw))
			if err != nil {
				return err
			}
		}

		// Read and parse layout
		if layout != "" {
			if raw, err := engine.fs.ReadFile(layout); os.IsNotExist(err) {
				return fmt.Errorf("%s layout template not found", layout)
			} else if err != nil {
				return err
			} else {
				_, err := tpl.New("layout::" + layoutId).Parse(string(raw))
				if err != nil {
					return err
				}
			}
		}

		for i := range partials {
			if raw, err := engine.fs.ReadFile(partials[i]); os.IsNotExist(err) {
				return fmt.Errorf("%s partial template not found", partials[i])
			} else if err != nil {
				return err
			} else {
				_, err := tpl.New(partialsId[i]).Parse(string(raw))
				if err != nil {
					return err
				}
			}
		}

		// Store to cache
		if !engine.option.Dev && engine.option.Cache {
			engine.templates[key] = tpl
		}
	}

	// Add built-in pipes
	viewPipe(tpl, nil)
	existsPipe(tpl)
	includePipe(tpl)
	requirePipe(tpl)

	// Render
	if layout == "" {
		return tpl.ExecuteTemplate(w, "view::"+viewId, underlyingValue(data))
	} else {
		// Render child view to layout
		var buf bytes.Buffer
		err = tpl.ExecuteTemplate(&buf, "view::"+viewId, underlyingValue(data))
		if err != nil {
			return err
		}
		viewPipe(tpl, buf.Bytes())

		return tpl.ExecuteTemplate(w, "layout::"+layoutId, underlyingValue(data))
	}
}

func (engine *tplEngine) Compile(name, layout string, data any, partials ...string) ([]byte, error) {
	var buf bytes.Buffer
	err := engine.Render(&buf, name, data, append([]string{layout}, partials...)...)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
