package middleware

import (
	"goji.io"
	"runtime/debug"
	"net/http"
	"github.com/hifx/bingo/infra/log"
	"golang.org/x/net/context"
)

func ApplyRecoverer(l log.Logger) func(goji.Handler) goji.Handler {
	return func(handler goji.Handler) goji.Handler {
		fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			reqID := GetReqID(ctx)
			defer func() {
				if err := recover(); err != nil {
					l.Error(
						"type", "Runtime Panic",
						"req_id", reqID,
						"uri", r.RequestURI,
						"method", r.Method,
						"remote", r.RemoteAddr,
						"error_message", err,
						"error_stack", debug.Stack())

					http.Error(w, http.StatusText(500), 500)
				}
			}()
			handler.ServeHTTPC(ctx, w, r)
		}
		return goji.HandlerFunc(fn)
	}
}

