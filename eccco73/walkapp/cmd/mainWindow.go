package main

import (
	"github.com/lxn/walk"

	"fmt"
	"github.com/fpawel/eccco73/internal/eccco73"
	"log"
)

type AppMainWindow struct {
	*walk.MainWindow
	app                    *App
	tblProducts            *walk.TableView
	tblWorks               *walk.TableView
	btnCancel              *walk.PushButton
	btnCancelDelay         *walk.PushButton
	btnCloseClickedProduct *walk.ToolButton
	actionDeleteParty      *walk.Action
	svWork                 *walk.ScrollView
	lblWork                *walk.Label
	lblWorkMessage         *walk.TextEdit
	progressBar            *walk.ProgressBar
	lblParty               *walk.Label
	tvParties              *walk.TreeView
	ivChart                *walk.ImageView
	clickedProduct         eccco73.Product
	initialized            bool
}

func NewAppMainWindow(app *App) *AppMainWindow {
	return &AppMainWindow{
		app: app,
	}
}

func (x *AppMainWindow) Initialize() {
	x.progressBar.SetVisible(false)
	x.btnCancel.SetVisible(false)
	x.btnCancelDelay.SetVisible(false)
	x.SetPartyTitle()
	x.svWork.SetVisible(false)
	x.tvParties.SetVisible(false)
	x.ivChart.SetVisible(false)
	x.btnCloseClickedProduct.SetVisible(false)
	x.initialized = true
}

func (x *AppMainWindow) TableProductsItemActivated() {
	i := x.tblProducts.CurrentIndex()
	p := x.app.products[i]
	if x.clickedProduct.ProductID == p.ProductID && x.ivChart.Visible() && x.app.IsLastParty() {
		x.app.ExecuteEditProductDialog()
	} else {

		x.clickedProduct = p
		if len(p.FlashBytes) > 0 {
			x.ShowClickedProduct()
		} else {
			x.HideClickedProduct()
		}
		x.app.productsTableModel.PublishRowsReset()
		if err := x.tblProducts.SetCurrentIndex(i); err != nil {
			log.Panic(err)
		}
	}
}

func (x *AppMainWindow) TreeViewPartiesItemActivated() {
	switch n := x.tvParties.CurrentItem().(type) {
	case *PartiesPartyNode:
		x.app.OpenPartyByPartyID(n.party.PartyID)
		x.SetPartyTitle()
	}
}

func (x *AppMainWindow) TreeViewPartiesItemChanged() {
	switch n := x.tvParties.CurrentItem().(type) {
	case *PartiesPartyNode:
		if err := x.actionDeleteParty.SetText(fmt.Sprintf("Удалить партию: %q", n.party.What())); err != nil {
			log.Panic(err)
		}
		x.actionDeleteParty.SetVisible(true)
	default:
		x.actionDeleteParty.SetVisible(false)
	}
}

func (x *AppMainWindow) SetPartyTitle() {
	p := x.app.party
	t := p.CreatedAt
	err := x.lblParty.SetText(fmt.Sprintf("%s №%d  %s %s %s",
		p.ProductType(x.app.productTypes).Name,
		p.PartyID,
		t.Format("02"),
		monthNumberToName(t.Month()),
		t.Format("2006 15:04"),
	))
	if err != nil {
		log.Panic(err)
	}
}
