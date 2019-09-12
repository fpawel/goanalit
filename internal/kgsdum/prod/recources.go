package main

import "github.com/lxn/walk"

const (
	IconAppID      = 11
	IconSettingsID = 21
	IconDBID       = 25
)

func NewIconFromResourceId(id int) *walk.Icon {
	icon, err := walk.NewIconFromResourceId(id)
	check(err)
	return icon
}
