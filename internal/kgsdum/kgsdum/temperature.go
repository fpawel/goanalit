package kgsdum

import "fmt"

type Temperature float64

const (
	Temperature20     Temperature = 20
	Temperature50     Temperature = 50
	TemperatureMinus5 Temperature = -5
)

func (x Temperature) String() string {
	return fmt.Sprintf("%v\"C", float64(x))
}
