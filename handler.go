package logx

import "os"

// Handlers are different ways to handle incoming logs.
// For example, the handler provided below simply writes
// the log to STDOUT.
type Handler interface {
	Handle(Log) (int, error)
}

// Wrapper for quick functions.
type HandlerFunc func(l Log) (int, error)
func (h HandlerFunc) Handle(l Log) (int, error) {
	return h(l)
}

// stdHandler is a base handler which prints to the
// standard output.
func stdHandler(l Log) (int, error) {
	return os.Stdout.Write(l.Byte())
}
var StdHandler = HandlerFunc(stdHandler)


