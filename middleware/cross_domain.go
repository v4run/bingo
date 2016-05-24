package middleware

import (
	"net/http"

	"goji.io"
	"golang.org/x/net/context"
)

func CrossDomainRequestAllower(handler goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		//w.Header().Set("Access-Control-Allow-Credentials", "true")
		handler.ServeHTTPC(ctx, w, r)

	})
}
