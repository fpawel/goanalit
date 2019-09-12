package main

import (
	"fmt"
	"github.com/fpawel/eccco73/internal/eccco73"
	"github.com/lxn/walk"
	"strconv"
	"time"
)

type PartiesReader interface {
	GetYears() (xs []int64)
	GetMonthsOfYear(year int64) (xs []int64)
	GetDaysOfYearMonth(year, month int64) (xs []int64)
	GetPartiesOfYearMonthDay(year, month, day int64) (xs []eccco73.Party1)
}

type PartiesTreeViewModel struct {
	walk.TreeModelBase
	partiesReader PartiesReader
	party         *eccco73.Party
	years         []*PartiesYearNode
}

type PartiesYearNode struct {
	year   int64
	m      *PartiesTreeViewModel
	months []*PartiesMonthNode
}

type PartiesMonthNode struct {
	nodeYear *PartiesYearNode
	month    int64
	days     []*PartiesDayNode
}

type PartiesDayNode struct {
	nodeMonth *PartiesMonthNode
	day       int64
	parties   []*PartiesPartyNode
}

type PartiesPartyNode struct {
	nodeDay *PartiesDayNode
	party   eccco73.Party1
}

func (x *PartiesTreeViewModel) LazyPopulation() bool {
	return true
}

func (x *PartiesTreeViewModel) CurrentParty() *PartiesPartyNode {
	for _, y := range x.years {
		for _, m := range y.months {
			for _, d := range m.days {
				for _, p := range d.parties {
					if p.IsCurrentParty() {
						return p
					}
				}
			}
		}
	}
	return nil
}

func (x *PartiesTreeViewModel) PublishChanged() {
	for _, y := range x.years {
		x.PublishItemChanged(y)
		for _, m := range y.months {
			x.PublishItemChanged(m)
			for _, d := range m.days {
				x.PublishItemChanged(d)
				for _, p := range d.parties {
					x.PublishItemChanged(p)
				}
			}
		}
	}
}

func (x *PartiesTreeViewModel) RootCount() int {
	if len(x.years) == 0 {
		for _, y := range x.partiesReader.GetYears() {
			x.years = append(x.years, &PartiesYearNode{m: x, year: y})
		}
	}
	return len(x.years)
}

func (x *PartiesTreeViewModel) RootAt(i int) walk.TreeItem {
	return x.years[i]
}

func (x *PartiesYearNode) ChildCount() int {
	if len(x.months) == 0 {
		for _, v := range x.m.partiesReader.GetMonthsOfYear(x.year) {
			x.months = append(x.months, &PartiesMonthNode{nodeYear: x, month: v})
		}
	}
	return len(x.months)
}

func (x *PartiesYearNode) ChildAt(index int) walk.TreeItem {
	return x.months[index]
}

func (x *PartiesYearNode) Text() string {
	return strconv.Itoa(int(x.year))
}

func (x *PartiesYearNode) Parent() walk.TreeItem {
	return nil
}

func (x *PartiesMonthNode) ChildCount() int {
	if len(x.days) == 0 {
		for _, v := range x.nodeYear.m.partiesReader.GetDaysOfYearMonth(x.nodeYear.year, x.month) {
			x.days = append(x.days, &PartiesDayNode{nodeMonth: x, day: v})
		}
	}
	return len(x.days)
}

func (x *PartiesMonthNode) ChildAt(index int) walk.TreeItem {
	return x.days[index]
}

func (x *PartiesMonthNode) Text() string {
	return monthNumberToName(time.Month(x.month))
}

func (x *PartiesMonthNode) Parent() walk.TreeItem {
	return x.nodeYear
}

func (x *PartiesDayNode) ChildCount() int {
	if len(x.parties) == 0 {
		for _, v := range x.nodeMonth.nodeYear.m.partiesReader.GetPartiesOfYearMonthDay(x.nodeMonth.nodeYear.year, x.nodeMonth.month, x.day) {
			x.parties = append(x.parties, &PartiesPartyNode{nodeDay: x, party: v})
		}
	}
	return len(x.parties)
}

func (x *PartiesDayNode) ChildAt(index int) walk.TreeItem {
	return x.parties[index]
}

func (x *PartiesDayNode) Text() string {
	return fmt.Sprintf("%02d", x.day)
}

func (x *PartiesDayNode) Parent() walk.TreeItem {
	return x.nodeMonth
}

func (x *PartiesPartyNode) ChildCount() int {
	return 0
}

func (x *PartiesPartyNode) ChildAt(index int) walk.TreeItem {
	return nil
}

func (x *PartiesPartyNode) Text() string {
	s := ""
	if x.party.Note.Valid {
		s = ", " + x.party.Note.String
	}
	return fmt.Sprintf("%s%s", x.party.ProductTypeName, s)
}

func (x *PartiesPartyNode) Parent() walk.TreeItem {
	return x.nodeDay
}

func (x *PartiesDayNode) HasCurrentParty() bool {
	t := x.nodeMonth.nodeYear.m.party.CreatedAt
	return x.nodeMonth.HasCurrentParty() && t.Day() == int(x.day)
}

func (x *PartiesMonthNode) HasCurrentParty() bool {
	t := x.nodeYear.m.party.CreatedAt
	return x.nodeYear.HasCurrentParty() && t.Month() == time.Month(x.month)
}

func (x *PartiesYearNode) HasCurrentParty() bool {
	t := x.m.party.CreatedAt
	return t.Year() == int(x.year)
}

func (x *PartiesPartyNode) IsCurrentParty() bool {
	return x.nodeDay.nodeMonth.nodeYear.m.party.PartyID == x.party.PartyID
}

func (x *PartiesPartyNode) Image() interface{} {
	if x.IsCurrentParty() {
		return "assets/png16/checkmark.png"
	}
	return "assets/png16/folder2.png"
}

func (x *PartiesDayNode) Image() interface{} {
	if x.HasCurrentParty() {
		return "assets/png16/forward.png"
	}
	return "assets/png16/calendar-day.png"
}

func (x *PartiesMonthNode) Image() interface{} {
	if x.HasCurrentParty() {
		return "assets/png16/forward.png"
	}
	return "assets/png16/calendar-month.png"
}

func (x *PartiesYearNode) Image() interface{} {
	if x.HasCurrentParty() {
		return "assets/png16/forward.png"
	}
	return "assets/png16/calendar-year.png"
}
