package kgsdum

import (
	"fmt"
	coef2 "github.com/fpawel/goanalit/internal/coef"
	"time"
)

type WorkProvider interface {
	SetupGas(gas Gas) error
	SetupTemperature(temperature Temperature) error
	Delay(what string, duration time.Duration) error
	WriteCoefficient(coefficient coef2.Coefficient, value float64) error
	ReadVar(Var) error
	FixTestValue(test Test, varCode Var, n byte) error
	Config() Config
	Print(test Test, args ...interface{})
	Printf(test Test, fmt string, args ...interface{})
}

type Worker interface {
	fmt.Stringer
	Action(WorkProvider) error
}

func Works() (r []Worker) {
	for _, y := range works {
		r = append(r, y)
	}
	return
}

type work struct {
	what   string
	action func(WorkProvider) error
}

func NewWork(what string, action func(WorkProvider) error) Worker {
	return &work{what, action}
}

func (x *work) Action(w WorkProvider) error {
	return x.action(w)
}

func (x *work) String() string {
	return x.what
}

func (x *work) Index() int {
	for i, y := range works {
		if y == x {
			return i
		}
	}
	panic(*x)
}

func (x Test) Index() int {
	for i, y := range works {
		switch y := y.(type) {
		case Test:
			if y == x {
				return i
			}
		}
	}
	panic(x)
}

var CorrectScaleProgram = &work{
	"Калибровка",
	func(w WorkProvider) error {
		w.Print("калибровка", Gas1)

		if err := blowGas(w, Gas1); err != nil {
			return err
		}
		if err := w.ReadVar(100); err != nil {
			return err
		}

		w.Print("калибровка", Gas4)
		if err := blowGas(w, Gas4); err != nil {
			return err
		}
		if err := w.WriteCoefficient(28, w.Config().Gas(Gas4)); err != nil {
			return err
		}
		return w.ReadVar(101)
	},
}

var works = []Worker{
	&work{
		"Термоциклирование",
		func(w WorkProvider) error {
			for i := 0; i < 3; i++ {
				w.Print("термоциклирование", i+1, -60)
				if err := holdTemperature(w, -60); err != nil {
					return err
				}
				w.Print("термоциклирование", i+1, 80)
				if err := holdTemperature(w, 80); err != nil {
					return err
				}
			}
			return holdTemperature(w, 20)
		},
	},

	CorrectScaleProgram,

	&work{
		"Линеаризация",
		func(w WorkProvider) error {

			if err := blowGas(w, Gas2); err != nil {
				return err
			}
			if err := w.WriteCoefficient(44, 1); err != nil {
				return err
			}
			if err := w.WriteCoefficient(48, w.Config().Gas(Gas2)); err != nil {
				return err
			}
			if err := w.ReadVar(102); err != nil {
				return err
			}

			if err := blowGas(w, Gas3); err != nil {
				return err
			}
			if err := w.WriteCoefficient(44, 0); err != nil {
				return err
			}
			if err := w.WriteCoefficient(29, w.Config().Gas(Gas3)); err != nil {
				return err
			}
			if err := w.ReadVar(102); err != nil {
				return err
			}

			if err := blowGas(w, Gas1); err != nil {
				return err
			}
			if err := w.WriteCoefficient(44, 2); err != nil {
				return err
			}
			return w.ReadVar(102)
		},
	},
	&work{
		"Термокомпенсация",
		func(w WorkProvider) error {

			if err := holdTemperature(w, Temperature20); err != nil {
				return err
			}
			if err := blowGas(w, Gas1); err != nil {
				return err
			}
			if err := w.ReadVar(103); err != nil {
				return err
			}
			if err := blowGas(w, Gas4); err != nil {
				return err
			}
			if err := w.ReadVar(107); err != nil {
				return err
			}
			if err := blowGas(w, Gas1); err != nil {
				return err
			}

			if err := holdTemperature(w, TemperatureMinus5); err != nil {
				return err
			}
			if err := blowGas(w, Gas1); err != nil {
				return err
			}
			if err := w.ReadVar(105); err != nil {
				return err
			}
			if err := blowGas(w, Gas4); err != nil {
				return err
			}
			if err := w.ReadVar(109); err != nil {
				return err
			}
			if err := blowGas(w, Gas1); err != nil {
				return err
			}

			if err := holdTemperature(w, Temperature50); err != nil {
				return err
			}
			if err := blowGas(w, Gas1); err != nil {
				return err
			}
			if err := w.ReadVar(104); err != nil {
				return err
			}
			if err := blowGas(w, Gas4); err != nil {
				return err
			}
			if err := w.ReadVar(108); err != nil {
				return err
			}
			if err := blowGas(w, Gas1); err != nil {
				return err
			}

			if err := w.ReadVar(106); err != nil {
				return err
			}
			w.Delay("выдержка БО в течении минуты", time.Minute)
			return w.ReadVar(110)
		},
	},
	Test20,
	TestMinus5,
	Test50,
	Test20Ret,
}

func holdTemperature(w WorkProvider, temperature Temperature) error {
	if err := w.SetupTemperature(temperature); err != nil {
		return err
	}
	return w.Delay(fmt.Sprintf(`выдержка при %v"C`, temperature), w.Config().HoldTemperatureDuration)
}

func blowGas(w WorkProvider, gas Gas) error {
	if err := w.SetupGas(gas); err != nil {
		return err
	}
	return w.Delay(fmt.Sprintf(`продувка ПГС%d`, gas), w.Config().BlowGasDuration)
}
