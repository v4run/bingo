package bingo

import (
	"io"
	"net/http"
)

// Handler encapsulates the handlers
type Handler struct {
	srv     http.Handler
	closers []io.Closer
}

// Close cleans up any resources
func (h *Handler) Close() error {
	for _, cls := range h.closers {
		if err := cls.Close(); err != nil {
			return err
		}
	}
	return nil
}

// AddCloser adds a closer
func (h *Handler) AddCloser(c io.Closer) {
	h.closers = append(h.closers, c)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.srv.ServeHTTP(w, r)
}

//Wrap wraps http.Handler with the bingo Handler
func Wrap(h http.Handler) *Handler {
	return &Handler{srv: h}
}

//TODO: check constants in https://github.com/labstack/echo/blob/a2d757eddc4d3bf17919d8e4b0a9b60743d38ecc/echo.go
