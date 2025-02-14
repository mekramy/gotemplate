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

func (t *tplEngine) Load() error {
	var err error

	// Safe race condition
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Initialize
	t.templates = make(map[string]*template.Template)
	t.base = template.New("").
		Delims(t.option.leftDelim, t.option.rightDelim).
		Funcs(t.option.Pipes)

	// Add built-in pipes
	viewPipe(t.base, nil)
	existsPipe(t.base)
	includePipe(t.base)
	requirePipe(t.base)

	// Generate partial pattern
	if t.option.partials != "" {
		t.partialRx, err = regexp.Compile(extPattern(
			t.option.partials,
			t.option.extension,
		))
		if err != nil {
			return err
		}
	}

	// Read files from fs
	files, err := t.fs.Lookup(
		t.option.root,
		extPattern("", t.option.extension),
	)
	if err != nil {
		return err
	}

	// Load partials
	if t.option.partials != "" {
		for _, file := range files {
			// Skip non partials
			if !t.partialRx.MatchString(file) {
				continue
			}

			// Generate friendly name
			name := toName(file, t.option.partials, t.option.extension)
			name = "@partials/" + name

			// Read file
			content, err := t.fs.ReadFile(file)
			if err != nil {
				return err
			}

			_, err = t.base.New(name).Parse(string(content))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *tplEngine) Render(w io.Writer, name string, data interface{}, layouts ...string) error {
	var err error

	// Reload on development mode
	if t.option.Dev {
		if err := t.Load(); err != nil {
			return err
		}
	}

	// Resolve and normalize view
	view := toPath(name, t.option.root, t.option.extension)
	viewId := toName(view, t.option.root, t.option.extension)

	// Resolve and normalize layout and partials
	layout := ""
	layoutId := ""
	partials := make([]string, 0)
	partialsId := make([]string, 0)
	if len(layouts) > 0 {
		for i := range layouts {
			if i == 0 {
				layout = toPath(layouts[0], t.option.root, t.option.extension)
				layoutId = toName(layout, t.option.root, t.option.extension)
			} else if layouts[i] != "" {
				name := toPath(layouts[i], t.option.root, t.option.extension)
				id := toName(name, t.option.root, t.option.extension)
				partials = append(partials, name)
				partialsId = append(partialsId, id)
			}
		}

	}

	// Generate key
	key := toKey(append([]string{viewId, layoutId}, partialsId...)...)

	// Check partials render
	if t.partialRx != nil && t.partialRx.MatchString(view) {
		return fmt.Errorf("%s partial cannot render directly", view)
	}
	if layout != "" && t.partialRx != nil && t.partialRx.MatchString(layout) {
		return fmt.Errorf("%s partial cannot render directly", layout)
	}
	for _, partial := range partials {
		if t.partialRx != nil && t.partialRx.MatchString(partial) {
			return fmt.Errorf("%s partial already loaded globally", layout)
		}
	}

	// Safe race condition
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Resolve Template
	tpl, ok := t.templates[key]
	if !ok {
		// Clone from base engine
		tpl, err = t.base.Clone()
		if err != nil {
			return err
		}

		// Read and parse view
		if raw, err := t.fs.ReadFile(view); os.IsNotExist(err) {
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
			if raw, err := t.fs.ReadFile(layout); os.IsNotExist(err) {
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
			if raw, err := t.fs.ReadFile(partials[i]); os.IsNotExist(err) {
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
		if !t.option.Dev && t.option.Cache {
			t.templates[key] = tpl
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

func (t *tplEngine) Compile(name, layout string, data any, partials ...string) ([]byte, error) {
	var buf bytes.Buffer
	err := t.Render(&buf, name, data, append([]string{layout}, partials...)...)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
