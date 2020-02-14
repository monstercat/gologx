package logx

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"time"
)

const (
	MsgTypeDecode        = "Decode"
	MsgTypeHeartbeat     = "Heartbeat"
	MsgTypeRegister      = "Register"
	MsgTypeAuthorization = "Authorization"
)

var (
	ErrCertRequired = errors.New("certificate required")
)

// Host Handler will communicate with the server
// and send all detailed logs to said server.
//
// It will try to send the logs to the server within
// a certain frequency, but at certain times
// (e.g., host panic or signal interruption)
// a Flush() command is provided to flush that data
// directly to the server.
type HostHandler struct {
	// This is the endpoint of the host to which
	// this logger is connected.
	Endpoint string

	// Certificate
	CertFile, KeyFile string

	// Certificate pair
	pair tls.Certificate

	// Origin machine - the name of the current machine
	Machine string

	// Service - name of the current 'service', if any.
	Service string

	// Password to use to login.
	Password string

	// Duration for which we wait before sending logs
	// to the host process.
	WaitDuration time.Duration

	// Period at which to send the heartbeat.
	HeartBeatDuration time.Duration

	// Cache to store unsent messages.
	// This should be a directory. The system on startup
	// will attempt to read this directory for any existing files and
	// send them to the host.
	CacheFileLocation string
}

// Host Message is messages that are sent to the host.
type HostMessage struct {
	Id      string
	Type    string
	Time    time.Time
	Message []byte
	Context []byte

	// Only used by the register message
	Machine string
	Service string
}

// Client messages are messsages sent to the client.
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

// Handle handles incoming logs by storing them
// in files in CacheFileLocation.
func (h HostHandler) Handle(l Log) (int, error) {
	hostLog, ok := l.(HostLog)
	if !ok {
		return 0, nil
	}

	// TODO: store in the cache file location

	byt, err := json.Marshal(hostLog.Context())
	if err != nil {
		return 0, err
	}
	b := hostLog.HostLog()
	b.context = byt

	return len(b.context), nil
}

// Run will run the host handler. As it contains
// an infinite loop, it should be called in a
// go routine.
//
// This function will call startup as well, so
// calling startup is not necessary.
func (h HostHandler) Run(errCh chan error) {

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

// SendLogs sends logs to the host. In general, this function
// should not be called directly, as it is called in the Run function.
//
// It does this by reading the CacheFileLocation for any files containing
// a log or logs.
func (h HostHandler) SendLogs(wrCh chan HostMessage, errCh chan error) {
	for {
		select {
		case <-time.After(h.WaitDuration):
			var ls []BaseHostLog

			// Go through the folder to see if there is anything to send.
			// Try to send everything.

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

// Startup starts up the connection to the server. It does this by
// ensuring the certificate and key are present. If they are not, it
// will try to generate the cert and the key.
//
// Then, it will register itself with the host.
func (h HostHandler) Startup() error {
	if h.CertFile == "" || h.KeyFile == "" {
		return ErrCertRequired
	}

	// Try to load the key pair. If they don't exist,
	// then try to create them!
	var err error
	h.pair, err = tls.LoadX509KeyPair(h.CertFile, h.KeyFile)

	// If it has expired, or is invalid, then create new ones as well!
	if err != nil || h.pair.Leaf.NotAfter.Before(time.Now()) {
		cert, key, err := GenerateCerts(time.Hour * 24 * 365)
		if err != nil {
			return err
		}
		if err := WriteCertificate(cert, h.CertFile); err != nil {
			return err
		}
		if err := WritePrivateKey(key, h.KeyFile); err != nil {
			return err
		}

		// This should not error! because we just created it.
		privBytes, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return err
		}
		h.pair, err = tls.X509KeyPair(cert.Raw, privBytes)
		if err != nil {
			return err
		}
	}

	return h.Register()
}

// This function continually reads responses from the server and decodes it. If there
// are any errors, it will send the errors back through the error channel.
// Otherwise, it will process the messages appropriately.
func (h HostHandler) ReadResponses(conn *tls.Conn, errCh chan error) {
	dec := json.NewDecoder(conn)
	for {
		var m ClientMessage
		if err := dec.Decode(&m); err != nil {
			errCh <- err
		}

		// TODO: handle the client message!
	}
}

// Registration with the host involves sending the origin and the
// service information to the host with a Register Message and a password.
//
// The host will associate a hash of the certificate with the
// provided origin and service, as long as the password is correct.
//
// Any future requests can be made using the same certificate
// and will automatically be associated with said service and origin.
//
// This allows for public key encryption for most of the log messages
// which ensures that services with the password cannot easily
// mimic another service/origin.
func (h HostHandler) Register() error {
	conn, err := h.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := HostMessage{
		Machine: h.Machine,
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

	// TODO: read the client message properly.

	return nil
}

// A heartbeat is a signal that is sent to the host to tell the host
// that the client process is still alive.
func (h HostHandler) RunHeartbeat(wrCh chan HostMessage) {
	for {
		select {
		case <-time.After(h.HeartBeatDuration):
			wrCh <- HostMessage{
				Machine: h.Machine,
				Service: h.Service,
				Type:    MsgTypeHeartbeat,
			}
		}
	}
}

func (h HostHandler) connect() (*tls.Conn, error) {
	conn, err := tls.Dial("tcp", h.Endpoint, &tls.Config{
		Certificates:       []tls.Certificate{h.pair},
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (h HostHandler) sendToHost(conn *tls.Conn, msg HostMessage) error {
	byt, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(byt)
	return err
}
