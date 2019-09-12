package main

import (
	"encoding/json"
	kgsdum2 "github.com/fpawel/goanalit/internal/kgsdum/kgsdum"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/gutils/utils"
	"io/ioutil"
	"os/user"
	"sort"
	"time"
)

type SerialPortsConfig struct {
	PortProducts, PortGas, PortTemperature comport.Config
}

type AppConfig struct {
	SerialPorts       SerialPortsConfig
	UncheckedProducts map[byte]struct{}
	UncheckedWorks    map[string]struct{}
	VarsStr, CoefsStr string
	SaveReportsDir    string
	UserInput         struct {
		SendCommand struct {
			Code, Value float64
		}
	}
	Work kgsdum2.Config
}

type ComportType int

const (
	PortProducts ComportType = iota
	PortGas
	PortTemperature
)

func (x SerialPortsConfig) Port(t ComportType) comport.Config {
	switch t {
	case PortProducts:
		return x.PortProducts
	case PortGas:
		return x.PortGas
	case PortTemperature:
		return x.PortTemperature
	default:
		panic("unexpected")
	}
}

func NewAppConfig() AppConfig {
	u, err := user.Current()
	check(err)

	portConfig9600 := comport.Config{
		Name:        "COM1",
		ReadTimeout: time.Millisecond,
		Baud:        9600,
	}
	return AppConfig{
		Work: kgsdum2.DefaultConfig(),
		SerialPorts: SerialPortsConfig{
			PortProducts: comport.Config{
				Port: portConfig9600,
				Mode: fetch.Config{
					MaxAttemptsRead: 1,
					ReadTimeout:     time.Second,
					ReadByteTimeout: 50 * time.Millisecond,
				},
			},
			PortGas: comport.Config{
				Port: portConfig9600,
				Mode: fetch.Config{
					MaxAttemptsRead: 1,
					ReadTimeout:     500 * time.Millisecond,
					ReadByteTimeout: 50 * time.Millisecond,
				},
			},
			PortTemperature: comport.Config{
				Port: portConfig9600,
				Mode: fetch.Config{
					MaxAttemptsRead: 1,
					ReadTimeout:     500 * time.Millisecond,
					ReadByteTimeout: 50 * time.Millisecond,
				},
			},
		},
		CoefsStr:          "1-100",
		VarsStr:           "60-63",
		UncheckedProducts: make(map[byte]struct{}),
		UncheckedWorks:    make(map[string]struct{}),
		SaveReportsDir:    u.HomeDir,
	}
}

func (x AppConfig) Save(path string) {
	// сохранить конфиг
	b, err := json.MarshalIndent(x, "", "    ")
	check(err)
	check(ioutil.WriteFile(path, b, 0644))
}

func (x AppConfig) Vars() (r []kgsdum2.Var) {
	xs := utils.ParseIntRanges(x.VarsStr)
	sort.Ints(xs)
	for _, v := range xs {
		r = append(r, kgsdum2.Var(v))
	}
	return
}
