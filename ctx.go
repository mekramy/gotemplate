package gotemplate

// Context represents a collection of key-value pairs value for template data.
type Context struct {
	data map[string]any
}

// Ctx initializes and returns a new Context instance.
func Ctx() *Context {
	return &Context{
		data: make(map[string]any),
	}
}

// ToCtx converts the given value to a Context instance.
// This function returns empty context if value not a valid map or context.
func ToCtx(v any) *Context {
	switch val := v.(type) {
	case map[string]any:
		return &Context{data: val}
	case Context:
		return &val
	case *Context:
		return val
	}
	return Ctx()
}

// Add adds a key-value pair to the Vars instance.
func (ctx *Context) Add(k string, v any) *Context {
	if k != "" {
		ctx.data[k] = v
	}
	return ctx
}

// Map returns underlying data map of context
func (ctx *Context) Map() map[string]any {
	return ctx.data
}
