package logx

// Type error is the special error type made by logx.
// It is an error with a context.
type Log struct {
	BaseError error  // base error object for comparison purposes.
	Message   string // custom message for this specific error. will otherwise default to the message in BaseError
	Severity  string // Severity of the error / log
}

func NewLog(msg string) *Log {
	return newLog(LOG, msg)
}
func NewWarn(msg string) *Log {
	return newLog(WARN, msg)
}
func NewFatal(msg string) *Log {
	return newLog(FATAL, msg)
}

func newLog(severity, msg string) *Log {
	return &Log{
		Severity: severity,
		Message:  msg,
	}
}

func (e *Log) String() string {
	return e.Error()
}

func (e *Log) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.BaseError.Error()
}

// Checks for equality with
// - object pointer
// - base error equality
func (e *Log) Equals(b error) bool {
	if e == b || e.BaseError == b {
		return true
	}
	if v, ok := b.(*Log); !ok {
		return false
	} else if v.BaseError == e.BaseError {
		return true
	}
	return false
}
