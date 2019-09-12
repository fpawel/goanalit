package kgsdum

type Test string

const (
	Test20     Test = "Проверка НКУ"
	TestMinus5 Test = "Проверка -5\"C"
	Test50     Test = "Проверка 50\"C"
	Test20Ret  Test = "Проверка возврат НКУ"
)

func (x Test) String() string {
	return string(x)
}

func (x Test) Temperature() Temperature {
	switch x {
	case Test20, Test20Ret:
		return Temperature20
	case TestMinus5:
		return TemperatureMinus5
	case Test50:
		return Temperature50
	default:
		return Temperature20
	}
}

func Tests() []Test {
	r := make([]Test, len(tests))
	copy(r, tests)
	return r
}

var tests = []Test{
	Test20,
	TestMinus5,
	Test50,
	Test20Ret,
}

func (x Test) Action(w WorkProvider) error {

	Print("перевод термокамеры для проверки БО", x, x.Temperature())

	if err := holdTemperature(w, x.Temperature()); err != nil {
		return err
	}
	if x == Test20 {
		Print("калибровка для проверки БО", x, x.Temperature())
		if err := CorrectScaleProgram.Action(w); err != nil {
			return err
		}
	}

	for i, gas := range TestGases() {
		Print("проверка БО", x.Temperature(), "№", i, gas)
		if err := blowGas(w, gas); err != nil {
			return err
		}
		for n, varCode := range []Var{12, 0, 1, 3} {
			if err := FixTestValue(x, varCode, byte(n)); err != nil {
				return err
			}
		}
	}
	return nil
}
