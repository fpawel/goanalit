package main

import (
	"fmt"
	"testing"
	"time"
)

func TestTest(t *testing.T) {

	t1 := time.Now()
	layout := "02.01.2006.15:04:05.000MST"

	s1 := t1.Format(layout)
	t2, err := time.Parse(layout, s1)

	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	fmt.Println(t1)
	fmt.Println(t2)
	fmt.Println(s1)
	fmt.Println(t2.Format(layout))

}
