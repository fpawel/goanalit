package wask6

import (
	"fmt"
	"testing"
)

func TestCRC(t *testing.T) {
	b1, b2 := crc([]byte{1, 2, 3, 4})
	if b1 != 242 || b2 != 252 {
		t.Error("[1,2,3,4] != 242,252")
	}

	b1, b2 = crc([]byte{1, 2, 3, 4, 100, 101, 102, 103})
	if b1 != 193 || b2 != 92 {
		t.Error("[1,2,3,4,100,101,102,103] != 193, 92")
	}
}

func mustEqual(t *testing.T, a, b []byte) {

	if len(a) != len(b) {
		t.Error(fmt.Errorf("%v, must be %v", a, b))
		return
	}

	for i := range a {
		if a[i] != b[i] {
			t.Errorf("%v, must be %v", a, b)
			return
		}
	}
}

func TestNewRead(t *testing.T) {
	mustEqual(t, Request{
		DeviceAddr: 121,
		ValueAddr:  128,
		Direction:  IODirRead,
	}.Bytes(), []byte{121, 176, 128, 0, 0, 0, 0, 38, 131})
	mustEqual(t,
		Request{
			DeviceAddr: 121,
			ValueAddr:  128,
			Direction:  IODirWrite,
			Value:      987.456,
		}.Bytes(),
		[]byte{121, 164, 128, 3, 24, 116, 86, 181, 22})
}

func TestParse(t *testing.T) {
	value, err := Request{
		DeviceAddr: 2,
		ValueAddr:  62,
		Direction:  IODirRead,
	}.Parse([]byte{0x02, 0xB0, 0x3E, 0x05, 0x75, 0x21, 0x37, 0x3B, 0xCB})
	if value != 7.52137 || err != nil {
		t.Errorf("%v: %v", value, err)
	}
}

//02 B0 3E 05 75 21 37 3B CB
