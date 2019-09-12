package main

import (
	"log"
)

var buildtime = "(undefined)"
var majorVersion = 0
var minorVersion = 0
var bugfixVersion = 0
var debug = ""
var prod = ""

func main() {

	println("debug:", debug, "runner:", prod)
	log.SetFlags(log.Lshortfile)
	app := NewApp()
	app.mw.Run()
	app.Close()
}
