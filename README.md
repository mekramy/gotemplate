# GoTemplate

GoTemplate is a flexible and powerful wrapper for Go standard templating engine, designed to simplify the process of rendering HTML templates with layouts, partials and custom functions.

## Features

- In-Memory caching
- Layout based rendering
- Global partials view
- Support for development and production environments

## Installation

To install GoTemplate, use `go get`:

```sh
go get github.com/mekramy/gotemplate
```

## Template Syntax

For layout template you can use `{{ view }}` function to render child view. All partials template can accessed by `@partials/path/to/file` or `template-name`.

**NOTE**: Use `include` function to import instead of builtin `template` function to prevent errors.

### Builtin functions

- `{{ view }}`: render child template in layout. If used in non-layout template generate error!
- `{{ exists "template name or path" }}`: check if template name or path exists.
- `{{ include "template name or path" (optional data) }}`: includes and executes a template with the given name or path and data if exists.
- `{{ require "template name or path" (optional data) }}`: includes and executes a template with the given name or path and data or returning an error if the template does not exist.

## Usage

### Basic Example

```html
<!-- parts/header.tpl -->
{{ define "site-header" }}
<header>...</header>
{{ end }}

<!-- parts/sub/footer.tpl -->
<footer>...</footer>

<!-- pages/home.tpl -->
<section>
  <h1>Home Page</h1>
  <p>{{ .Title }}</p>
</section>
{{ define "title" }}Home Page{{ end }}

<!-- layout.tpl -->
<html>
  <head>
    {{ if exists "title" }}
    <title>{{ include "title" }}</title>
    {{ else }}
    <title>My App</title>
    {{ end }}
  </head>
  <body>
    {{- require "site-header" . }}
    {{- view }}
    {{- include "@partials/sub/footer" }}
  </body>
</html>
```

```go
package main

import (
    "os"
    "github.com/mekramy/gotemplate"
    "github.com/mekramy/gofs"
)

func main() {
    fs := gofs.NewDir("./views")
    tpl := gotemplate.New(fs, gotemplate.WithPartials("parts"))

    err := tpl.Load()
    if err != nil {
        panic(err)
    }

    data := gotemplate.Ctx().Add("Title", "Hello, World!")
    err = tpl.Render(os.Stdout, "pages/home", data, "layout")
    if err != nil {
        panic(err)
    }
}
```

### Custom Options

```go
tpl := gotemplate.New(fs,
    gotemplate.WithRoot("."),
    gotemplate.WithPartials("partials"),
    gotemplate.WithExtension(".tpl"),
    gotemplate.WithDelimeters("{{", "}}"),
    gotemplate.WithEnv(true),
    gotemplate.WithCache(),
    gotemplate.WithUUIDPipe(),
    gotemplate.WithTernaryPipe(),
)
```

## API

### Template Interface

```go
type Template interface {
    Load() error
    Render(w io.Writer, view string, data interface{}, layouts ...string) error
    Compile(name, layout string, data any) ([]byte, error)
}
```

### Options

- `WithRoot(root string) Options`: Sets the root directory for templates.
- `WithPartials(path string) Options`: Sets the directory for partial templates.
- `WithExtension(ext string) Options`: Sets the file extension for templates.
- `WithDelimeters(left, right string) Options`: Sets the delimiters for template tags.
- `WithEnv(isDev bool) Options`: Sets the environment mode (development or production).
- `WithCache() Options`: Enables template caching.
- `WithPipes(name string, fn any) Options`: Registers a custom function (pipe) for templates.

### Context

Helper struct to pass data to template.

```go
type Context struct {
    data map[string]any
}

func Ctx() *Context
func ToCtx(v any) *Context
func (ctx *Context) Add(k string, v any) *Context
func (ctx *Context) Map() map[string]any
```

### Custom Pipes

- `WithUUIDPipe() Options`: Adds a UUID generation pipe.
- `WithTernaryPipe() Options`: Adds a ternary operation pipe.
- `WithNumberFmtPipe() Options`: Adds a number formatting pipe.
- `WithRegexpFmtPipe() Options`: Adds a regular expression formatting pipe.
- `WithJSONPipe() Options`: Adds a JSON formatting pipe.
- `WithDictPipe() Options`: Adds a dictionary creation pipe.
- `WithIsSetPipe() Options`: Adds a pipe to check if a value is set.
- `WithAlterPipe() Options`: Adds a pipe to alter a value.
- `WithDeepAlterPipe() Options`: Adds a pipe to deeply alter a value.
- `WithBrPipe() Options`: Adds a pipe to convert `\n` to `<br>`.

## License

This project is licensed under the ISC License.
