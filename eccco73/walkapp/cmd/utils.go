package main

import "time"

var monthNames = []string{
	"",
	"Январь",
	"Февраль",
	"Март",
	"Апрель",
	"Май",
	"Июнь",
	"Июль",
	"Август",
	"Сентябрь",
	"Октябрь",
	"Ноябрь",
	"Декабрь",
}

func monthNumberToName(month time.Month) string {

	if month < 1 || month > 12 {
		panic("month must be number from 1 to 12")
	}
	return monthNames[month]

}
