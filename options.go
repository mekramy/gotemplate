package gotemplate

import (
	"html/template"
	"strings"
)

// Options represents a configuration option for the Template.
type Options func(*Option)
type Option struct {
	root       string
	partials   string
	extension  string
	leftDelim  string
	rightDelim string

	Dev   bool
	Cache bool
	Pipes template.FuncMap
}

// WithRoot set root directory for templates.
// The default root is "."
func WithRoot(root string) Options {
	root = normalizePath(root)
	return func(opt *Option) {
		if root != "" {
			opt.root = root + "/"
		} else {
			opt.root = "."
		}
	}
}

// WithPartials sets the partials path to template.
func WithPartials(path string) Options {
	path = normalizePath(path)
	return func(opt *Option) {
		if path != "" && path != "." {
			opt.partials = path + "/"
		}
	}
}

// WithExtension sets the file extension for the template driver.
// The default extension is "tpl".
func WithExtension(ext string) Options {
	ext = strings.TrimSpace(ext)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return func(opt *Option) {
		if ext != "" {
			opt.extension = ext
		}
	}
}

// WithDelimeters sets the left and right delimiters for template engine.
// The default delimeter is "{{" and "}}"
func WithDelimeters(left, right string) Options {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	return func(opt *Option) {
		if left != "" && right != "" {
			opt.leftDelim = left
			opt.rightDelim = right
		}
	}
}

// WithEnv sets the environment to dev or production. In dev mode
// cache not worked and Load method called on each compile.
// CAUTION: disable development mode on production.
func WithEnv(isDev bool) Options {
	return func(opt *Option) {
		opt.Dev = isDev
	}
}

// WithCache sets the cache mode for the template driver. Cache disabled by default.
// Enable it to optimize performance on production.
func WithCache() Options {
	return func(opt *Option) {
		opt.Cache = true
	}
}

// WithPipes registers a custom function to be used in the template with the given name.
func WithPipes(name string, fn any) Options {
	name = strings.TrimSpace(name)
	return func(tpl *Option) {
		if name != "" && fn != nil {
			tpl.Pipes[name] = fn
		}
	}
}
