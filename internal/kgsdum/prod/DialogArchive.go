package main

import (
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	productsView2 "github.com/fpawel/goanalit/internal/kgsdum/productsView"
	"github.com/fpawel/gutils/walkUtils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"log"
)

type DialogArchive struct {
	*walk.MainWindow
	tblLogs, tblProduct *walk.TableView
	treeItem            walk.TreeItem
	treeView            *walk.TreeView
	btnDelete           *walk.PushButton
	app                 *App
}

func (x *App) ExecuteDialogArchive() {
	dlg := &DialogArchive{app: x}
	x.mw.Hide()
	check(dlg.Markup().Create())
	dlg.Invalidate()
	win.ShowWindow(dlg.Handle(), win.SW_MAXIMIZE)
	if windowResult := dlg.Run(); windowResult != 0 {
		log.Println("PartiesDialog  windowResult:", windowResult)
	}
	x.mw.Show()
	x.tableProductsModel.PublishRowsReset()
	x.tableLogsModel.PublishRowsReset()
}

func ProductsInfo() productsView2.DeviceInfoProvider {
	return productsView2.DeviceInfoProvider{
		GoodProduct: func(product products2.Product) bool {
			return false
		},
		BadProduct: func(product products2.Product) bool {
			return false
		},
		FormatProductType: func(i int) string {
			return ""
		},
	}
}

func (x *DialogArchive) Invalidate() {
	m := productsView2.NewPartiesTreeViewModel(x.app.db, ProductsInfo())
	check(x.treeView.SetModel(m))

	var currentPartyTime products2.PartyTime
	x.app.db.View(func(tx products2.Tx) {
		currentPartyTime = tx.Party().PartyTime
	})
	itemContainsCurrentParty := func(it walk.TreeItem) (result bool) {
		switch it := it.(type) {
		case productsView2.Node:
			result = it.ContainsParty(currentPartyTime)
		}
		return
	}
	walkUtils.SetExpandedTreeviewItems(x.treeView, true, itemContainsCurrentParty)
}

func (x *DialogArchive) deleteSelectedNode() {
	if x.treeItem == nil {
		return
	}
	x.app.db.Update(func(tx products2.Tx) {
		node := x.treeItem.(productsView2.Node)
		productsView2.DeleteNode(node, tx)

	})
	x.Invalidate()
}

func (x *DialogArchive) Markup() MainWindow {

	var partyTime products2.PartyTime
	x.app.db.View(func(tx products2.Tx) {
		partyTime = tx.Party().PartyTime
	})

	const fontPointSize = 10

	return MainWindow{
		Icon:     NewIconFromResourceId(IconDBID),
		AssignTo: &x.MainWindow,
		Title:    "Обзор партий",
		Layout:   HBox{},

		Children: []Widget{
			TreeView{
				Model:    productsView2.NewPartiesTreeViewModel(x.app.db, ProductsInfo()),
				AssignTo: &x.treeView,
				MaxSize:  Size{200, 0},

				Font: Font{
					Family:    "Arial",
					PointSize: 12,
				},
				OnCurrentItemChanged: x.treeItemChanged,
			},

			TableView{
				AssignTo:              &x.tblLogs,
				AlternatingRowBGColor: walk.RGB(239, 239, 239),
				CheckBoxes:            false,
				ColumnsOrderable:      false,
				MultiSelection:        true,
				Font:                  Font{PointSize: fontPointSize},
				Model:                 productsView2.NewTableLogsViewModel(x.app.db, partyTime),
				Columns:               productsView2.TableLogsColumns(),
			},

			TableView{
				AssignTo:              &x.tblProduct,
				AlternatingRowBGColor: walk.RGB(239, 239, 239),
				CheckBoxes:            false,
				ColumnsOrderable:      false,
				MultiSelection:        true,
				Font:                  Font{PointSize: fontPointSize},
			},

			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{
					PushButton{
						AssignTo:  &x.btnDelete,
						Text:      "Удалить",
						OnClicked: x.deleteSelectedNode,
					},
				},
			},
		},
	}
}

func (x *DialogArchive) treeItemChanged() {
	x.treeItem = x.treeView.CurrentItem()

	switch node := x.treeItem.(type) {
	case *productsView2.NodeParty:
		x.tblLogs.SetModel(productsView2.NewTableLogsViewModel(x.app.db, node.Party().PartyTime))
		x.tblLogs.SetVisible(true)
		x.tblProduct.SetVisible(false)
		x.btnDelete.SetVisible(true)
	case *productsView2.NodeProduct:
		SetTableProductColumns(x.tblProduct.Columns())

		//m := &TableProductModel{product: node.Product(), db: x.app.db}
		m := &TableProductModel{product: node.Product(), db: x.app.db}
		check(x.tblProduct.SetModel(m))
		x.tblProduct.SetCellStyler(m)
		x.tblLogs.SetVisible(false)
		x.tblProduct.SetVisible(true)
		x.btnDelete.SetVisible(false)
	default:
		x.tblLogs.SetVisible(false)
		x.btnDelete.SetVisible(true)
		check(x.tblProduct.Columns().Clear())
		x.tblProduct.SetVisible(true)
	}
}
