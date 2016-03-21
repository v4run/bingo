/*
Package mux is wrapper over goji.Mux

It adds a predefined list of middlewares to muxes and sub-muxes created via the New() and Sub() functions

For e.g. to add 404 and metrics/stats handlers to all sub-muxes initialize this package with

	mux.InitSub(
			middleware.ApplySubStats,
			middleware.Apply404)


To add logger, 404, metrics and 404 handlers to all muxes, use

	mux.Init(
		middleware.ApplyLog(acsslog),
		middleware.Apply404,
		middleware.ApplyStats,
		middleware.ApplyReqID)

*/
package mux

import "goji.io"

var mlist []func(goji.Handler) goji.Handler
var submlist []func(goji.Handler) goji.Handler

/*
New is a wrapper over goji.NewMux(). It adds
the pre-defined list of middlewares to the mux
*/
func New() *goji.Mux {
	m := goji.NewMux()
	for _, mware := range mlist {
		m.UseC(mware)
	}

	return m
}

/*
Sub is a wrapper over goji.SubMux(). It adds
the pre-defined list of middlewares to the submux
*/
func Sub() *goji.Mux {
	m := goji.SubMux()
	for _, mware := range submlist {
		m.UseC(mware)
	}
	return m
}

// Init sets the middlewares to be used for all muxes
// Init is not intended to be used concurrently from multiple goroutines
func Init(m ...func(goji.Handler) goji.Handler) {
	mlist = m
}

// InitSub sets the middlewares to be used for all sub-muxes
// InitSub is not intended to be used concurrently from multiple goroutines
func InitSub(m ...func(goji.Handler) goji.Handler) {
	submlist = m
}
