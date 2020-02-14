package logx

import (
	"io"
)

// This is the base log object.
// It is what will be written to the LogContext
// by the LogWriters.
type Log interface {
	// Returns the message as a byte array.
	Byte() []byte

	// Sets the message bytes
	SetMessage([]byte)
}

// LogWriters are what are input into the default logging
// system. The application can create its own, however, some
// loggers are created here by default.
type LogWriter interface {
	io.Writer
}

type StdLog struct {
	Message []byte
}
func (l *StdLog) SetMessage(b []byte){
	l.Message = b
}
func (l *StdLog) Byte() []byte {
	return l.Message
}

type StdLogWriter struct {
	ctx *LogHandler
}

func (w *StdLogWriter) Write(byt []byte) (int, error) {
	log := &StdLog{
		Message: byt,
	}
	return w.ctx.Run(log)
}