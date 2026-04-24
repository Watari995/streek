package apperror

import (
	"errors"
	"net/http"
)

type MyError interface {
	Error() string // satisfies the built-in error interface
	Message() string
	Code() ErrorCode
	Status() int
	SetMessage(msg string) MyError
	SetCode(code ErrorCode) MyError
}

type myError struct {
	generalCode ErrorCode
	message     string
	code        ErrorCode
	status      int
}

func (e *myError) Error() string {
	if e.message != "" {
		return e.message
	}
	if e.code != "" {
		return string(e.code)
	}
	return string(e.generalCode)
}

func (e *myError) Message() string {
	return e.message
}

func (e *myError) Code() ErrorCode {
	if e.code != "" {
		return e.code
	}
	return e.generalCode
}

func (e *myError) Status() int {
	return e.status
}

func (e *myError) SetMessage(msg string) MyError {
	e.message = msg
	return e
}

func (e *myError) SetCode(code ErrorCode) MyError {
	e.code = code
	return e
}

// factory function to create a new MyError
func NewBadRequestError() MyError {
	return &myError{
		generalCode: CodeBadRequest,
		status:      http.StatusBadRequest,
	}
}

func NewUnauthorizedError() MyError {
	return &myError{
		generalCode: CodeUnauthorized,
		status:      http.StatusUnauthorized,
	}
}

func NewForbiddenError() MyError {
	return &myError{
		generalCode: CodeForbidden,
		status:      http.StatusForbidden,
	}
}

func NewNotFoundError() MyError {
	return &myError{
		generalCode: CodeNotFound,
		status:      http.StatusNotFound,
	}
}

func NewConflictError() MyError {
	return &myError{
		generalCode: CodeConflict,
		status:      http.StatusConflict,
	}
}

func NewInternalServerError() MyError {
	return &myError{
		generalCode: CodeInternalError,
		status:      http.StatusInternalServerError,
	}
}

// use handler to convert MyError to HTTP response
func AsMyError(err error) (MyError, bool) {
	if err == nil {
		return nil, false
	}
	var e MyError
	// if err meets the MyError interface, return it and true
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
