package server

import (
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	dbutil "github.com/monstercat/golib/db"
	"github.com/monstercat/logx"
)

// Host Server which stores the incoming logs in a central database
type Server struct {
	CertFile string
	KeyFile  string

	Password string // Master password to register

	DB *sqlx.DB

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
		Message: "Unauthorized",
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

// Exponential backoff (2^x) until duration of 1 second.
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

			// If a temporary error, then try to delay and restart.
			// This uses exponential backoff.
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

// This function verifies the signature of the incoming connection.
// It will return the service that corresponds to the signature,
// if any.
//
// This should be called after Serve() is called, as the
// signature cache needs to be initiated.
func (s *Server) VerifySignature(sig []byte) (*Service, error) {
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

func (s *Server) marshalHash(sig []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(sig))
}

// Handles the connection from the server.
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

	// Writer channel initiation to write messages back to the client.
	wrCh := make(chan logx.ClientMessage)

	done := make(chan bool)
	defer func() {
		done <- true
	}()
	go func() {
		for {
			select {
			case msg := <-wrCh:
				if err := sendToClient(conn, msg); err != nil {
					eh(err)
				}
			case <-done:
				return
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

	// Binary signature is not storagble in UTF-8. Therefore, we need to marshal
	// in a way that it can be represented.
	connDetails.Hash = s.marshalHash(tlsConn.ConnectionState().PeerCertificates[0].Signature)

	service, err := s.VerifySignature(connDetails.Hash)
	if err != sql.ErrNoRows && err != nil {
		eh(err)
		return
	}

	// Service in the connection details would be
	// completed if verified. Otherwise, it would
	// be nil. By being nil, the connection would be
	// considered unauthorized.
	connDetails.Service = service

	//Parse message right away.
	dec := json.NewDecoder(conn)
	for {
		var m logx.HostMessage

		//TODO: log all incoming message errors somewhere including the service details
		// ONLY if the service is available.
		if err := dec.Decode(&m); err != nil {
			eh(err)
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				sendToClient(conn, logx.ClientMessage{
					Type:    logx.MsgTypeDecode,
					Status:  logx.ClientMessageStatusFailed,
					Message: "Timeout",
				})
				return
			}
			if err == io.EOF {
				return
			}
			sendToClient(conn, logx.ClientMessage{
				Type:    logx.MsgTypeDecode,
				Status:  logx.ClientMessageStatusFailed,
				Message: "400: Could not decode message. " + err.Error(),
			})
			return
		}

		// Special handling for registration type. We need to stop
		// processing if the passwords don't match.
		if m.Type == logx.MsgTypeRegister {
			if !s.CheckPassword(string(m.Message)) {
				sendToClient(conn, logx.ClientMessage{
					Type:    logx.MsgTypeRegister,
					Status:  logx.ClientMessageStatusFailed,
					Message: "Password doesn't match",
				})
				return
			}
			service, err := s.RegisterService(m, connDetails)
			if err != nil {
				sendToClient(conn, logx.ClientMessage{
					Type:    logx.MsgTypeRegister,
					Status:  logx.ClientMessageStatusFailed,
					Message: "Could not register service: " + err.Error(),
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
		switch m.Type {
		case logx.MsgTypeHeartbeat:
			HeartbeatHandler(s.DB, m, connDetails)
		default:
			DefaultMessageHandler(s.DB, m, connDetails)
		}
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
		conn.WrCh <- logx.ClientMessage{
			Type:    msg.Type,
			Status:  logx.ClientMessageStatusFailed,
			Id:      msg.Id,
			Message: "Failed to store messaage: " + err.Error(),
		}
		return
	}
	conn.WrCh <- logx.ClientMessage{
		Type:   msg.Type,
		Status: logx.ClientMessageStatusSuccessful,
		Id:     msg.Id,
	}
}

func (s *Server) addToSigCache(service *Service, hash []byte) {
	s.SigCacheMutex.Lock()
	s.SigCache[string(hash)] = service
	s.SigCacheMutex.Unlock()
}

// This function registers the service with the current host.
// if the service is already registered, it will return the registered service
// after ensuring that the name of origin / service is updated.
//
// if the service hasn't been already registered, it will attempt to register the
// service.
func (s *Server) RegisterService(msg logx.HostMessage, conn ConnDetails) (*Service, error) {
	db := s.DB

	// First, check if the service exists by hash!
	service, err := GetServiceByHash(db, conn.Hash)
	if err != sql.ErrNoRows && err != nil {
		return nil, err
	}

	// If the name / machine doesn't match, we can assume the register is trying
	// to update the machine or service name. We can do that here.
	if service != nil {
		if service.Name != msg.Service || service.Machine != msg.Machine {
			service.Name = msg.Service
			service.Machine = msg.Machine
			if err := dbutil.TxNow(db, service.Update); err != nil {
				return nil, err
			}
		}
		s.addToSigCache(service, conn.Hash)
		return service, nil
	}

	// If the service is null, that means we couldn't find it by
	// hash. We should find it by name instead. Then, we can update the hash.
	service, err = GetServiceByName(db, msg.Machine, msg.Service)
	if err != sql.ErrNoRows && err != nil {
		return nil, err
	}
	if service != nil {
		if err := service.UpdateHash(db); err != nil {
			return nil, err
		}
		s.addToSigCache(service, conn.Hash)
		return service, nil
	}

	// Otherwise, we don't have this value in the database *at all*. So
	// we need to insert it.
	service = &Service{
		Machine:  msg.Machine,
		Name:     msg.Service,
		LastSeen: time.Now(),
		SigHash:  conn.Hash,
	}
	if err := dbutil.TxNow(db, service.Insert); err != nil {
		return nil, err
	}
	s.addToSigCache(service, conn.Hash)
	return service, nil
}

func HeartbeatHandler(db *sqlx.DB, msg logx.HostMessage, conn ConnDetails) {
	if err := conn.Service.UpdateLastSeen(db); err != nil {
		conn.WrCh <- logx.ClientMessage{
			Type:    logx.MsgTypeHeartbeat,
			Status:  logx.ClientMessageStatusFailed,
			Message: "Could not update heartbeat. " + err.Error(),
		}
	}
}
