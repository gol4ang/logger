package logger

import (
	"strings"
)

type Context map[string]Field

func (c *Context) Merge(context Context) *Context {
	for name, field := range context {
		c.Set(name, field)
	}
	return c
}

func (c *Context) Set(name string, value Field) *Context {
	(*c)[name] = value
	return c
}

func (c *Context) Has(name string) bool {
	_, ok := (*c)[name]
	return ok
}

func (c *Context) Get(name string, field *Field) Field {
	if c.Has(name) {
		return (*c)[name]
	}
	return *field
}

func (c *Context) Add(name string, value interface{}) *Context {
	return c.Set(name, Any(value))
}

func (c *Context) Skip(name string, value string) *Context {
	return c.Set(name, Skip(value))
}
func (c *Context) Binary(name string, value []byte) *Context {
	return c.Set(name, Binary(value))
}
func (c *Context) ByteString(name string, value []byte) *Context {
	return c.Set(name, ByteString(value))
}

func (c *Context) stringTo(builder *strings.Builder) *Context {
	if len(*c) == 0 {
		builder.WriteString("nil")
		return c
	}
	i := 0
	for name, field := range *c {
		if i != 0 {
			builder.WriteRune(' ')
		}
		builder.WriteString(name)
		builder.WriteRune(':')
		builder.WriteString(field.String())
		i++
	}
	return c
}

func (c *Context) String() string {
	if len(*c) == 0 {
		return "<nil>"
	}
	builder := &strings.Builder{}
	c.stringTo(builder)
	return builder.String()
}

// TODO MOVE context serialization
// fmt GoStringer
// usefull when you fmt.Printf("%#v", GoStringer)
func (c *Context) GoString() string {
	builder := &strings.Builder{}
	builder.WriteString("logger.context<")
	c.stringTo(builder)
	builder.WriteString(">")
	return builder.String()
}

func Ctx(name string, value interface{}) *Context {
	return NewContext().Add(name, value)
}

func NewContext() *Context {
	return &Context{}
}
