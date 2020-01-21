package logx

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

// Host Handler will communicate with the server
// and send all detailed logs to said server.
//
// It will try to send the logs to the server within
// a certain frequency, but at certain times
// (e.g., host panic or signal interruption)
// a Flush() command is provided to flush that data
// directly to the server.
//
// This specific host handler will store the
type DbHostHandler struct {
	// This is the endpoint of the host to which
	// this logger is connected.
	Endpoint string

	// Origin machine - the name of the current machine
	Origin string

	// Service - name of the current 'service', if any.
	Service string

	// Duration for which we wait before sending logs
	// to the host process.
	Duration time.Duration

	// Database to which the logs are stored temporarily
	DB *sqlx.DB

	// Table name for the logs in said database.
	LogTable string
}

// Logs being sent to DbHostHandler must follow
// this interface Otherwise, they will be ignored.
type HostLog interface {
	Context() interface{}
	HostLog() BaseHostLog
}

type BaseHostLog struct {
	Time    time.Time
	Message []byte

	id      string
	context []byte
}

type hostMessage struct {
	Origin  string
	Service string
	Time    time.Time
	Message []byte
	Context []byte
}

func (l BaseHostLog) HostLog() BaseHostLog {
	return l
}

func (l BaseHostLog) SetMessage(byt []byte) {
	l.Message = byt
}

func (l BaseHostLog) Byte() []byte {
	return l.Message
}

func (h DbHostHandler) Handle(l Log) (int, error) {
	hostLog, ok := l.(HostLog)
	if !ok {
		return 0, nil
	}

	byt, err := json.Marshal(hostLog.Context())
	if err != nil {
		return 0, err
	}
	b := hostLog.HostLog()
	b.context = byt

	_, err = h.DB.Exec(
		`INSERT INTO `+h.LogTable+`(log_time, message, context) VALUES ($1, $2, $3, $4, $5)`,
		b.Time,
		b.Message,
		b.context,
	)

	return len(b.context), nil
}

func (h DbHostHandler) Run(errCh chan error) {
	now := time.Now()
	for {
		time.Sleep(h.Duration - time.Now().Sub(now))
		now = time.Now()

		var ls []BaseHostLog
		if err := h.DB.Select(&ls, `SELECT * FROM `+h.LogTable); err != nil {
			errCh <- err
			return
		}

		ids := make([]string, 0, len(ls))
		for _, l := range ls {
			msg := hostMessage{
				Origin:  h.Origin,
				Service: h.Service,
				Time:    l.Time,
				Message: l.Message,
				Context: l.context,
			}
			if err := h.sendToHost(msg); err != nil {
				errCh <- err
			}else{
				ids = append(ids, l.id)
			}
		}

		_, err := h.DB.Exec(`DELETE FROM `+h.LogTable+`WHERE id = ANY($1)`, ids)
		if err != nil {
			errCh <- err
		}

	}
}

func (h DbHostHandler) sendToHost(msg hostMessage) error {
	// TODO: send to the host!
	return nil
}