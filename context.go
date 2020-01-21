package logx

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
type LogHandler struct {
	Handlers []Handler
}

// Runs the handlers in order.
func (c *LogHandler) Run(l Log) (n int, e error) {

	//TODO: handle errors better.
	for _, v := range c.Handlers {
		n, e = v.Handle(l)
	}
	return
}
