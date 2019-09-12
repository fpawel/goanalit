package wask6

import (
	"errors"
	"github.com/fpawel/guartutils/modbus"

	"fmt"
	"strconv"
)

type DeviceAddr byte
type ValueAddr byte
type IODir byte

const (
	IODirWrite IODir = 0xA0
	IODirRead  IODir = 0xB0
)

var ErrorAnswerLengthLessThen9 = errors.New("длина ответа менее 9")
var ErrorDeviceAddrMismatch = errors.New("не совпадает адрес платы стенда")
var ErrorValueAddrMismatch = errors.New("не совпадает адрес значения")
var ErrorIODirectionMismatch = errors.New("не совпадает код направления передачи")
var ErrorCrcMismatch = errors.New("не совпадает CRC")
var ErrorWrongBCD = errors.New("не правильный BCD")

type Request struct {
	DeviceAddr DeviceAddr
	ValueAddr  ValueAddr
	Direction  IODir
	Value      float32
}

type Fetcher interface {
	Fetch(request []byte) ([]byte, error)
}

func (x Request) GetResponse(fetcher Fetcher) (float32, error) {
	b, err := fetcher.Fetch(x.Bytes())
	if err != nil {
		return 0, err
	}
	return x.Parse(b)
}

func (x Request) Bytes() (r []byte) {
	r = make([]byte, 9)

	r[0] = byte(x.DeviceAddr)
	r[1] = byte(x.Direction)
	r[2] = byte(x.ValueAddr)
	copy(r[3:7], modbus.BCD6(float64(x.Value)))
	pack(r[1:7])
	r[7], r[8] = crc(r[1:7])
	return
}

func (x Request) String() string {

	var s string
	if x.Direction == IODirRead {
		s = "READ"
	} else {
		s = "WRITE:" + strconv.FormatFloat(float64(x.Value), 'f', -1, 32)
	}
	return fmt.Sprintf("KGS:%d VAR:%d %s", x.DeviceAddr, x.ValueAddr, s)
}

func (x Request) Parse(bb []byte) (float32, error) {
	if len(bb) == 0 {
		return 0, fetch.ErrorNoAnswer
	}
	if len(bb) < 9 {
		return 0, ErrorAnswerLengthLessThen9
	}

	b := make([]byte, 9)
	copy(b, bb[:9])

	c1, c2 := crc(b[1:7])
	if c1 != b[7] || c2 != b[8] {
		return 0, ErrorCrcMismatch
	}
	if b[0] != byte(x.DeviceAddr) {
		return 0, ErrorDeviceAddrMismatch
	}
	if b[1]&0xF0 != byte(x.Direction) {
		return 0, ErrorIODirectionMismatch
	}
	if b[2] != byte(x.ValueAddr) {
		return 0, ErrorValueAddrMismatch
	}

	unpack(b[1:7])
	value, ok := modbus.ParseBCD(b[3:7])

	if !ok {
		return 0, ErrorWrongBCD
	}

	if x.Direction == IODirWrite && value != x.Value {
		return 0, fmt.Errorf("запрос %v, ответ %v", x.Value, value)
	}

	return value, nil
}

func crc(bs []byte) (byte, byte) {
	var a uint16
	for _, b := range bs {
		var b1, b3, b4 byte
		for i := 0; i < 8; i++ {
			if i == 0 {
				b1 = b
			} else {
				b1 <<= 1
			}
			if b1&0x80 != 0 {
				b3 = 1
			} else {
				b3 = 0
			}
			if a&0x8000 == 0x8000 {
				b4 = 1
			} else {
				b4 = 0
			}
			a <<= 1
			if b3 != b4 {
				a ^= 0x1021
			}
		}
	}
	a ^= 0xFFFF
	return byte(a >> 8), byte(a)
}

type npos struct {
	nbit, nbyte byte
}

var nposs = []npos{
	{3, 2},
	{2, 3},
	{1, 4},
	{0, 5},
}

func pack(bs []byte) {
	for _, x := range nposs {
		setBit(x.nbit, getBit(7, bs[x.nbyte]), &bs[0])
		bs[x.nbyte] &= 0x7F
	}
}

func unpack(bs []byte) {
	for _, x := range nposs {
		setBit(7, getBit(x.nbit, bs[0]), &bs[x.nbyte])
		setBit(x.nbit, false, &bs[0])
	}
}

func setBit(pos byte, value bool, b *byte) {
	if value {
		*b |= 1 << pos
	} else {
		*b &= ^(1 << pos)
	}
}

func getBit(pos byte, b byte) bool {
	return b&(1<<pos) != 0
}

func (x IODir) String() string {
	switch x {
	case IODirRead:
		return "считывание"
	case IODirWrite:
		return "запись"
	default:
		panic(x)
	}
}
