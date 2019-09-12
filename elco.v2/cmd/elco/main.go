package main

import (
	"github.com/fpawel/elco.v2/internal/app"
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
	_ "runtime/cgo"
)

func main() {

	structlog.DefaultLogger.
		SetPrefixKeys(
			structlog.KeyApp, structlog.KeyPID, structlog.KeyLevel, structlog.KeyUnit, structlog.KeyTime,
		).
		SetDefaultKeyvals(
			structlog.KeyApp, filepath.Base(os.Args[0]),
			structlog.KeySource, structlog.Auto,
		).
		SetSuffixKeys(
			structlog.KeyStack,
		).
		SetSuffixKeys(structlog.KeySource).
		SetKeysFormat(map[string]string{
			structlog.KeyTime:   " %[2]s",
			structlog.KeySource: " %6[2]s",
			structlog.KeyUnit:   " %6[2]s",
			"config":            " %+[2]v",
			"запрос":            " %[1]s=`% [2]X`",
			"ответ":             " %[1]s=`% [2]X`",
			"работа":            " %[1]s=`%[2]s`",
		}).SetTimeFormat("15:04:05")

	data.Open(false)
	defer log.ErrIfFail(data.Close)

	closeHttpServer := startHttpServer()
	defer closeHttpServer()

	app.Run()
}

var (
	cancelComport = func() {}
	skipDelay     = func() {}
	log           = structlog.New()
)
