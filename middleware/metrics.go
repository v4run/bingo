package middleware

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/hifx/bingo/infra/metrics"
	"github.com/hifx/bingo/middleware/mutil"
	"goji.io"
	"goji.io/middleware"
	"golang.org/x/net/context"
)

//PATKEY is the key used to store matched patterns in context
const PATKEY = "metrics.pattern"

/*
ApplyStats wraps a handler to track request counts and response status
code counts name-spaced by goji Pattern. It will only include
patterns that implement fmt.Stringer. For example, if a request
matches the pattern /foo/:bar and returns a 204 status code, it
will increment "foo.:bar.request" and "foo.:bar.response.204".
In addition it updates a histogram with response latencies

largely influenced by https://github.com/metcalf/saypi
*/
func ApplyStats(h goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var patterns []goji.Pattern
		start := time.Now()

		curr := middleware.Pattern(ctx)
		if curr != nil {
			patterns = append(patterns, curr)
		}

		ctx = context.WithValue(ctx, PATKEY, &patterns)

		ww := mutil.WrapWriter(w)
		h.ServeHTTPC(ctx, ww, r)

		patstrs := make([]string, len(patterns))
		for i, pattern := range patterns {
			patstr, ok := pattern.(fmt.Stringer)
			if !ok {
				continue
			}

			patstrs[i] = strings.TrimSuffix(patstr.String(), "/*")
		}
		patclean := strings.Trim(strings.Replace(path.Join(patstrs...), "/", ".", -1), ".")

		if patclean != "" {
			metrics.AddCounter(fmt.Sprintf("%s.request", patclean))
			metrics.AddCounter(fmt.Sprintf("%s.response.%d", patclean, ww.Status()))
			metrics.UpdateTimerSince(fmt.Sprintf("%s.latency", patclean), start)
		}
	})
}

// ApplySubStats is a helper for using ApplyStats with nested muxes. It
// stores the pattern matched in the current mux but does not track
// any metrics independently. It should be included in every mux
// except the outer one.
func ApplySubStats(h goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		patterns, ok := ctx.Value(PATKEY).(*[]goji.Pattern)
		if ok {
			curr := middleware.Pattern(ctx)
			if curr != nil {
				*patterns = append(*patterns, curr)
			}
		}

		h.ServeHTTPC(ctx, w, r)
	})
}
