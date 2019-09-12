package main

import (
	"github.com/fpawel/daf.v0/internal/data"
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
			KeyAddr:             " %[2]X",
			KeyEN6408:           " %+[2]v",
			"config":            " %+[2]v",
			"duration":          " %[2]q",
			"запрос":            " %[1]s=`% [2]X`",
			"ответ":             " %[1]s=`% [2]X`",
		}).SetTimeFormat("15:04:05")

	log := structlog.New()
	log.Info("start", structlog.KeyTime, now())

	data.Open()

	if err := runMainWindow(); err != nil {
		panic(err)
	}

}
