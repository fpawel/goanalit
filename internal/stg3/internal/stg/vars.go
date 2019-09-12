package stg

import "github.com/fpawel/guartutils/modbus"

type Var struct {
	Var modbus.Var
	Cmd modbus.DeviceCommandCode
}

var Vars = []Var{
	{256, 0},
	{258, 0x0100},
	{2, 0x1000},
	{4, 0x1100},
}
