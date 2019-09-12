package productsView

import (
	"fmt"
	"github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/lxn/walk"
	"sort"
	"time"
)

type TreeModel struct {
	walk.TreeModelBase
	years []*NodeYear
}

type DeviceInfoProvider struct {
	GoodProduct       func(products.Product) bool
	BadProduct        func(products.Product) bool
	FormatProductType func(int) string
}

func NewPartiesTreeViewModel(db products.DB, m DeviceInfoProvider) (x *TreeModel) {

	db.View(func(tx products.Tx) {
		x = &TreeModel{}
		tree := make(map[int]map[time.Month]map[int][]products.Party)
		for _, p := range tx.Parties() {
			t := time.Time(p.PartyTime)
			y := t.Year()
			m := t.Month() //monthNumerToName()
			d := t.Day()

			if _, ok := tree[y]; !ok {
				tree[y] = make(map[time.Month]map[int][]products.Party)
			}

			if _, ok := tree[y][m]; !ok {
				tree[y][m] = make(map[int][]products.Party)
			}

			tree[y][m][d] = append(tree[y][m][d], p)
		}

		for year, months := range tree {
			nodeYear := &NodeYear{
				year: year,
			}

			x.years = append(x.years, nodeYear)

			for month, days := range months {
				nodeMonth := &NodeMonth{
					parentYearNode: nodeYear,
					month:          month,
				}

				nodeYear.months = append(nodeYear.months, nodeMonth)

				for day, parties := range days {
					nodeDay := &NodeDay{
						parentMonthNode: nodeMonth,
						day:             day,
					}

					nodeMonth.days = append(nodeMonth.days, nodeDay)

					for _, party := range parties {

						nodeParty := &NodeParty{
							parentDayNode: nodeDay,
							party:         party.Info(),
							prodType:      m.FormatProductType(party.ProductTypeIndex()),
						}
						nodeDay.parties = append(nodeDay.parties, nodeParty)

						for _, p := range party.Products() {
							nodeProduct := &NodeProduct{
								parentPartyNode: nodeParty,
								product:         p.Info(),
								what:            fmt.Sprintf("%d: %d", p.Addr(), p.Serial()),
								good:            m.GoodProduct(p),
								bad:             m.BadProduct(p),
							}
							nodeParty.products = append(nodeParty.products, nodeProduct)
						}
					}
				}
			}
		}
	})

	sort.Slice(x.years, func(i, j int) bool {
		return x.years[i].year < x.years[j].year
	})

	for _, nodeYear := range x.years {

		sort.Slice(nodeYear.months, func(i, j int) bool {
			return nodeYear.months[i].month < nodeYear.months[j].month
		})

		for _, nodeMonth := range nodeYear.months {
			sort.Slice(nodeMonth.days, func(i, j int) bool {
				return nodeMonth.days[i].day < nodeMonth.days[j].day
			})
			for _, nodeDay := range nodeMonth.days {
				sort.Slice(nodeDay.parties, func(i, j int) bool {
					return nodeDay.parties[i].party.PartyTime.Time().Before(nodeDay.parties[j].party.PartyTime.Time())
				})
			}
		}
	}

	return

}

func (m *TreeModel) RootCount() int {
	return len(m.years)
}

func (m *TreeModel) RootAt(index int) walk.TreeItem {
	return m.years[index]
}

var _ walk.TreeModel = new(TreeModel)
