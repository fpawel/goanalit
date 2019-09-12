package procmq

import (
	"encoding/binary"
	"encoding/json"
	"github.com/fpawel/goutils"
	"github.com/pkg/errors"
	"math"
	"net"
	"time"
)

type Conn struct {
	net.Conn
}

var errorReadFailed = errors.New("ошибка считывания")
var errorWriteFailed = errors.New("ошибка записи")

func (x Conn) readBytes(count int) ([]byte, error) {
	b := make([]byte, count)

	for pos := 0; pos < count; {
		n, err := x.Read(b[pos:])
		if err != nil {
			return nil, errors.Wrap(errorReadFailed, err.Error())
		}
		if n == 0 {
			return nil, errorReadFailed
		}
		pos += n
	}
	return b, nil
}

func (x Conn) writeBytes(b []byte) error {
	var n int
	n, err := x.Write(b)
	if err != nil {
		return errors.Wrap(errorWriteFailed, err.Error())
	}
	if n != len(b) {
		return errors.Wrapf(errorWriteFailed, "передано %d из %d", n, len(b))
	}
	return nil
}

func (x Conn) WriteBytes(b []byte) error {

	if err := x.WriteUInt32(uint32(len(b))); err != nil {
		return errors.Wrapf(err, "WriteBytes: WriteUInt32(len):")
	}
	if len(b) == 0 {
		return nil
	}

	return x.writeBytes(b)
}

func (x Conn) WriteString(s string) error {

	//b, err := utf8ToWindows1251([]byte(s))
	//if err != nil {
	//	panic(errors.Wrapf(err, "WriteString"))
	//}
	b := goutils.UTF16FromString(s)
	return x.WriteBytes(b)
}

func (x Conn) ReadString() (string, error) {

	n, err := x.ReadUInt32()
	if err != nil {
		return "", errors.Wrap(err, "ReadString: read len")
	}
	if n == 0 {
		return "", nil
	}

	b, err := x.readBytes(int(n))
	if err != nil {
		return "", errors.Wrapf(err, "ReadString: read %d bytes", n)
	}

	bt, err := goutils.UTF8FromUTF16(b)
	if err != nil {
		return "", errors.Wrapf(err, "ReadString: read %d bytes: decode", n)
	}
	return string(bt), nil
}

func (x Conn) ReadBytes() ([]byte, error) {

	n, err := x.ReadUInt32()
	if err != nil {
		return nil, errors.Wrap(err, "ReadBytes: read len")
	}
	if n == 0 {
		return nil, nil
	}

	return x.readBytes(int(n))
}

func (x Conn) WriteUInt32(v uint32) error {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return x.writeBytes(b)
}

func (x Conn) ReadUInt32() (uint32, error) {
	b, err := x.readBytes(4)
	if err != nil {
		return 0, errors.Wrap(err, "ReadUInt32: read 4 bytes")
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (x Conn) WriteUInt64(v uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return x.writeBytes(b)
}

func (x Conn) ReadUInt64() (uint64, error) {
	b, err := x.readBytes(8)
	if err != nil {
		return 0, errors.Wrap(err, "ReadUInt64: read 8 bytes")
	}
	return binary.LittleEndian.Uint64(b), nil
}

func (x Conn) WriteFloat32(v float32) error {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(v))
	return x.writeBytes(b)
}

func (x Conn) ReadFloat32() (float32, error) {
	b, err := x.readBytes(4)
	if err != nil {
		return 0, errors.Wrap(err, "ReadFloat32: read 4 bytes")
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(b)), nil
}

func (x Conn) WriteFloat64(v float64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
	return x.writeBytes(b)
}

func (x Conn) ReadFloat64() (float64, error) {
	b, err := x.readBytes(8)
	if err != nil {
		return 0, errors.Wrap(err, "ReadFloat64: read 8 bytes")
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b)), nil
}

func (x Conn) WriteTime(t time.Time) error {
	return x.WriteUInt64(uint64(t.UnixNano() / 1000000))
}

func (x Conn) WriteJSON(v interface{}) error {

	s, isStr := v.(string)
	if !isStr {
		var err error
		b, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		s = string(b)
	}
	return x.WriteString(s)
}

func (x Conn) WriteMsgJSON(msg string, v interface{}) error {

	if err := x.WriteUInt32(0x5555); err != nil {
		return err
	}

	if err := x.WriteString(msg); err != nil {
		return err
	}

	return x.WriteJSON(v)
}

func (x Conn) ReadMsgJSON() (msg string, b []byte, err error) {

	var tmp uint32
	tmp, err = x.ReadUInt32()
	if err != nil {
		err = errors.Wrap(err, "read 0x5555")
		return
	}
	if tmp != 0x5555 {
		err = errors.Errorf("read 0x5555: %d", tmp)
		return
	}
	msg, err = x.ReadString()
	if err != nil {
		err = errors.Wrap(err, "read msg")
		return
	}
	var n uint32
	n, err = x.ReadUInt32()
	if err != nil {
		err = errors.Wrapf(err, "msg=%q: read json str len", msg)
		return
	}
	if n == 0 {
		return
	}

	var bs []byte
	bs, err = x.readBytes(int(n))
	if err != nil {
		err = errors.Wrapf(err, "msg=%q: read json str", msg)
		return
	}

	b, err = goutils.UTF8FromUTF16(bs)
	if err != nil {
		err = errors.Wrapf(err, "msg=%q: str=%q: DecodeUTF16", msg, string(bs))
		return
	}

	return
}
