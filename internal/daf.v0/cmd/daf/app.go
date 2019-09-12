package main

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/daf.v0/internal/viewmodel"
	"github.com/lxn/walk"
	"github.com/powerman/structlog"
	"time"
)

var (
	dafMainWindow DafMainWindow
	prodsMdl      *viewmodel.DafProductsTable

	log = structlog.New()

	cancelWorkFunc = func() {}
	skipDelay      = func() {}
	ctxApp         = context.TODO()

	portDaf = comport.NewReadWriter(func() comport.Config {
		portName, f := walk.App().Settings().Get("COMPORT_PRODUCTS")
		if !f {
			portName = "COM1"
		}
		return comport.Config{Baud: 9600, Name: portName}
	}, func() comm.Config {
		return comm.Config{
			ReadByteTimeoutMillis: 50,
			ReadTimeoutMillis:     1000,
			MaxAttemptsRead:       2,
		}
	})

	portHart = comport.NewReadWriter(func() comport.Config {
		portName, f := walk.App().Settings().Get("COMPORT_HART")
		if !f {
			portName = "COM1"
		}
		return comport.Config{
			Name:        portName,
			Baud:        1200,
			ReadTimeout: time.Millisecond,
			Parity:      comport.ParityOdd,
			StopBits:    comport.Stop1,
		}
	}, func() comm.Config {
		return comm.Config{
			ReadByteTimeoutMillis: 50,
			ReadTimeoutMillis:     2000,
			MaxAttemptsRead:       5,
		}
	})

	ErrNoOkProducts = merry.New("отстутсвуют приборы, которые отмеченны галочками и не имеют ошибок связи")

	currentWorkName = ""
)
