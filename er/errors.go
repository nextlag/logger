package er

import (
	"errors"
	"fmt"
)

// ResponseError - is a response error with code.
type ResponseError struct {
	Message string
	Err     error
}

// Error returns pair data message and error of ResponseError.
func (e *ResponseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

// Unwrap returns an error for implementation errors.Is and errors.As.
func (e *ResponseError) Unwrap() error {
	return e.Err
}

// Is checks whether the error or its nested error is of a certain type.
func (e *ResponseError) Is(target error) bool {
	if target == nil {
		return e == nil
	}

	return errors.Is(e.Err, target)
}

// As checks whether the error can be cast to a specific type.
func (e *ResponseError) As(target interface{}) bool {
	if err, ok := target.(**ResponseError); ok {
		*err = e

		return true
	}

	return errors.As(e.Err, &target)
}

// Format создает новую ошибку с отформатированным сообщением.
func Format(format string, args ...interface{}) error {
	return &ResponseError{
		Message: fmt.Sprintf(format, args...),
	}
}

// FormatWithPrefix создает новую ошибку с префиксом и отформатированным сообщением.
func FormatWithPrefix(prefix, format string, args ...interface{}) error {
	return &ResponseError{
		Message: fmt.Sprintf("%s: %s", prefix, fmt.Sprintf(format, args...)),
	}
}

var (
	ErrIncorrectReader = &ResponseError{Message: "incorrect reader"}
	ErrReadBuffer      = &ResponseError{Message: "failed to read buffer"}
	ErrConfigParse     = &ResponseError{Message: "Не удалось проанализировать конфигурацию"}
)
