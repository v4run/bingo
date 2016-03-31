package bingo

import (
	"fmt"
	"net/http"
)

// HTTPErr represents an error that occured while handling a request.
type HTTPErr struct {
	ID         int
	Message    string
	HTTPStatus int
	Strace     string
}

func (e HTTPErr) Error() string {
	return fmt.Sprintf("err: %d: %s: %s", e.ID, e.Message)
}

// NewHTTPErr returns an app Error Instance
func NewHTTPErr(id int, message string, code int) *HTTPErr {
	return &HTTPErr{ID: id, Message: message, HTTPStatus: code}
}

var (
	ErrInternalServer    = &HTTPErr{ID: 1000, Message: "Internal server error", HTTPStatus: http.StatusInternalServerError}
	ErrNotFound          = &HTTPErr{ID: 1002, Message: "Requested content not found", HTTPStatus: http.StatusNotFound}
	ErrInvalidParameters = &HTTPErr{ID: 1003, Message: "Invalid request parameters", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized      = &HTTPErr{ID: 1004, Message: "Unauthorized request", HTTPStatus: http.StatusUnauthorized}
	ErrMethodNotAllowed  = &HTTPErr{ID: 1005, Message: "Method not allowed for the request", HTTPStatus: http.StatusMethodNotAllowed}
)
