package gotemplate

import (
	"html/template"
	"io"

	"github.com/mekramy/gofs"
)

// Template represents a template interface with methods for checking existence,
// compiling, loading, and rendering templates.
type Template interface {
	// Load loads shared templates from the filesystem.
	Load() error

	// Render renders a template to the provided writer with
	// the given view, data, and optional layouts.
	Render(w io.Writer, view string, data interface{}, layouts ...string) error

	// Compile compiles a template with the given name and layout and data.
	Compile(name, layout string, data any, partials ...string) ([]byte, error)
}

// New creates a new Template instance with the provided filesystem and options.
func New(fs gofs.FlexibleFS, options ...Options) Template {
	// Create option
	option := &Option{
		root:       ".",
		partials:   "",
		extension:  ".tpl",
		leftDelim:  "{{",
		rightDelim: "}}",
		Dev:        false,
		Cache:      false,
		Pipes:      make(template.FuncMap),
	}
	for _, opt := range options {
		opt(option)
	}

	// Create template driver
	driver := new(tplEngine)
	driver.option = *option
	driver.fs = fs
	return driver
}
