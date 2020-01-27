package logx

import (
	"fmt"
	"net/http"
	"time"
)

const RouteHostLogWithSeverityType = "RouteHostLogWithSeverity"

type Severity string

const (
	SeverityInfo  Severity = "INFO"
	SeverityWarn  Severity = "WARN"
	SeverityFatal Severity = "FATAL"
)

type RouteLogger struct {
	ctx *LogHandler
	RouteContext
}

type RouteHostLogWithSeverity struct {
	BaseHostLog
	RouteContextWithSeverity
}

type RouteContextWithSeverity struct {
	RouteContext
	Severity Severity
}

func (l RouteHostLogWithSeverity) Context() interface{} {
	return l.RouteContextWithSeverity
}

func (w *RouteLogger) log(severity Severity, message []byte) (int, error) {
	log := RouteHostLogWithSeverity{
		BaseHostLog: BaseHostLog{
			Time: time.Now(),

			// Used for storing data in special tables in the backend.
			Type: RouteHostLogWithSeverityType,
		},
		RouteContextWithSeverity: RouteContextWithSeverity{
			RouteContext: w.RouteContext,
			Severity:     severity,
		},
	}
	log.SetMessage(message)
	return w.ctx.Run(log)
}

func (w *RouteLogger) Write(byt []byte) (int, error) {
	return w.log(SeverityInfo, byt)
}

func (w *RouteLogger) Print(v ...interface{}) {
	w.Info(v...)
}

func (w *RouteLogger) Println(v ...interface{}) {
	w.log(SeverityInfo, []byte(fmt.Sprintln(v...)))
}

func (w *RouteLogger) Printf(format string, v ...interface{}) {
	w.Infof(format, v...)
}

func (w *RouteLogger) Info(v ...interface{}) {
	w.log(SeverityInfo, []byte(fmt.Sprint(v...)))
}

func (w *RouteLogger) Infof(format string, v ...interface{}) {
	w.log(SeverityInfo, []byte(fmt.Sprintf(format, v...)))
}

func (w *RouteLogger) Warn(v ...interface{}) {
	w.log(SeverityWarn, []byte(fmt.Sprint(v...)))
}

func (w *RouteLogger) Warnf(format string, v ...interface{}) {
	w.log(SeverityWarn, []byte(fmt.Sprintf(format, v...)))
}

func (w *RouteLogger) Fatal(v ...interface{}) {
	w.log(SeverityFatal, []byte(fmt.Sprint(v...)))
}

func (w *RouteLogger) Fatalf(format string, v ...interface{}) {
	w.log(SeverityFatal, []byte(fmt.Sprintf(format, v...)))
}

// Creates a logger that mimics the log.Print functions as well as
// more functions for severity.
//
// Info & Print are the same.
func NewRouteLoggerWithSeverity(r *http.Request, ctx *LogHandler) *RouteLogger {
	var b []byte
	body, err := r.GetBody()
	if err == nil {
		_, _ = body.Read(b)
	}

	return &RouteLogger{
		ctx: ctx,
		RouteContext: RouteContext{
			Method:  r.Method,
			Path:    r.URL.Path,
			IP:      r.RemoteAddr,
			Body:    b,
			Headers: r.Header,
		},
	}
}
