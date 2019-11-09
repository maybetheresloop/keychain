package resp

const (
	SimpleString = byte('+')
	Error        = byte('-')
	Integer      = byte(':')
	BulkString   = byte('$')
	Array        = byte('*')
)

type RespError struct {
	message string
}

func NewRespError(message string) RespError {
	return RespError{message: message}
}

func (e *RespError) Error() string {
	return "RESP error - " + e.message
}
