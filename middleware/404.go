package middleware

import (
	"fmt"
	"net/http"

	"goji.io/middleware"

	"goji.io"
	"golang.org/x/net/context"
)

//Apply404 is a middleware that supplies a custom 404 handler
//TODO: Change it to accept a custom 404 handler
func Apply404(h goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		handler := middleware.Handler(ctx)
		if handler == nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(404)
			fmt.Fprintln(w, "404")
			return
		}
		h.ServeHTTPC(ctx, w, r)

	})
}
