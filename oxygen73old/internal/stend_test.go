package internal

import (
	"fmt"
	"testing"
)

func TestCRC16(t *testing.T) {
	bs := []byte{1, 2, 3, 4, 5, 67, 8, 9}
	x, y := getCRC16(bs)
	a, b := getCRC16(append(bs, x, y))
	fmt.Println(x, y, a, b)
	if a != 0 || b != 0 {

		t.Error("noy null")
	}
}
