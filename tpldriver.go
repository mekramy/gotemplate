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

func (engine *tplEngine) Render(w io.Writer, view string, data interface{}, layouts ...string) error {
	var err error

	// Resolve layout
	layout := ""
	if len(layouts) > 0 {
		layout = layouts[0]
	}

	// Reload on development mode
	if engine.option.Dev {
		if err := engine.Load(); err != nil {
			return err
		}
	}

	// Normalize
	view = toPath(view, engine.option.root, engine.option.extension)
	layout = toPath(layout, engine.option.root, engine.option.extension)
	viewNorm := toName(view, engine.option.root, engine.option.extension)
	layoutNorm := toName(layout, engine.option.root, engine.option.extension)
	key := viewNorm + ":" + layoutNorm

	// Check partials render
	if engine.partialRx != nil && engine.partialRx.MatchString(view) {
		return fmt.Errorf("%s partial cannot render directly", view)
	}
	if layout != "" && engine.partialRx != nil && engine.partialRx.MatchString(layout) {
		return fmt.Errorf("%s partial cannot render directly", layout)
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
			_, err := tpl.New("view::" + viewNorm).Parse(string(raw))
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
				_, err := tpl.New("layout::" + layoutNorm).Parse(string(raw))
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
		return tpl.ExecuteTemplate(w, "view::"+viewNorm, underlyingValue(data))
	} else {
		// Render child view to layout
		var buf bytes.Buffer
		err = tpl.ExecuteTemplate(&buf, "view::"+viewNorm, underlyingValue(data))
		if err != nil {
			return err
		}
		viewPipe(tpl, buf.Bytes())

		return tpl.ExecuteTemplate(w, "layout::"+layoutNorm, underlyingValue(data))
	}
}

func (engine *tplEngine) Compile(name, layout string, data any) ([]byte, error) {
	var buf bytes.Buffer
	err := engine.Render(&buf, name, data, layout)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
