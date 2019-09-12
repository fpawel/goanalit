package productsView

import (
	"fmt"
	"github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/fpawel/gutils/utils"
	"github.com/lxn/walk"
	"time"
)

type Node interface {
	walk.TreeItem
	ContainsParty(p products.PartyTime) bool
	What() string
}

type NodeYear struct {
	months []*NodeMonth
	year   int
}

type NodeMonth struct {
	parentYearNode *NodeYear
	month          time.Month
	days           []*NodeDay
}

type NodeDay struct {
	parentMonthNode *NodeMonth
	day             int
	parties         []*NodeParty
}

type NodeParty struct {
	parentDayNode *NodeDay
	party         products.PartyInfo
	prodType      string
	products      []*NodeProduct
}

type NodeProduct struct {
	parentPartyNode *NodeParty
	product         products.ProductInfo
	good, bad       bool
	what            string
}

func (x *NodeYear) Year() int { return x.year }
func (x *NodeYear) Text() string {
	return fmt.Sprintf("%d", x.year)
}
func (x *NodeYear) Parent() walk.TreeItem {
	return nil
}
func (x *NodeYear) ChildCount() int {
	return len(x.months)
}
func (x *NodeYear) ChildAt(index int) walk.TreeItem {
	return x.months[index]
}
func (x *NodeYear) What() string {
	return fmt.Sprintf("Каталог %d год", x.year)
}

func (x *NodeMonth) Year() int         { return x.parentYearNode.Year() }
func (x *NodeMonth) Month() time.Month { return x.month }
func (x *NodeMonth) Text() string {
	return fmt.Sprintf("%s", utils.MonthNumberToName(x.month))
}
func (x *NodeMonth) Parent() walk.TreeItem {
	return x.parentYearNode
}
func (x *NodeMonth) ChildCount() int {
	return len(x.days)
}
func (x *NodeMonth) ChildAt(index int) walk.TreeItem {
	return x.days[index]
}
func (x *NodeMonth) What() string {
	return fmt.Sprintf("Каталог %s %d", utils.MonthNumberToName(x.month), x.parentYearNode)
}

func (x *NodeDay) Year() int         { return x.parentMonthNode.Year() }
func (x *NodeDay) Month() time.Month { return x.parentMonthNode.Month() }
func (x *NodeDay) Day() int          { return x.day }
func (x *NodeDay) Text() string {
	return fmt.Sprintf("%d", x.day)
}
func (x *NodeDay) Parent() walk.TreeItem {
	return x.parentMonthNode
}
func (x *NodeDay) ChildCount() int {
	return len(x.parties)
}
func (x *NodeDay) ChildAt(index int) walk.TreeItem {
	return x.parties[index]
}
func (x *NodeDay) What() string {
	return fmt.Sprintf("Каталог %d %s %d", x.day, utils.MonthNumberToName(x.Month()), x.Year())
}

func (x *NodeParty) Year() int         { return x.parentDayNode.Year() }
func (x *NodeParty) Month() time.Month { return x.parentDayNode.Month() }
func (x *NodeParty) Day() int          { return x.parentDayNode.Day() }
func (x *NodeParty) Party() products.PartyInfo {
	return x.party
}
func (x *NodeParty) Text() string {
	return fmt.Sprintf("%s %s", x.party.PartyTime.Time().Format("15:04"), x.prodType)
}
func (x *NodeParty) Parent() walk.TreeItem {
	return x.parentDayNode
}
func (x *NodeParty) ChildCount() int {
	return len(x.products)
}
func (x *NodeParty) ChildAt(index int) walk.TreeItem {
	return x.products[index]
}
func (x *NodeParty) What() (result string) {
	return fmt.Sprintf("Партия %s, %d приборов, %s",
		x.party.PartyTime.Time().Format("2006 01 02 03:04:05"),
		len(x.products), x.prodType)
}

func (x *NodeProduct) Year() int         { return x.parentPartyNode.Year() }
func (x *NodeProduct) Month() time.Month { return x.parentPartyNode.Month() }
func (x *NodeProduct) Day() int          { return x.parentPartyNode.Day() }
func (x *NodeProduct) Party() products.PartyInfo {
	return x.parentPartyNode.Party()
}
func (x *NodeProduct) Product() products.ProductInfo {
	return x.product
}
func (x *NodeProduct) Text() (result string) {
	return x.what
}

func (x *NodeProduct) Parent() walk.TreeItem {
	return x.parentPartyNode
}
func (x *NodeProduct) ChildCount() int {
	return 0
}
func (x *NodeProduct) ChildAt(index int) walk.TreeItem {
	return nil
}
func (x *NodeProduct) What() string {
	return x.Text() + ": " + x.parentPartyNode.What()
}

func (x *NodeYear) Image() interface{} {
	return ImgCalendarYearPng16
}
func (x *NodeMonth) Image() interface{} {
	return ImgCalendarMonthPng16
}
func (x *NodeDay) Image() interface{} {
	return ImgCalendarDayPng16
}
func (x *NodeParty) Image() interface{} {

	for _, p := range x.products {
		if p.bad {
			return ImgErrorPng16
		}
	}

	for _, p := range x.products {
		if !p.good {
			return ImgPartyNodePng16
		}
	}

	return ImgCheckmarkPng16
}
func (x *NodeProduct) Image() interface{} {
	if x.good {
		return ImgCheckmarkPng16
	}
	if x.bad {
		return ImgErrorPng16
	}
	return ImgProductNodePng16
}

func (x *NodeYear) ContainsParty(p products.PartyTime) bool {
	return time.Time(p).Year() == x.year
}
func (x *NodeMonth) ContainsParty(p products.PartyTime) bool {
	t := time.Time(p)
	return t.Year() == x.Year() && t.Month() == x.Month()
}
func (x *NodeDay) ContainsParty(p products.PartyTime) bool {
	t := time.Time(p)
	return t.Year() == x.Year() && t.Month() == x.Month() && t.Day() == x.Day()
}
func (x *NodeParty) ContainsParty(p products.PartyTime) bool {

	return time.Time(p).Equal(time.Time(x.party.PartyTime))
}
func (x *NodeProduct) ContainsParty(p products.PartyTime) bool {
	return false
}

var _ walk.TreeItem = new(NodeYear)
var _ walk.TreeItem = new(NodeMonth)
var _ walk.TreeItem = new(NodeDay)
var _ walk.TreeItem = new(NodeParty)
var _ walk.TreeItem = new(NodeProduct)

var _ Node = new(NodeYear)
var _ Node = new(NodeMonth)
var _ Node = new(NodeDay)
var _ Node = new(NodeParty)
var _ Node = new(NodeProduct)

var ImgCalendarYearPng16 walk.Image

var ImgCalendarMonthPng16 walk.Image
var ImgCalendarDayPng16 walk.Image
var ImgErrorPng16 walk.Image
var ImgPartyNodePng16 walk.Image
var ImgCheckmarkPng16 walk.Image
var ImgProductNodePng16 walk.Image
var ImgWindowIcon walk.Image
