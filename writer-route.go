package logx

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

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
			Time:    time.Now(),
		},
		RouteContext: w.RouteContext,
	}
	log.SetMessage(byt)
	return w.ctx.Run(log)
}

// Creates a logger with values from GIN. The intended use is like:
//
// ...
//    r.GET("...", func(c *gin.Context) {
//        routeHandler(c, NewGinRouteLogger(c, ctx))
//    }
// ...
//
// func routeHandler(c *gin.Context, log *log.Logger) {
//   // no changes to the use of log from before.
//   ...
// }
//
func NewGinRouteLogger(c *gin.Context, ctx *LogHandler) *log.Logger {
	var b []byte
	body, err := c.Request.GetBody()
	if err == nil {
		_, _ = body.Read(b)
	}
	w := &RouteWriter{
		ctx: ctx,
		RouteContext: RouteContext{
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			IP:      c.ClientIP(),
			Body:    b,
			Headers: c.Request.Header,
		},
	}
	return log.New(w, "", 0)
}
