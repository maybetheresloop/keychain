package resp

import (
	"fmt"
)

type InvalidArrayLength struct {
	length int64
}

func (e *InvalidArrayLength) Error() string {
	return fmt.Sprintf("invalid array length: %d", e.length)
}

type InvalidBulkStringLength struct {
	length int64
}

func (e *InvalidBulkStringLength) Error() string {
	return fmt.Sprintf("invalid bulk string length: %d", e.length)
}

type InvalidType struct {
	message string
}

func ErrInvalidType(obj interface{}) *InvalidType {
	return &InvalidType{
		message: fmt.Sprintf("invalid resp type: %T", obj),
	}
}

func (e *InvalidType) Error() string {
	return e.message
}
