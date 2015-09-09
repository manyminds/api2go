package api2go

import (
	"golang.org/x/net/context"
	"time"
)

// APIContext api2go context for handlers, nil implementations related to Deadline and Done.
type APIContext struct {
	keys map[string]interface{}
}

// Set a string key value in the context
func (c *APIContext) Set(key string, value interface{}) {
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}
	c.keys[key] = value
}

// Get a key value from the context
func (c *APIContext) Get(key string) (value interface{}, exists bool) {
	if c.keys != nil {
		value, exists = c.keys[key]
	}
	return
}

// reset resets all values on Context, making it safe to reuse
func (c *APIContext) reset() {
	c.keys = nil
}

// Deadline implements net/context
func (c *APIContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done implements net/context
func (c *APIContext) Done() <-chan struct{} {
	return nil
}

// Err implements net/context
func (c *APIContext) Err() error {
	return nil
}

// Value implements net/context
func (c *APIContext) Value(key interface{}) interface{} {
	if keyAsString, ok := key.(string); ok {
		val, _ := c.Get(keyAsString)
		return val
	}
	return nil
}

// Compile time check
var _ context.Context = &APIContext{}

// ContextQueryParams fetches the QueryParams if Set
func ContextQueryParams(c *APIContext) map[string][]string {
	qp, ok := c.Get("QueryParams")
	if ok == true {
		qp = make(map[string][]string)
		c.Set("QueryParams", qp)
	}
	return qp.(map[string][]string)
}
