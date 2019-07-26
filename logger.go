//Package logx provides logging utility which provides a context to
//an error message.
package logx

import (
	"encoding/json"
)

const (
	LOG   = "Log"
	WARN  = "Warn"
	FATAL = "Fatal"
)

type LogContext struct {
	Context interface{} // context for the current error message
	Name    string      // name for the context. Used as key-value pair.

	Logs []error // logs connected to this context. so that we can return the context and call String()
	// child contexts are also stored here.

	// Parameters that are passed down as defaults.
	Stringer     func(*LogContext) string // Changes the error into a string. Needs to handle children
}

// Creates a new log context, with a name
func New(name string) *LogContext {
	return &LogContext{
		Name: name,
	}
}

// Creates a new log context with a context.
func NewWithContext(name string, ctx interface{}) *LogContext {
	return &LogContext{
		Name:    name,
		Context: ctx,
	}
}

func DefaultStringer(c *LogContext) string {
	m := c.mapper(true)
	str, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}
	return string(str)
}

// Returns the error in a specific map
func (c *LogContext) Error() string {
	if c.Stringer != nil {
		return c.Stringer(c)
	}
	return DefaultStringer(c)
}

func (c *LogContext) childMaps(forString bool) []interface{} {
	if c.Logs == nil || len(c.Logs) == 0 {
		return nil
	}
	child := make([]interface{}, 0, len(c.Logs))
	for _, l := range c.Logs {
		if v, ok := l.(*LogContext); ok {
			if forString {
				child = append(child, v.Error())
			} else {
				child = append(child, v.Map())
			}
		} else {
			child = append(child, l.Error())
		}
	}
	return child
}

// Returns a map of the error. The map will have
// - Message
// - Context
// - Details: map for the children, keyed by [name]
//   it is up to the application to make sure names are unique.
//   blank names will be substituted with UUID, but it is
//   better to put a proper name.
func (c *LogContext) Map() map[string]interface{} {
	return c.mapper(false)
}

func (c *LogContext) mapper(forString bool) map[string]interface{} {
	m := map[string]interface{}{
		"Name":    c.Name,
		"Context": c.Context,
	}
	child := c.childMaps(forString)
	if child != nil {
		m["Logs"] = child
	}
	return m
}

// Logging functions
// ============================================
// Log functions should add the log to the list of "logs".
// These logs can be any type of error, including the LogContext
// which should also satisfy the error interface.
//
// Therefore, they need to return the *log context* and not the log.
func (c *LogContext) Wrap(e error) *LogContext {
	c.addLog(e)
	return c
}
func (c *LogContext) Warn(msg string) *LogContext {
	return c.Wrap(NewWarn(msg))
}
func (c *LogContext) Log(msg string) *LogContext {
	return c.Wrap(NewLog(msg))
}
func (c *LogContext) Fatal(msg string) *LogContext {
	return c.Wrap(NewFatal(msg))
}
func (c *LogContext) addLog(e error) {
	if c.Logs == nil {
		c.Logs = make([]error, 0, 10)
	}
	c.Logs = append(c.Logs, e)
}
