package main

import (
	"fmt"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
)

type ErrorProducts []ErrorProduct
type ErrorProduct struct {
	products2.ProductInfo
	error
}

func (x ErrorProducts) HasErrors() bool {
	for _, v := range x {
		if v.error != nil {
			return true
		}
	}
	return false
}

func (x ErrorProducts) Error() string {
	m := make(map[string]string)
	for _, v := range x {
		var str string
		if v.error == nil {
			str = "успешно"
		} else {
			str = v.error.Error()
		}
		s, f := m[str]
		if f {
			s += ", "
		}
		s += fmt.Sprintf("%d", v.Addr)
		m[str] = s
	}
	s := ""
	for k, v := range m {
		if s != "" {
			s += ", "
		}
		s += fmt.Sprintf("%s: %s", k, v)
	}
	return s
}
