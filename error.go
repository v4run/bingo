package bingo

import "github.com/hifx/errgo"

import "strings"

// Err represents an error that occured while handling a request.
// swagger:response error
type Err struct {
	errgo.Err
	stack string
}

//Stack returns the error stack
func (e Err) Stack() string {
	return e.stack
}

// NewErr returns an app Error Instance
func NewErr(message string, code int) Err {
	err := Err{Err: errgo.NewErr(message)}
	err.SetCode(code)
	err.SetLocation(1)
	err.stack = strings.Join(err.StackTrace(), ";")
	return err
}

// NewErrWithCause returns an app Error Instance caused by other error.
func NewErrWithCause(other error, code int, format string, args ...interface{}) Err {
	err := Err{Err: errgo.NewErrWithCause(other, format, args...)}
	err.SetCode(code)
	err.SetLocation(1)
	err.stack = strings.Join(err.StackTrace(), ";")
	return err
}

var (
	// ErrInternalServer represents an internal server error.
	ErrInternalServer = errgo.InternalServerf("Internal server error")
	// ErrNotFound represents not found error.
	ErrNotFound = errgo.NotFoundf("Requested content not found")
	// ErrInvalidParameters represents invalid parameters error.
	ErrInvalidParameters = errgo.BadRequestf("Invalid request parameters")
	// ErrUnauthorized represents unauthorized error.
	ErrUnauthorized = errgo.Unauthorizedf("Unauthorized request")
	// ErrMethodNotAllowed represents method not allowed error.
	ErrMethodNotAllowed = errgo.MethodNotAllowedf("Method not allowed for the request")
)
