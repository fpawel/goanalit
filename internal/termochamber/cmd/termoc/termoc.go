package main

import (
	"flag"
	"fmt"
	"github.com/fpawel/goutils/serial-comm/comm"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/serial-comm/termochamber"
	"github.com/pkg/errors"
	"github.com/tarm/serial"
	"math"
	"time"
)

func main() {
	var (
		help bool
		comportName,
		tcType,
		cmd string
		setup float64
	)
	flag.BoolVar(&help, "?", false, "использование программы")
	flag.StringVar(&comportName, "port", "", "имя компорта")
	flag.StringVar(&tcType, "type", "800", "тип термокамеры - 800 или 2500")
	flag.Float64Var(&setup, "setup", math.MinInt64, "уставка")
	flag.StringVar(&cmd, "cmd", "r", "команда: read - считать температуру, start - старт, stop - стоп")

	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	if err := do(comportName, setup, tcType, cmd); err != nil {
		fmt.Println(err)
	} else {

	}
}

func do(comportName string, setup float64, tcType string, cmd string) error {

	port := comport.NewPort(comport.Config{
		Serial: serial.Config{
			Name:        comportName,
			Baud:        9600,
			ReadTimeout: time.Millisecond,
		},
		Fetch: comm.Config{
			MaxAttemptsRead: 1,
			ReadTimeout:     500 * time.Millisecond,
			ReadByteTimeout: 50 * time.Millisecond,
		},
	})

	if err := port.Open(); err != nil {
		return err
	}

	var tc termochamber.Hardware

	switch tcType {
	case "800":
		tc = termochamber.T800{}
	case "2500":
		tc = termochamber.T800{}
	case "":
		return errors.New("не задан тип термокамеры")
	default:
		return errors.New("не верный тип термокамеры")
	}

	if setup != math.MinInt64 {
		return tc.Setup(port, setup)
	}

	switch cmd {
	case "read":
		v, err := tc.Read(port)
		if err != nil {
			return err
		}
		fmt.Println(v, "\"C")
	case "stop":
		return tc.Stop(port)
	case "start":
		return tc.Start(port)
	default:
		return errors.Errorf("не правильная команда: %q", cmd)
	}
	return nil
}
