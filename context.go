package logx

import "strings"

// LogContext is a general helper that can be used
// to quickly setup loggers which log to multiple locations.
//
// e.g., a log context with handlers as such
//
// ctx := &LogContext{
//    Handlers: []Handler{
//       HostHandler,
//       StdHandler,
//    },
// }
//
// will send the log to the standard output as well as
// log it to the host handler.
//
// In general, all log types will take a LogHandler as an
// input argument.
type LogHandler struct {
	Handlers []Handler
}

// LoggerErrors is a way to group errors from handlers
// to send at once to the function above.
type LoggerErrors struct {
	Errors []error
}

func (e LoggerErrors) Error() string {
	str := make([]string, len(e.Errors))
	for _, ee := range e.Errors {
		str = append(str, ee.Error())
	}
	return strings.Join(str, "\n")
}

func (e *LoggerErrors) AddError(err error) {
	e.Errors = append(e.Errors, err)
}

func (e LoggerErrors) Return() error {
	if len(e.Errors) == 0 {
		return nil
	}
	return e
}

// Runs the handlers in order.
func (c *LogHandler) Run(l Log) (n int, e error) {
	errs := LoggerErrors{}
	for _, v := range c.Handlers {
		var err error
		n, err = v.Handle(l)
		if err != nil {
			errs.AddError(err)
		}
	}
	return n, errs.Return()
}

func (c *LogHandler) Add(h Handler) *LogHandler {
	if c.Handlers == nil {
		c.Handlers = make([]Handler, 0, 10)
	}
	c.Handlers = append(c.Handlers, h)
	return c
}