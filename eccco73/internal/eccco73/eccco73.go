package eccco73

import (
	"database/sql"
		"time"

	"fmt"
	)

type PointsMethod int

const (
	PointsMethod2 = 2
	PointsMethod3 = 3
)

type Party struct {
	PartyID       PartyID        `db:"party_id"`
	CreatedAt     time.Time      `db:"created_at"`
	ProductTypeID ProductTypeID  `db:"product_type_id"`
	Note          sql.NullString `db:"note"`
	Gas1          float64        `db:"gas1"`
	Gas2          float64        `db:"gas2"`
	Gas3          float64        `db:"gas3"`
}

type Products []Product

type Product struct {
	ProductID     ProductID       `db:"product_id"`
	UpdatedAt     time.Time       `db:"updated_at"`
	PartyID       PartyID         `db:"party_id"`
	Order         int64           `db:"order_in_party"`
	ProductTypeID sql.NullInt64   `db:"product_type_id"`
	Serial        int64           `db:"serial_number"`
	Note          sql.NullString  `db:"note"`
	Fon20         sql.NullFloat64 `db:"fon20"`
	Sens20        sql.NullFloat64 `db:"sens20"`
	I13           sql.NullFloat64 `db:"i13"`
	I24           sql.NullFloat64 `db:"i24"`
	I35           sql.NullFloat64 `db:"i35"`
	I26           sql.NullFloat64 `db:"i26"`
	I17           sql.NullFloat64 `db:"i17"`
	In            sql.NullFloat64 `db:"not_measured"`
	FonMinus20    sql.NullFloat64 `db:"fon_minus20"`
	SensMinus20   sql.NullFloat64 `db:"sens_minus20"`
	Fon50         sql.NullFloat64 `db:"fon50"`
	Sens50        sql.NullFloat64 `db:"sens50"`
	FlashBytes    []byte          `db:"flash"`
	Production    bool            `db:"production"`
}



type ProductType struct {
	ProductTypeID               ProductTypeID   `db:"product_type_id"`
	Name                        string          `db:"product_type_name"`
	Gas                         string          `db:"gas"`
	Units                       string          `db:"units"`
	Scale                       float64         `db:"scale"`
	NobleMetalContent           float64         `db:"noble_metal_content"`
	LifetimeMonths              int             `db:"lifetime_months"`
	LC64                        bool            `db:"lc64"`
	PointsMethod                PointsMethod    `db:"points_method"`
	MaxFonCurrent               sql.NullFloat64 `db:"max_fon_curr"`
	MaxDeltaFonCurrent          sql.NullFloat64 `db:"max_delta_fon_curr"`
	MinCoefficientSensitivity   sql.NullFloat64 `db:"min_coefficient_sens"`
	MaxCoefficientSensitivity   sql.NullFloat64 `db:"max_coefficient_sens"`
	MinDeltaTemperature         sql.NullFloat64 `db:"min_delta_temperature"`
	MaxDeltaTemperature         sql.NullFloat64 `db:"max_delta_temperature"`
	MinCoefficientSensitivity50 sql.NullFloat64 `db:"min_coefficient_sens40"`
	MaxCoefficientSensitivity50 sql.NullFloat64 `db:"max_coefficient_sens40"`
	MaxDeltaNotMeasured         sql.NullFloat64 `db:"max_delta_not_measured"`
}

type PartyID int64
type ProductID int64
type ProductTypeID int64

type Party1 struct {
	PartyID         PartyID        `db:"party_id"`
	CreatedAt       time.Time      `db:"created_at"`
	Note            sql.NullString `db:"note"`
	ProductTypeName string         `db:"product_type_name"`
}
type NewParty struct {
	Serials          [8 * 12]int64
	Gas1, Gas2, Gas3 float64
	ProductType      ProductType
	Note             sql.NullString
}

type FoundProduct struct {
	ProductID              ProductID      `db:"product_id"`
	PartyID                PartyID        `db:"party_id"`
	CreatedAt              time.Time      `db:"created_at"`
	ProductProductTypeName sql.NullString `db:"product_type_name"`
	PartyProductTypeName   string         `db:"party_product_type_name"`
	Note                   sql.NullString `db:"note"`
	PartyNote              sql.NullString `db:"party_note"`
	Flash 					bool `db:"flash"`
}

func (x Party1) What() string {
	t := x.CreatedAt.Add(3 * time.Hour)
	return fmt.Sprintf("Партия ЭХЯ №%d  %s %s %s %s",
		x.PartyID,
		t.Format("02"),
		MonthNumberToName(t.Month()),
		t.Format("2006 15:04"),
		x.Note.String,
	)
}

func (x Party) What() string {
	return fmt.Sprintf("%d от %v", x.PartyID, x.CreatedAt.Add(time.Hour*3))
}

func (x Party) ProductType(types []ProductType) ProductType {
	for _, t := range types {
		if t.ProductTypeID == x.ProductTypeID {
			return t
		}
	}
	panic(fmt.Sprintf("invalid product type: %d", x.ProductTypeID))
}

func (x Products) ProductIndexByOrder(order int64) int {
	for i, p := range x {
		if p.Order == order {
			return i
		}
	}
	return -1
}

func (x Products) ProductByOrder(order int64) (Product, bool) {
	for _, p := range x {
		if p.Order == order {
			return p, true
		}
	}
	return Product{}, false
}

func FormatOrder8(n int64) string {
	return fmt.Sprintf("%d.%d", n/8+1, n%8+1)
}

func (x FoundProduct) ProductTypeName() string {
	if x.ProductProductTypeName.Valid {
		return x.ProductProductTypeName.String
	}
	return x.PartyProductTypeName

}
