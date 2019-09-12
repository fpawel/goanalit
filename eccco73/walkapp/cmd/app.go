package main

import (
	"github.com/fpawel/eccco73/internal/data/sqlite"
	"github.com/lxn/walk"

	"database/sql"
	"fmt"
	"github.com/fpawel/eccco73/internal/eccco73"
	"github.com/fpawel/eccco73/internal/walkui/dialogs/dialogEditParty"
	"github.com/fpawel/eccco73/internal/walkui/dialogs/dialogEditProduct"
	"github.com/fpawel/eccco73/internal/walkui/dialogs/dialogFindProduct"
	"github.com/fpawel/eccco73/internal/walkui/dialogs/dialogNewParty"
	"github.com/lxn/win"
	"log"
	"time"
)

type App struct {
	*walk.Application
	config                   AppConfig
	mw                       *AppMainWindow
	cancellationDelay        int32
	db                       sqlite.DB
	productsTableModel       *ProductsTableModel
	productProductTypesModel []dialogEditProduct.ProductTypeModel
	productTypes             []eccco73.ProductType
	products                 eccco73.Products
	party                    eccco73.Party
	partiesTreeViewModel     *PartiesTreeViewModel
}

func NewApp() *App {
	db := sqlite.MustConnect(appFolderFileName("products.db"))
	x := &App{
		Application:  walk.App(),
		config:       NewAppConfig(),
		db:           db,
		productTypes: db.GetProductTypes(),
		party:        db.GetLastParty(),
		products:     db.GetLastPartyProducts(),
	}
	x.productsTableModel = &ProductsTableModel{
		app: x,
	}
	x.UpdateProductProductTypesModel()
	x.partiesTreeViewModel = &PartiesTreeViewModel{
		partiesReader: db,
		party:         &x.party,
	}
	x.mw = NewAppMainWindow(x)
	setupWalkApplication(x.Application)
	x.config = NewAppConfig()

	err := x.mw.Markup().Create()
	if err != nil {
		log.Panic(err)
	}
	x.mw.Initialize()

	return x
}

func (x *App) IsLastParty() bool {
	return x.db.GetLastPartyID() == x.party.PartyID
}

func (x *App) OpenPartyByPartyID(partyID eccco73.PartyID) {
	m := x.partiesTreeViewModel
	x.party, x.products = x.db.GetPartyProductByPartyID(partyID)
	m.PublishChanged()
	x.productsTableModel.PublishRowsReset()
}

func (x *App) UpdateProductProductTypesModel() {
	type M = dialogEditProduct.ProductTypeModel
	x.productProductTypesModel = []M{{}}
	for _, t := range x.db.GetProductTypes() {
		k := int64(t.ProductTypeID)
		x.productProductTypesModel = append(x.productProductTypesModel,
			M{
				Name:          t.Name,
				ProductTypeID: sql.NullInt64{Valid: true, Int64: k},
			})
	}
}

func (x *App) Close() {

	err := x.Settings().(*walk.IniFileSettings).Save()
	if err != nil {
		log.Panic(err)
	}
	x.config.Save()
	for err := x.db.Close(); err != nil; err = x.db.Close() {
		time.Sleep(time.Millisecond * 100)
	}
}

func (x *App) ExecuteDeletePartyDialog() {

	var party eccco73.Party1
	switch n := x.mw.tvParties.CurrentItem().(type) {
	case *PartiesPartyNode:
		party = n.party
	default:
		return
	}

	msg := fmt.Sprintf("Подтвердите необходимость удаления партии:\n\n%s.", party.What())
	if walk.MsgBox(x.mw, "Подтверждение", msg, walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) != win.IDOK {
		return
	}
	x.db.DeletePartyByID(party.PartyID)
	for _, y := range x.partiesTreeViewModel.years {
		if y.year == int64(party.CreatedAt.Year()) {
			x.partiesTreeViewModel.PublishItemsReset(y)
		}
	}
	if party.PartyID == x.party.PartyID {
		x.SetLastParty()
	} else {
		x.ResetPartiesTreeViewModel()
	}
}

func (x *App) SetLastParty() {
	x.party = x.db.GetLastParty()
	x.products = x.db.GetLastPartyProducts()
	x.productsTableModel.PublishRowsReset()
	x.mw.SetPartyTitle()
	x.ResetPartiesTreeViewModel()
}

func (x *App) ResetPartiesTreeViewModel() {
	tm := &PartiesTreeViewModel{
		party:         &x.party,
		partiesReader: x.db,
	}
	x.partiesTreeViewModel = tm
	if err := x.mw.tvParties.SetModel(tm); err != nil {
		log.Panic(err)
	}
}
func (x *App) ExecuteNewPartyDialog() {
	if newParty, ok := dialogNewParty.Execute(x.mw, x.productTypes); ok {
		x.db.CreateNewParty(newParty)
		x.SetLastParty()
	}
}

func (x *App) ExecuteEditPartyDialog() {
	if !x.IsLastParty() {
		x.SetLastParty()
	}

	if party, ok := dialogEditParty.Execute(x.mw, x.party, x.productTypes); ok {
		x.db.UpdateParty(party)
		x.party = party
		x.productsTableModel.PublishRowsReset()
		x.ResetPartiesTreeViewModel()
		x.mw.SetPartyTitle()
	}
}

func (x *App) ExecuteEditProductDialog() {
	if p, ok := dialogEditProduct.Execute(x.mw, x.party, x.mw.clickedProduct, x.productProductTypesModel); ok {
		x.db.UpdateProduct(p)

		for i := range x.products {
			if x.products[i].ProductID == p.ProductID {
				x.products[i] = p
				break
			}
		}
		x.productsTableModel.PublishRowsReset()
	}
}
func (x *App) ExecuteFindProductBySerialDialog() {
	if p, ok := dialogFindProduct.Execute(x.mw, x.db, x.config.FindProductSerial); ok {
		x.OpenPartyByPartyID(p.PartyID)
		x.mw.SetPartyTitle()
		for i, xp := range x.products {
			if xp.ProductID == p.ProductID {
				x.config.FindProductSerial = xp.Serial
				if err := x.mw.tblProducts.SetCurrentIndex(i); err != nil {
					log.Panic(err)
				}
				x.mw.TableProductsItemActivated()
			}
		}

	}
}
