package wask6

import (
	"fmt"
	coef2 "github.com/fpawel/goanalit/internal/coef"
)

func fetch(r Request, fetcher Fetcher) (float32, error) {
	b, err := Fetch(r.Bytes())
	var value float32
	if err == nil {
		value, err = r.Parse(b)
	}
	if err != nil {
		s := ""
		if b != nil {
			s = fmt.Sprintf(" -> % X", b)
		}
		fmt.Printf("%v: %v % X%s\n", r, err, r.Bytes(), s)
	}
	return value, err
}

func ReadVar(deviceAddr DeviceAddr, valueAddr ValueAddr, fetcher Fetcher) (float32, error) {
	return fetch(Request{
		DeviceAddr: deviceAddr,
		Direction:  IODirRead,
		ValueAddr:  valueAddr,
	}, fetcher)
}

func SendCommand(deviceAddr DeviceAddr, valueAddr ValueAddr, value float32, fetcher Fetcher) error {
	_, err := fetch(Request{
		DeviceAddr: deviceAddr,
		Direction:  IODirWrite,
		ValueAddr:  valueAddr,
		Value:      value,
	}, fetcher)
	return err
}

func ReadCoefficient(pc coef2.AddrCoefficient, fetcher Fetcher) (float32, error) {

	k := (uint16(pc.Coefficient) / 60) * 60
	addr := DeviceAddr(pc.Addr)
	if _, err := fetch(Request{
		DeviceAddr: addr,
		Direction:  IODirRead,
		ValueAddr:  97,
		Value:      float32(k),
	}, fetcher); err != nil {
		return 0, err
	}
	return ReadVar(addr, ValueAddr(uint16(pc.Coefficient)-k), fetcher)
}

func WriteCoefficient(value float32, pc coef2.AddrCoefficient, fetcher Fetcher) (result error) {

	k := (uint16(pc.Coefficient) / 60) * 60
	addr := DeviceAddr(pc.Addr)
	if _, err := fetch(Request{
		DeviceAddr: addr,
		Direction:  IODirRead,
		ValueAddr:  97,
		Value:      float32(k),
	}, fetcher); err != nil {
		return err
	}

	if _, err := fetch(Request{
		DeviceAddr: addr,
		Direction:  IODirWrite,
		ValueAddr:  ValueAddr(uint16(pc.Coefficient) - k),
		Value:      value,
	}, fetcher); err != nil {
		return err
	}

	v, err := ReadCoefficient(pc, fetcher)
	if err != nil {
		return err
	}

	if v != value {
		return fmt.Errorf("KGS:%d, запись коэффициента %d: записанное значение %v не равно считанному занчению %v",
			pc.Addr, pc.Coefficient, value, v)
	}

	return nil
}
