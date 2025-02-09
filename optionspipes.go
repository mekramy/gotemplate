package gotemplate

import (
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/mekramy/goutils"
)

// WithUUIDPipe returns an Options function that adds a "uuid" pipe to the template.
// The "uuid" pipe generates a new UUID string.
//
// code block:
//
//	{{ $id := uuid }}
func WithUUIDPipe() Options {
	return func(tpl *Option) {
		tpl.Pipes["uuid"] = func() string {
			return uuid.NewString()
		}
	}
}

// WithTernaryPipe returns an Options function that adds a "iif" pipe to the template.
// The "iif" pipe acts as a ternary operator, returning y if cond is true, otherwise n.
//
// code block:
//
//	{{ $res := iif .IsAuthorized "YES" "NO" }}
func WithTernaryPipe() Options {
	return func(tpl *Option) {
		tpl.Pipes["iif"] = func(cond bool, y, n any) any {
			if cond {
				return y
			}
			return n
		}
	}
}

// WithNumberFmtPipe returns an Options function that adds a "numberFmt" pipe to the template.
// The "numberFmt" pipe formats numbers according to the specified layout.
//
// code block:
//
//	{{ $formatted := numberFmt "%d $" 1000000 }}
func WithNumberFmtPipe() Options {
	return func(tpl *Option) {
		tpl.Pipes["numberFmt"] = func(layout string, v ...any) string {
			return goutils.FormatNumber(layout, v...)
		}
	}
}

// WithRegexpFmtPipe returns an Options function that adds a "regexpFmt" pipe to the template.
// The "regexpFmt" pipe formats strings using regular expressions.
//
// code block:
//
//	{{ $formatted := regexpFmt "123456" "(\d{2})(\d{3})(\d{1})" "($1) $2-$3" }}
func WithRegexpFmtPipe() Options {
	return func(tpl *Option) {
		tpl.Pipes["regexpFmt"] = func(data, pattern, repl string) (string, error) {
			return goutils.FormatRx(data, pattern, repl)
		}
	}
}

// WithJSONPipe returns an Options function that adds a "toJson" pipe to the template.
// The "toJson" pipe converts data to a JSON string.
//
// code block:
//
//	{{ $encoded := toJson .User }}
func WithJSONPipe() Options {
	return func(opt *Option) {
		opt.Pipes["toJson"] = func(data any) (string, error) {
			res, err := json.Marshal(data)
			if err != nil {
				return "", err
			}
			return string(res), nil
		}
	}
}

// WithDictPipe returns an Options function that adds a "dict" pipe to the template.
// The "dict" pipe creates a map from a list of key-value pairs.
//
// code block:
//
//	{{ $userGlobal := toJson
//				"name" .User.Name
//				"family" .User.Family
//				"email" .User.Email
//	}}
func WithDictPipe() Options {
	return func(tpl *Option) {
		tpl.Pipes["dict"] = func(kv ...any) (map[string]any, error) {
			// Validate item length
			if len(kv)%2 != 0 {
				return nil, fmt.Errorf("invalid number of arguments passed to dict function")
			}

			// Parse and validate values
			dict := make(map[string]any)
			for i := 0; i < len(kv); i += 2 {
				if key, ok := kv[i].(string); !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				} else {
					dict[key] = kv[i+1]
				}
			}
			return dict, nil
		}
	}
}

// WithIsSetPipe returns an Options function that adds an "isSet" pipe to the template.
// The "isSet" pipe checks if a given field is set in a map.
//
// code block:
//
//	{{ if (isSet dataMap "title") }}
//		<h1>{{ dataMap.title }}</h1>
//	{{ else }}
//		<h1>Unknown</h1>
//	{{ end }}
func WithIsSetPipe() Options {
	return func(opt *Option) {
		opt.Pipes["isSet"] = func(data map[string]any, field string) bool {
			_, ok := data[field]
			return ok
		}
	}
}

// WithAlterPipe returns an Options function that adds an "alter" pipe to the template.
// The "alter" pipe returns an alternative value if the original value is nil or not exists.
//
// code block:
//
//	{{ $safeTitle := alter .data.meta.Title "Greeting" }}
func WithAlterPipe() Options {
	return func(opt *Option) {
		opt.Pipes["alter"] = func(val, alt any) any {
			if val == nil {
				return alt
			}
			return val
		}
	}
}

// WithDeepAlterPipe returns an Options function that adds a "deepAlter" pipe to the template.
// The "deepAlter" pipe returns an alternative value if the original value is nil or zero.
//
// code block:
//
//	{{ $safeTitle := deepAlter .data.meta.Title "Greeting" }}
func WithDeepAlterPipe() Options {
	return func(opt *Option) {
		opt.Pipes["deepAlter"] = func(val, alt any) any {
			if val == nil {
				return alt
			}

			v := reflect.ValueOf(val)
			switch v.Kind() {
			case reflect.String, reflect.Slice, reflect.Map, reflect.Chan:
				if v.Len() == 0 {
					return alt
				}
			case reflect.Ptr, reflect.Interface:
				if v.IsNil() {
					return alt
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64:
				if v.IsZero() {
					return alt
				}
			}

			return val
		}
	}
}

// WithBrPipe returns an Options function that adds a "br" pipe to the template.
// The "br" pipe replaces newline characters with HTML line break tags.
//
// code block:
//
//	{{ $out := br .comment }}
func WithBrPipe() Options {
	return func(opt *Option) {
		opt.Pipes["br"] = func(text string) template.HTML {
			escaped := template.HTMLEscapeString(text)
			return template.HTML(strings.ReplaceAll(escaped, "\n", "<br/>"))
		}
	}
}
