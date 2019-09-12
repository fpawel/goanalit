package internal

import (
	"encoding/binary"
	"log"
	"os"
	"time"

	"github.com/tarm/serial"

	"fmt"
	"github.com/lxn/walk"
)

var logger = log.New(os.Stdout, "OXYGEN|", log.LstdFlags|log.Lshortfile)

var monthNames []string = []string{
	"",
	"Январь",
	"Февраль",
	"Март",
	"Апрель",
	"Май",
	"Июнь",
	"Июль",
	"Август",
	"Сентябрь",
	"Октябрь",
	"Ноябрь",
	"Декабрь",
}

func formatBytesHex(b []byte) (s string) {
	for _, x := range b {
		s += fmt.Sprintf("%02X ", x)
	}
	return
}

func monthNumerToName(month time.Month) string {
	if month < 1 || month > 12 {
		logger.Panic("month must be number from 1 to 12")
	}
	return monthNames[month]

}

// itob returns an 8-byte big endian representation of v.
func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(xs []byte) int64 {
	return int64(binary.BigEndian.Uint64(xs))
}

func getSerialPortNames() (serials []string) {
	for i := 1; i < 100; i++ {
		s := fmt.Sprintf("COM%d", i)
		port, err := serial.OpenPort(&serial.Config{Name: s, Baud: 9600})
		if err == nil {
			err = port.Close()
			if err != nil {
				logger.Panic(err)
			}
			serials = append(serials, s)

		}
	}
	return
}

func setLabelColorText(label *walk.Label, text string, font *walk.Font, color walk.Color) {
	label.Synchronize(func() {
		label.SetText(text)
		label.SetFont(font)
		b, err := walk.NewSolidColorBrush(color)
		if err != nil {
			logger.Panic(err)
		}
		label.SetBackground(b)
	})
}
