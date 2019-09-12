package termochamber

import (
	"fmt"
)

type T800 struct{}

func (T800) Start(reader responseGetter) error {
	return T800Start(reader)
}

func (T800) Stop(reader responseGetter) error {
	return T800Stop(reader)
}

func (T800) Setup(reader responseGetter, value float64) error {
	return T800Setup(reader, value)
}

func (T800) Read(reader responseGetter) (float64, error) {
	return T800Read(reader)
}

func T800Start(reader responseGetter) error {
	_, err := getResponse(reader, "01WRD,01,0101,0001")
	return err
}

func T800Stop(reader responseGetter) error {
	_, err := getResponse(reader, "01WRD,01,0101,0004")
	return err
}

func T800Setup(reader responseGetter, value float64) error {
	v := int64(value * 10)
	s := fmt.Sprintf("01WRD,01,0102,%04X", v)
	_, err := getResponse(reader, s)
	return err
}

func T800Read(reader responseGetter) (float64, error) {
	return getResponse(reader, "01RRD,02,0001,0002")
}
