package bingo

import (
	"net/http"
	"strings"

	"github.com/facebookgo/stack"
)

// Err represents an error that occured while handling a request.
type Err struct {
	Message string
	Code    int
	Strace  string
}

//Error returns the error string
func (e Err) Error() string {
	return e.Message
}

//ErrCode returns the http status code
func (e Err) ErrCode() int {
	return e.Code
}

//Stack returns the error stack
func (e Err) Stack() string {
	return e.Strace
}

// NewErr returns an app Error Instance
func NewErr(message string, code int) Err {
	s := stack.Callers(1).String()
	s = s[0:strings.Index(s, "\n")]
	return Err{Message: message, Code: code, Strace: s}
}

var (
	ErrInternalServer    = &Err{Message: "Internal server error", Code: http.StatusInternalServerError}
	ErrNotFound          = &Err{Message: "Requested content not found", Code: http.StatusNotFound}
	ErrInvalidParameters = &Err{Message: "Invalid request parameters", Code: http.StatusBadRequest}
	ErrUnauthorized      = &Err{Message: "Unauthorized request", Code: http.StatusUnauthorized}
	ErrMethodNotAllowed  = &Err{Message: "Method not allowed for the request", Code: http.StatusMethodNotAllowed}
)
