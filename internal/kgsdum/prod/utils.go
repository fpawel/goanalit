package main

import (
	"github.com/lxn/win"
	"log"
	"path/filepath"
	"runtime"
)

func check(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Panicf("%s:%d %v\n", filepath.Base(file), line, err)
	}
}

func messageFromError(err error, okText string) (string, int) {
	if err == nil {
		return okText, win.NIIF_INFO
	} else {
		return err.Error(), win.NIIF_ERROR
	}
}
