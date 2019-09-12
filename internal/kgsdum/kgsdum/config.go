package kgsdum

import "time"

type Config struct {
	Gas1, Gas2, Gas3, Gas4  float64
	BlowGasDuration         time.Duration
	HoldTemperatureDuration time.Duration
}

func DefaultConfig() Config {
	return Config{
		BlowGasDuration:         5 * time.Minute,
		HoldTemperatureDuration: 3 * time.Hour,
	}
}

func (x Config) Gas(gas Gas) float64 {
	switch gas {
	case Gas1:
		return x.Gas1
	case Gas2:
		return x.Gas2
	case Gas3:
		return x.Gas3
	case Gas4:
		return x.Gas4
	default:
		panic(gas)
	}
}
