package logx

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

const MsgTypeHeartbeat = "Heartbeat"
const MsgTypeRegister = "Register"
const MsgTypeAuthorization = "Authorization"

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

	// Certificate
	CertFile, KeyFile string

	// Origin machine - the name of the current machine
	Origin string

	// Service - name of the current 'service', if any.
	Service string

	// Password to use to login.
	Password string

	// Duration for which we wait before sending logs
	// to the host process.
	WaitDuration time.Duration

	// Period at which to send the heartbeat.
	HeartBeatDuration time.Duration

	// Database to which the logs are stored temporarily
	DB *sqlx.DB

	// Table name for the logs in said database.
	LogTable string
}

// Host Message is messages that are sent to the host.
type HostMessage struct {
	Id      string
	Type    string
	Time    time.Time
	Message []byte
	Context []byte

	// Only used by the register message
	Origin  string
	Service string
}

type ClientMessage struct {
	Type    string
	Status  ClientMessageStatus
	Message []byte
}

type ClientMessageStatus string

const (
	ClientMessageStatusFailed     ClientMessageStatus = "Failed"
	ClientMessageStatusSuccessful ClientMessageStatus = "Successful"
)

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
		`INSERT INTO `+h.LogTable+`(log_type, log_time, message, context) VALUES ($1, $2, $3, $4, $5, $6)`,
		b.Type,
		b.Time,
		b.Message,
		b.context,
	)

	return len(b.context), nil
}

func (h DbHostHandler) Run(errCh chan error) {

	if err := h.Startup(); err != nil {
		errCh <- err
		return
	}

	conn, err := h.connect()
	if err != nil {
		errCh <- err
		return
	}
	defer conn.Close()

	wrCh := make(chan HostMessage)

	go h.RunHeartbeat(wrCh)
	go h.ReadResponses(conn, errCh)
	go h.SendLogs(wrCh, errCh)

	// This for loop actually writes all the responses.
	for {
		select {
		case msg := <-wrCh:
			if err := h.sendToHost(conn, msg); err != nil {
				errCh <- err
				//TODO: make the connection restart
			}
		}
	}
}

func (h DbHostHandler) SendLogs(wrCh chan HostMessage, errCh chan error) {
	for {
		select {
		case <-time.After(h.WaitDuration):
			var ls []BaseHostLog
			if err := h.DB.Select(&ls, `SELECT * FROM `+h.LogTable); err != nil && err != sql.ErrNoRows {
				errCh <- err
				return
			}
			if len(ls) == 0 {
				continue
			}
			for _, l := range ls {
				// After registration, Host Message doesn't need the
				// origin and service anymore.
				wrCh <- HostMessage{
					Id:      l.id,
					Type:    l.Type,
					Time:    l.Time,
					Message: l.Message,
					Context: l.context,
				}
			}

		}
	}
}

func (h DbHostHandler) Startup() error {
	if h.CertFile == "" || h.KeyFile == "" {
		// TODO: generate the keys!
	}
	return h.Register()
}

func (h DbHostHandler) ReadResponses(conn *tls.Conn, errCh chan error) {
	dec := json.NewDecoder(conn)
	for {
		var m ClientMessage
		if err := dec.Decode(&m); err != nil {
			errCh <- err

		}
	}
}

func (h DbHostHandler) Register() error {
	conn, err := h.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := HostMessage{
		Origin:  h.Origin,
		Service: h.Service,
		Type:    MsgTypeRegister,
		Message: []byte(h.Password),
	}

	if err := h.sendToHost(conn, msg); err != nil {
		return err
	}

	// Read response.
	dec := json.NewDecoder(conn)
	var m ClientMessage
	if err := dec.Decode(&m); err != nil {
		return err
	}

}

func (h DbHostHandler) RunHeartbeat(wrCh chan HostMessage) {
	for {
		select {
		case <-time.After(h.HeartBeatDuration):
			wrCh <- HostMessage{
				Origin:  h.Origin,
				Service: h.Service,
				Type:    MsgTypeHeartbeat,
			}
		}
	}
}

func (h DbHostHandler) connect() (*tls.Conn, error) {
	pair, err := tls.LoadX509KeyPair(h.CertFile, h.KeyFile)
	if err != nil {
		return nil, err
	}
	conn, err := tls.Dial("tcp", h.Endpoint, &tls.Config{
		Certificates:       []tls.Certificate{pair},
		InsecureSkipVerify: true,
	})
	return conn, nil
}

func (h DbHostHandler) sendToHost(conn *tls.Conn, msg HostMessage) error {
	byt, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(byt)
	return err
}
