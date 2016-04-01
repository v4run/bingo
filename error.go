package bingo

import (
	"net/http"
	"strings"

	"github.com/facebookgo/stack"
)

// HTTPErr represents an error that occured while handling a request.
type HTTPErr struct {
	Message    string
	HTTPStatus int
	Strace     string
}

//Error returns the error string
func (e HTTPErr) Error() string {
	return e.Message
}

//Stack returns the error stack
func (e HTTPErr) Stack() string {
	return e.Strace
}

// NewHTTPErr returns an app Error Instance
func NewHTTPErr(message string, code int) HTTPErr {
	s := stack.Callers(1).String()
	s = s[0:strings.Index(s, "\n")]
	return HTTPErr{Message: message, HTTPStatus: code, Strace: s}
}

var (
	ErrInternalServer    = &HTTPErr{Message: "Internal server error", HTTPStatus: http.StatusInternalServerError}
	ErrNotFound          = &HTTPErr{Message: "Requested content not found", HTTPStatus: http.StatusNotFound}
	ErrInvalidParameters = &HTTPErr{Message: "Invalid request parameters", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized      = &HTTPErr{Message: "Unauthorized request", HTTPStatus: http.StatusUnauthorized}
	ErrMethodNotAllowed  = &HTTPErr{Message: "Method not allowed for the request", HTTPStatus: http.StatusMethodNotAllowed}
)
