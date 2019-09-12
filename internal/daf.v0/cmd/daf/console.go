package main

import (
	"syscall"
	"unsafe"
)

type (
	logLevel byte
)

// Log levels.
const (
	DBG logLevel = iota
	INF
	WRN
	ERR
)

var (
	consoleDLL             = syscall.NewLazyDLL("console.dll")
	formConsoleShowProc    = consoleDLL.NewProc("FormConsoleShow")
	formConsoleNewLineProc = consoleDLL.NewProc("FormConsoleNewLine")
)

func formConsoleShow() {
	_, _, _ = formConsoleShowProc.Call()
}

func formConsoleNewLine(lev logLevel, text string) {
	_, _, _ = formConsoleNewLineProc.Call(uintptr(lev), uintptr(utf16StringPtr(text)))
}

func utf16StringPtr(s string) unsafe.Pointer {
	p, err := syscall.UTF16PtrFromString(s)
	if err == syscall.EINVAL {
		p, err = syscall.UTF16PtrFromString("")
	}
	if err != nil {
		panic(err)
	}
	return unsafe.Pointer(p)
}
