package internal

import (
	"bytes"
	"encoding/binary"
	"github.com/boltdb/bolt"
	"time"
)

type TimeValue struct {
	N byte
	X time.Time
	Y float32
}

func (x *TimeValue) restore(k []byte, v []byte) (err error) {

	buf := bytes.NewReader(k) // ключ - наносекунды даты-времени
	var nanos int64
	if err = binary.Read(buf, binary.LittleEndian, &nanos); err != nil {
		return
	}
	x.X = time.Unix(0, nanos)

	buf = bytes.NewReader(v)
	var valueY float32
	if err = binary.Read(buf, binary.LittleEndian, &valueY); err != nil {
		return
	}
	x.Y = float32(valueY)

	return
}

func (x TimeValue) serializeY(buf *bytes.Buffer) {
	if err := binary.Write(buf, binary.LittleEndian, x.Y); err != nil {
		logger.Panic(err)
	}
}

func (x TimeValue) serializeToSendToChartPipe() []byte {
	buf := new(bytes.Buffer)
	// индекс графика
	if n, err := buf.Write([]byte{x.N}); n != 1 || err != nil {
		logger.Panicln(x.N, err, n)
	}

	// количество миллисекунд метки времени
	var millis int64 = x.X.UnixNano() / 1000000
	if err := binary.Write(buf, binary.LittleEndian, millis); err != nil {
		logger.Panic(err)
	}

	x.serializeY(buf) // число
	return buf.Bytes()
}

func (x TimeValue) store(buck *bolt.Bucket) (err error) {
	bufKey := new(bytes.Buffer)
	bufValue := new(bytes.Buffer)

	var nanos int64 = x.X.UnixNano()
	if err = binary.Write(bufKey, binary.LittleEndian, &nanos); err != nil {
		return
	}

	x.serializeY(bufValue)
	return buck.Put(bufKey.Bytes(), bufValue.Bytes())
}
