package logx

import (
	"log"
	"net/http"
	"time"
)

const RouteHostLogType = "RouteHostLog"

// RouteWriter is used to write messages related to a route.
type RouteWriter struct {
	ctx *LogHandler
	RouteContext
}

type RouteContext struct {
	Method  string
	Path    string
	IP      string
	Body    interface{}
	Headers interface{}
}

type RouteHostLog struct {
	RouteContext
	BaseHostLog
}

func (l RouteHostLog) Context() interface{} {
	return l.RouteContext
}

// To follow the io.Writer interface, which is required for
// log.Logger to work.
//
// This will create a specialized RouteLog which will be
// sent to the LogContext's handlers.
func (w *RouteWriter) Write(byt []byte) (int, error) {
	log := RouteHostLog{
		BaseHostLog: BaseHostLog{
			Time: time.Now(),

			// Used for storing data in special tables in the backend.
			Type: RouteHostLogType,
		},
		RouteContext: w.RouteContext,
	}
	log.SetMessage(byt)
	return w.ctx.Run(log)
}

// Creates a logger with values from http Request. The intended use is like
// the following in the case of GIN:
// ...
//    r.GET("...", func(c *gin.Context) {
//        routeHandler(c, NewRouteLogger(c.Request, ctx))
//    }
// ...
//
// func routeHandler(c *gin.Context, log *log.Logger) {
//   // no changes to the use of log from before.
//   ...
// }
//
func NewRouteLogger(r *http.Request, ctx *LogHandler) *log.Logger {
	var b []byte
	body, err := r.GetBody()
	if err == nil {
		_, _ = body.Read(b)
	}

	w := &RouteWriter{
		ctx: ctx,
		RouteContext: RouteContext{
			Method:  r.Method,
			Path:    r.URL.Path,
			IP:      r.RemoteAddr,
			Body:    b,
			Headers: r.Header,
		},
	}
	return log.New(w, "", 0)
}

