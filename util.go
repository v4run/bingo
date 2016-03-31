package bingo

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/fatih/color"
	"github.com/hifx/banner"
	"github.com/hifx/graceful"
	"goji.io/pat"
)

//PrintName prints the app name
func PrintName(str string) {
	color.New(color.FgYellow).Add(color.Bold).Println(banner.PrintS(str))
}

// PrintBanner prints the banner
func PrintBanner(a ...interface{}) {
	hd := color.New(color.FgGreen).Add(color.Bold)
	hd.Println("----------------------------------")
	hd.Println(a...)
	hd.Println("----------------------------------")
}

// PrintError prints the error
func PrintError(a ...interface{}) {
	hd := color.New(color.FgRed).Add(color.Bold)
	hd.Println("----------------------------------")
	hd.Println(a...)
	hd.Println("----------------------------------")
}

//Run gracefully starts the http server
func Run(addr string, timeout time.Duration, h http.Handler) {
	http.Handle("/", h)
	graceful.Run(addr, timeout, http.DefaultServeMux)
}

//BoundParam returns the bound parameter with the given name. Wraps around goji's pat.Param
func BoundParam(ctx context.Context, name string) string {
	return pat.Param(ctx, name)
}
