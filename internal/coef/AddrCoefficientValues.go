package coef

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"sort"
	"strconv"
	"strings"
)

type Coefficient uint16

type AddrCoefficient struct {
	Addr        byte
	Coefficient Coefficient
}

type AddrCoefficientValues map[AddrCoefficient]float32

func (x AddrCoefficient) String() string {
	return fmt.Sprintf("%d:%d", x.Addr, x.Coefficient)
}

func (x AddrCoefficientValues) Addresses() (r []byte) {
	xs := make(map[byte]bool)
	for k := range x {
		xs[k.Addr] = true
	}
	for k := range xs {
		r = append(r, k)
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i] < r[j]
	})

	return
}

func (x AddrCoefficientValues) Coefficients() (r []Coefficient) {
	xs := make(map[Coefficient]struct{})
	for k := range x {
		xs[k.Coefficient] = struct{}{}
	}
	for k := range xs {
		r = append(r, k)
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i] < r[j]
	})
	return
}

const sheet1 = "sheet1"

func (x AddrCoefficientValues) SaveToFile(path string) error {

	xlsx := excelize.NewFile()
	xlsx.SetCellValue(sheet1, "A1", "Адрес")

	for i, addr := range x.Addresses() {
		xlsx.SetCellValue(sheet1, fmt.Sprintf("%s1", excelize.ToAlphaString(i+1)), addr)
	}

	for nk, k := range x.Coefficients() {
		xlsx.SetCellValue(sheet1, fmt.Sprintf("A%d", nk+2), k)
		for np, addr := range x.Addresses() {
			v := x[AddrCoefficient{Addr: addr, Coefficient: k}]
			s := strconv.FormatFloat(float64(v), 'f', -1, 32)
			cell := fmt.Sprintf("%s%d", excelize.ToAlphaString(np+1), nk+2)
			xlsx.SetCellValue(sheet1, cell, s)
		}
	}
	return xlsx.SaveAs(path)
}

func OpenFromFile(path string) (AddrCoefficientValues, error) {
	xlsx, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	x := make(AddrCoefficientValues)

	for col := 0; ; col++ {
		cell := fmt.Sprintf("%s1", excelize.ToAlphaString(col+1))
		s := xlsx.GetCellValue(sheet1, cell)
		if strings.TrimSpace(s) == "" {
			return x, nil
		}
		addr, err := strconv.Atoi(s)
		if err != nil || addr < 1 || addr > 127 {
			return nil, fmt.Errorf("не правильное значение ячейки %s=%q: ожидался адрес MODBUS от 1 до 127", cell, s)
		}

		for row := 0; ; row++ {
			cell := fmt.Sprintf("A%d", row+2)
			s := xlsx.GetCellValue(sheet1, cell)

			if strings.TrimSpace(s) == "" {
				break
			}

			coefficient, err := strconv.Atoi(s)
			if err != nil || coefficient < 0 || coefficient > 0xffff {
				return nil, fmt.Errorf("не правильное значение ячейки %s=%q: ожидался номер коэффициента от 0 до FFFF", cell, s)
			}

			cell = fmt.Sprintf("%s%d", excelize.ToAlphaString(col+1), row+2)
			s = xlsx.GetCellValue(sheet1, cell)
			if strings.TrimSpace(s) == "" {
				continue
			}

			v, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return nil, fmt.Errorf("не правильное значение ячейки %s=%q: ожидалось значение коэффициента", cell, s)
			}
			pc := AddrCoefficient{byte(addr), Coefficient(coefficient)}
			x[pc] = float32(v)
		}
	}

}
