package main

import (
	kgsdum2 "github.com/fpawel/goanalit/internal/kgsdum/kgsdum"
	"github.com/fpawel/gutils/walkUtils"
)

type ProductInfo struct {
	values     map[kgsdum2.Var]Float32Result
	connection *walkUtils.Message
}

type Float32Result struct {
	Value float64
	Error error
}
