package logx

import "os"

// stdHandler is a base handler which prints to the
// standard output.
func stdHandler(l Log) (int, error) {
	return os.Stdout.Write(l.Byte())
}
var StdHandler = HandlerFunc(stdHandler)


