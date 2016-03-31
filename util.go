package bingo

import (
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/hifx/banner"
	"github.com/hifx/graceful"
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

func Run(addr string, timeout time.Duration, h http.Handler) {
	graceful.Run(addr, timeout, h)
}
