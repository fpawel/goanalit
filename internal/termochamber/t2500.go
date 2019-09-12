package termochamber

import (
	"fmt"
)

type T2500 struct{}

func (T2500) Start(reader responseGetter) error {
	_, err := getResponse(reader, "01WRD,01,0102,0001")
	return err
}

func (T2500) Stop(reader responseGetter) error {
	_, err := getResponse(reader, "01WRD,01,0102,0004")
	return err
}

func (T2500) Setup(reader responseGetter, value float64) error {
	v := int64(value * 10)
	s := fmt.Sprintf("01WRD,01,0104,%04X", v)
	_, err := getResponse(reader, s)
	return err
}

func (T2500) Read(reader responseGetter) (float64, error) {
	return getResponse(reader, "01RRD,02,0001,0002")
}
