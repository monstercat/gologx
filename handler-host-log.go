package logx

import "time"

// Logs being sent to DbHostHandler must follow
// this interface Otherwise, they will be ignored.
type HostLog interface {
	Context() interface{}
	HostLog() BaseHostLog
}

type BaseHostLog struct {
	Type    string
	Time    time.Time
	Message []byte

	id      string
	context []byte
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
