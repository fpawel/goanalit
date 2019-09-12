package kgsdum

import (
	"bytes"
	"encoding/binary"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"log"
	"path/filepath"
	"runtime"
)

type Value struct {
	Product products2.Product
	Test    Test
	Index   byte
	Var     Var
}

func (x Value) Key() []byte {
	return []byte(x.Var.String())
}

func (x Value) Path() [][]byte {
	return append(x.Product.Test(x.Test.String()).Path(), []byte{x.Index})
}

func (x Value) Value() *float32 {
	v := x.Product.Party.Tx.Value(x.Path(), x.Key())
	buf := bytes.NewReader(v)
	var v32 float32
	if err := binary.Read(buf, binary.LittleEndian, &v32); err == nil {
		return &v32
	}
	return nil
}

func (x Value) SetValue(v float32) {
	buf := new(bytes.Buffer)
	check(binary.Write(buf, binary.LittleEndian, v))
	x.Product.Party.Tx.SetValue(x.Path(), x.Key(), buf.Bytes())
}

func check(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Panicf("%s:%d %v\n", filepath.Base(file), line, err)
	}
}
