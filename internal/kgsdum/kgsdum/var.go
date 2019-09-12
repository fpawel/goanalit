package kgsdum

import "strconv"

type Var byte

const (
	Var0  Var = 0
	Var1  Var = 1
	Var3  Var = 3
	Var12 Var = 12
)

const (
	VarChanel1 Var = 60 + iota
	VarChanel2
	VarPressure
	VarTemperature
)

type varInfo struct {
	Var
	What string
}

func (x Var) String() string {
	if x.index() == -1 {
		return strconv.Itoa(int(x))
	}
	return vars[x.index()].What
}

func (x Var) index() int {
	for n, y := range vars {
		if y.Var == x {
			return n
		}
	}
	return -1
}

func TestVars() (r []Var) {
	return []Var{Var0, Var1, Var3, Var12}
}

func Vars() (r []Var) {
	for _, x := range vars {
		r = append(r, x.Var)
	}
	return
}

var vars = []varInfo{
	{VarChanel1, "K1"},
	{VarChanel2, "K2"},
	{VarPressure, "P"},
	{VarTemperature, "T\"C"},
}
