package internal

import (
	"fmt"
	"github.com/lxn/walk"
	"time"
)

const (
	tnodeYear = 0 + iota
	tnodeMonth
	tnodeDay
	tnodeParty
	tnodeActiveParty
)

type PartiesNode struct {
	parent   walk.TreeItem
	text     string
	children []walk.TreeItem

	party     *Chart
	day, year int
	month     time.Month
	nodeType  int
}

type PartiesTreeModel struct {
	walk.TreeModelBase
	years    []*PartiesNode
	selected *PartiesNode
}

var _ walk.TreeModel = new(PartiesTreeModel)
var _ walk.TreeItem = new(PartiesNode)

func NewPartiesTreeViewModel(parties []*Chart, selectedPartyID time.Time,
	getPartyUpdatedAt func(p *Chart) (t time.Time)) (x *PartiesTreeModel) {

	x = new(PartiesTreeModel)
	tree := make(map[int]map[time.Month]map[int][]*Chart)
	for _, p := range parties {
		t := p.CreatedAt
		y := t.Year()
		m := t.Month() //monthNumerToName()
		d := t.Day()

		if _, ok := tree[y]; !ok {
			tree[y] = make(map[time.Month]map[int][]*Chart)
		}

		if _, ok := tree[y][m]; !ok {
			tree[y][m] = make(map[int][]*Chart)
		}

		tree[y][m][d] = append(tree[y][m][d], p)
	}

	for year, months := range tree {
		nodeYear := &PartiesNode{
			text:     fmt.Sprintf("%d", year),
			parent:   nil,
			nodeType: tnodeYear,
			year:     year,
		}

		x.years = append(x.years, nodeYear)

		for month, days := range months {
			nodeMonth := &PartiesNode{
				text:     monthNumerToName(month),
				parent:   nodeYear,
				nodeType: tnodeMonth,
				year:     year,
				month:    month,
			}

			nodeYear.children = append(nodeYear.children, nodeMonth)

			for day, parties := range days {
				nodeDay := &PartiesNode{
					text:     fmt.Sprintf("%d", day),
					parent:   nodeMonth,
					nodeType: tnodeDay,
					year:     year,
					month:    month,
					day:      day,
				}

				nodeMonth.children = append(nodeMonth.children, nodeDay)

				for _, party := range parties {

					updatedAt := getPartyUpdatedAt(party)
					text := party.CreatedAt.Format("15:04")

					if updatedAt != party.CreatedAt {
						text += " - "
						if party.CreatedAt.Day() != updatedAt.Day() ||
							party.CreatedAt.Month() != updatedAt.Month() ||
							party.CreatedAt.Year() != updatedAt.Year() {
							text += updatedAt.Format("02 Jan 2006 15:04")
						} else {
							text += updatedAt.Format("15:04")
						}

					}

					nodeParty := &PartiesNode{
						text:     text,
						parent:   nodeDay,
						party:    party,
						nodeType: tnodeParty,
					}
					if party.CreatedAt == selectedPartyID {
						x.selected = nodeParty
						nodeParty.nodeType = tnodeActiveParty
					}

					nodeDay.children = append(nodeDay.children, nodeParty)
				}

			}
		}
	}
	return x
}

func (x *PartiesTreeModel) Open() {

}

func (m *PartiesTreeModel) RootCount() int {
	return len(m.years)
}

func (m *PartiesTreeModel) RootAt(index int) walk.TreeItem {
	return m.years[index]
}

func (x *PartiesNode) Text() string {
	return x.text
}

func (x *PartiesNode) Parent() walk.TreeItem {
	return x.parent
}

func (x *PartiesNode) ChildCount() int {
	return len(x.children)
}

func (x *PartiesNode) ChildAt(index int) walk.TreeItem {
	return x.children[index]
}

func (x *PartiesNode) Image() interface{} {
	switch x.nodeType {
	case tnodeYear:
		return "./img/calendar-year.png"
	case tnodeMonth:
		return "./img/calendar-month.png"
	case tnodeDay:
		return "./img/calendar-day.png"
	case tnodeParty:
		return "./img/charts-party.png"
	case tnodeActiveParty:
		return "./img/charts-active-party.png"
	default:
		logger.Panic("bad node type")
		return nil
	}

}
