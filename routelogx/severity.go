package routelogx

import (
	"fmt"
	"net/http"
	"os"

	"github.com/monstercat/gologx"
)

type Severity string

const (
	SeverityInfo  Severity = "INFO"
	SeverityFatal Severity = "FATAL"
)

type Logger struct {
	ctx *logx.LogHandler
	Context
}

type ContextWithSeverity struct {
	Context
	Severity Severity
}

type LogWithSeverity struct {
	logx.StdLog
	ContextWithSeverity
}

func (l LogWithSeverity) Context() interface{} {
	return l.ContextWithSeverity
}

func (w *Logger) log(severity Severity, message []byte) (int, error) {
	log := LogWithSeverity{
		ContextWithSeverity: ContextWithSeverity{
			Context:  w.Context,
			Severity: severity,
		},
	}
	log.SetMessage(message)
	return w.ctx.Run(&log)
}

func (w *Logger) Write(byt []byte) (int, error) {
	return w.log(SeverityInfo, byt)
}

func (w *Logger) Print(v ...interface{}) {
	w.log(SeverityInfo, []byte(fmt.Sprint(v...)))
}

func (w *Logger) Println(v ...interface{}) {
	w.log(SeverityInfo, []byte(fmt.Sprintln(v...)))
}

func (w *Logger) Printf(format string, v ...interface{}) {
	w.log(SeverityInfo, []byte(fmt.Sprintf(format, v...)))
}

func (w *Logger) Fatal(v ...interface{}) {
	w.log(SeverityFatal, []byte(fmt.Sprint(v...)))
	os.Exit(1)
}

func (w *Logger) Fatalf(format string, v ...interface{}) {
	w.log(SeverityFatal, []byte(fmt.Sprintf(format, v...)))
	os.Exit(1)
}

func (w *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	w.log(SeverityFatal, []byte(s))
	panic(s)
}

func (w *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	w.log(SeverityFatal, []byte(s))
	panic(s)
}

// Creates a logger that mimics the log.Print functions as well as
// more functions for severity.
//
// Info & Print are the same.
func NewLoggerWithSeverity(r *http.Request, ctx *logx.LogHandler) *Logger {
	var b []byte
	body, err := r.GetBody()
	if err == nil {
		_, _ = body.Read(b)
	}

	return &Logger{
		ctx: ctx,
		Context: Context{
			Method:  r.Method,
			Path:    r.URL.Path,
			IP:      r.RemoteAddr,
			Body:    b,
			Headers: r.Header,
		},
	}
}
