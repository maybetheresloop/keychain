package resp

type RespError struct {
	message string
}

func NewRespError(message string) RespError {
	return RespError{message: message}
}

func (e *RespError) Error() string {
	return "RESP error - " + e.message
}
