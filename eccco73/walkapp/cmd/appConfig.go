package main

import (
	"encoding/json"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/fpawel/goutils/serial/fetch"
	"github.com/tarm/serial"
	"io/ioutil"
	"time"

	"fmt"
	"log"
)

type AppConfig struct {
	PortProducts      comport.Config
	UncheckedProducts map[byte]struct{}
	FindProductSerial int64
}

func NewAppConfig() AppConfig {

	portConfig9600 := serial.Config{
		Name:        "COM1",
		ReadTimeout: time.Millisecond,
		Baud:        9600,
	}
	r := AppConfig{
		PortProducts: comport.Config{
			Serial: portConfig9600,
			Fetch: uart.Config{
				MaxAttemptsRead: 1,
				ReadTimeout:     time.Second,
				ReadByteTimeout: 50 * time.Millisecond,
			},
		},
		UncheckedProducts: make(map[byte]struct{}),
		FindProductSerial: 100,
	}

	// считать настройки приложения из сохранённого файла json
	b, err := ioutil.ReadFile(appConfigFileName())
	if err != nil {
		fmt.Print("config.json error:", err)
	} else {
		if err := json.Unmarshal(b, &r); err != nil {
			fmt.Print("config.json content error:", err)
		}
	}
	return r
}

func (x AppConfig) Save() {
	// сохранить конфиг
	b, err := json.MarshalIndent(x, "", "    ")
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(appConfigFileName(), b, 0644)
	if err != nil {
		log.Panic(err)
	}
}
