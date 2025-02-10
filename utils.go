package gotemplate

import (
	"path/filepath"
	"regexp"
	"strings"
)

// normalizePath normalize path with join and slashed separator.
func normalizePath(path ...string) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(path...)))
}

// toName convert view path to name.
func toName(path, root, ext string) string {
	if path == "" {
		return ""
	}

	path = strings.TrimPrefix(path, root)
	path = strings.TrimSuffix(path, ext)
	return normalizePath(path)
}

// toPath convert view name to path.
func toPath(path, root, ext string) string {
	if path == "" {
		return ""
	}

	path = strings.TrimPrefix(path, root)
	path = strings.TrimSuffix(path, ext)
	return normalizePath(root, path+ext)
}

// toKey generate key for views list
func toKey(views ...string) string {
	res := ""
	for _, v := range views {
		if res == "" && v != "" {
			res = v
		} else if v != "" {
			res += ":" + v
		}
	}
	return res
}

// underlyingValue get underlying value of context.
func underlyingValue(v any) any {
	switch val := v.(type) {
	case Context:
		return val.data
	case *Context:
		return val.data
	}

	return v
}

// Create regexp pattern for path with extension.
func extPattern(path, ext string) string {
	if path == "" {
		return ".*" + regexp.QuoteMeta(ext)
	}

	return "^" + regexp.QuoteMeta(path) + ".*" + regexp.QuoteMeta(ext)
}
