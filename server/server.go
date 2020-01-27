package server

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/monstercat/logx"
)

type Handler func(db *sqlx.DB, message logx.HostMessage, details ConnDetails)

type Server struct {
	CertFile string
	KeyFile  string

	Password string // Master password to register

	DB *sqlx.DB

	Handlers map[string]Handler

	SigCache      map[string]*Service
	SigCacheMutex sync.RWMutex
}

func (s *Server) CheckPassword(password string) bool {
	return password == s.Password
}

type ConnDetails struct {
	Hash []byte

	// These details will be filled in if the user is
	// authenticated
	Service *Service

	// Write Channel
	WrCh chan logx.ClientMessage
}

func (d *ConnDetails) HandledUnauthorized() bool {
	if !IsAuthorized(*d) {
		d.WriteUnauthorized()
		return true
	}
	return false
}

func (d *ConnDetails) WriteUnauthorized() {
	d.WrCh <- logx.ClientMessage{
		Type:    logx.MsgTypeAuthorization,
		Status:  logx.ClientMessageStatusFailed,
		Message: []byte("Unauthorized"),
	}
}

func (s *Server) Listen(port int) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
	if err != nil {
		return nil, err
	}
	tlsConf := &tls.Config{
		ClientAuth:         tls.RequireAndVerifyClientCert,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		Rand:               rand.Reader,
	}
	return tls.Listen("tcp", ":"+strconv.Itoa(port), tlsConf)
}

func getNextDelay(t time.Duration) time.Duration {
	if t == 0 {
		return 5 * time.Millisecond
	} else {
		t *= 2
	}
	if t > time.Second {
		return time.Second
	}
	return t
}

func (s *Server) Serve(listener net.Listener, eh func(error)) {

	// Initialize this if not yet initialized.
	if s.SigCache == nil {
		s.SigCache = make(map[string]*Service)
	}

	var tempDelay time.Duration

	for {
		conn, err := listener.Accept()
		if err != nil {
			eh(err)

			// If a temporary error, the ntry delay.
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				tempDelay = getNextDelay(tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}

		// handle the connection
		go s.handleConn(conn, eh)
	}
}

func (s *Server) VerifySignature(sig []byte) (*Service, error) {
	if s.SigCache == nil {
		return nil, nil
	}
	s.SigCacheMutex.RLock()
	service, ok := s.SigCache[string(sig)]
	s.SigCacheMutex.RUnlock()
	if ok {
		return service, nil
	}
	service, err := GetServiceByHash(s.DB, sig)
	if err != nil {
		return nil, err
	}
	s.SigCacheMutex.Lock()
	s.SigCache[string(sig)] = service
	s.SigCacheMutex.Unlock()

	return service, nil
}

func (s *Server) handleConn(conn net.Conn, eh func(error)) {
	defer conn.Close()

	// Check handshake
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		eh(errors.New("invalid connection"))
		conn.Close()
		return
	}
	if err := tlsConn.Handshake(); err != nil {
		eh(err)
		conn.Close()
		return
	}

	// Writer channel.
	wrCh := make(chan logx.ClientMessage)
	go func() {
		for {
			select {
			case msg := <-wrCh:
				if err := sendToClient(conn, msg); err != nil {
					eh(err)
				}
			}
		}
	}()

	connDetails := ConnDetails{
		WrCh: wrCh,
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		eh(errors.New("tls connection required"))
		return
	}
	connDetails.Hash = tlsConn.ConnectionState().PeerCertificates[0].Signature

	service, err := s.VerifySignature(connDetails.Hash)
	if err != nil {
		eh(err)
		return
	}
	connDetails.Service = service

	//Parse message right away.
	dec := json.NewDecoder(conn)
	for {
		var m logx.HostMessage
		if err := dec.Decode(&m); err != nil {
			eh(err)
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				conn.Write([]byte("Timeout"))
				return
			}
			if err == io.EOF {
				return
			}
			conn.Write([]byte("400: Could not decode message. " + err.Error()))
			return
		}

		// Special handling for registration type. We need to stop
		// processing if the passwords don't match.
		if m.Type == logx.MsgTypeRegister {
			if !s.CheckPassword(string(m.Message)) {
				sendToClient(conn, logx.ClientMessage{
					Type:    logx.MsgTypeRegister,
					Status:  logx.ClientMessageStatusFailed,
					Message: []byte("Password doesn't match"),
				})
				return
			}
			service, err := s.RegisterHandler(m, connDetails)
			if err != nil {
				sendToClient(conn, logx.ClientMessage{
					Type:    logx.MsgTypeRegister,
					Status:  logx.ClientMessageStatusFailed,
					Message: []byte(err.Error()),
				})
			} else {
				connDetails.Service = service
				sendToClient(conn, logx.ClientMessage{
					Type:   logx.MsgTypeRegister,
					Status: logx.ClientMessageStatusSuccessful,
				})
			}
			continue
		}

		// At this point, all other messages need to have
		// a registered service.
		if connDetails.HandledUnauthorized() {
			continue
		}

		// At this point, we got a proper log message.
		// We can hand off the logging.
		h, ok := s.Handlers[m.Type]
		if !ok {
			h = DefaultMessageHandler
		}
		h(s.DB, m, connDetails)
	}
}

func sendToClient(conn net.Conn, msg logx.ClientMessage) error {
	byt, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(byt)
	return err
}

func IsAuthorized(details ConnDetails) bool {
	if details.Service == nil {
		return false
	}
	if details.Service.Id == "" {
		return false
	}
	return true
}

func DefaultMessageHandler(db *sqlx.DB, msg logx.HostMessage, conn ConnDetails) {
	if !IsAuthorized(conn) {
		return
	}
	err := InsertHostMessage(db, msg, conn.Service.Id)
	if err != nil {
	}
}

func (s *Server) RegisterHandler(msg logx.HostMessage, conn ConnDetails) (*Service, error) {
	db := s.DB

	service, err := GetServiceByName(db, msg.Service)
	if err != nil {
		return nil, errors.New("Service retrieval error. " + err.Error())
	}
	if service != nil && service.Id != "" {
		sig := string(service.SigHash)
		if sig == "" {
			service.SigHash = conn.Hash
			if err := service.UpdateHash(db); err != nil {
				return nil, errors.New("Service registration error. " + err.Error())
			}
			if err := service.UpdateLastSeen(db); err != nil {
				return nil, errors.New("Service registration error. " + err.Error())
			}
		} else if string(service.SigHash) != string(conn.Hash) {
			return nil, errors.New("Service registration error. Hash does not match.")
		}
		return service, nil
	}

	service, err = CreateService(db, msg.Origin, msg.Service, conn.Hash)
	if err != nil {
		return nil, errors.New("Could not register service: " + err.Error())
	}

	s.SigCacheMutex.Lock()
	s.SigCache[string(conn.Hash)] = service
	s.SigCacheMutex.Unlock()
	return service, nil
}

func HeartbeatHandler(db *sqlx.DB, msg logx.HostMessage, conn ConnDetails) {
	if err := conn.Service.UpdateLastSeen(db); err != nil {
		conn.WrCh <- logx.ClientMessage{
			Type:    logx.MsgTypeHeartbeat,
			Status:  logx.ClientMessageStatusFailed,
			Message: []byte("Could not update heartbeat. " + err.Error()),
		}
	}
}

func RouteWriterHandler(db *sqlx.DB, msg logx.HostMessage, conn ConnDetails) {
	//TODO: handler for MsgTypeRouteWriter
	// Decode the context
	// Write to the route.
}

func RouteLoggerHandler(db *sqlx.DB, msg logx.HostMessage, conn ConnDetails) {
	//TODO: handler for MsgTypeRouteWriterWithSeverity
	// Decode the context
	// Write to the route
}
