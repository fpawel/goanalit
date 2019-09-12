package main

import (
	"encoding/binary"
	"github.com/fpawel/goanalit/internal/utils"
	"net"
	"os"
)

type ClientPipeWriter struct {
	net.Conn
	level int
}

func (x ClientPipeWriter) Write(b []byte) (n int, err error) {
	defer func() {
		if err != nil {
			n, err = os.Stdout.Write(b)
		}
	}()

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(x.level))
	n, err = x.Conn.Write(bs)
	if err != nil {
		return
	}
	bt, err := utils.Utf8ToWindows1251(b)
	if err != nil {
		return
	}
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(bt)))
	_, err = x.Conn.Write(bs)
	if err != nil {
		return
	}
	_, err = x.Conn.Write(bt)
	if err != nil {
		return
	}
	return
}
