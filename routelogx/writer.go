package routelogx

import (
	"log"
	"net/http"

	"github.com/monstercat/gologx"
)

type Writer struct {
	ctx *logx.LogHandler
	Context
}

type Context struct {
	Method  string
	Path    string
	IP      string
	Body    interface{}
	Headers interface{}
}

type HostLog struct {
	logx.StdLog
	Ctx Context
}

func (l HostLog) Context() interface{} {
	return l.Ctx
}

// To follow the io.Writer interface, which is required for
// log.Logger to work.
//
// This will create a specialized RouteLog which will be
// sent to the LogContext's handlers.
func (w *Writer) Write(byt []byte) (int, error) {
	log := &HostLog{
		Ctx: w.Context,
	}
	log.SetMessage(byt)
	return w.ctx.Run(log)
}

// Creates a logger with values from http Request. The intended use is like
// the following in the case of GIN:
// ...
//    r.GET("...", func(c *gin.Context) {
//        routeHandler(c, NewLogger(c.Request, ctx))
//    }
// ...
//
// func routeHandler(c *gin.Context, log *log.Logger) {
//   // no changes to the use of log from before.
//   ...
// }
func NewLogger(r *http.Request, ctx *logx.LogHandler) *log.Logger {
	var b []byte
	if r.Body != nil && r.GetBody != nil {
		body, err := r.GetBody()
		if err == nil {
			_, _ = body.Read(b)
		}
	}

	w := &Writer{
		ctx: ctx,
		Context: Context{
			Method:  r.Method,
			Path:    r.URL.Path,
			IP:      r.RemoteAddr,
			Body:    b,
			Headers: r.Header,
		},
	}
	return log.New(w, "", 0)
}
