package main

import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	_ "runtime/cgo"
)

var buildTime = "(undefined)"
var majorVersion = 0
var minorVersion = 0
var bugFixVersion = 0

func main() {
	log.SetFlags(log.Lshortfile)
	app := NewApp()
	app.mw.Run()
	app.Close()
}
