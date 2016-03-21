package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zenazn/goji/web/mutil"

	"github.com/hifx/bingo/log"
	"goji.io"
	"golang.org/x/net/context"
)

// ApplyLog is a goji middleware that logs all requests to the logger provided
func ApplyLog(l log.Logger) func(goji.Handler) goji.Handler {
	return func(h goji.Handler) goji.Handler {
		return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqid := GetReqID(ctx)

			ww := mutil.WrapWriter(w)
			h.ServeHTTPC(ctx, ww, r)

			latency := float64(time.Since(start)) / float64(time.Millisecond)

			l.Info(
				"req_id", reqid,
				"uri", r.RequestURI,
				"method", r.Method,
				"remote", r.RemoteAddr,
				"status", ww.Status(),
				"latency", fmt.Sprintf("%6.4f ms", latency))
		})
	}
}
