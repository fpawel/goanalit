package eccco73

import (
	"math"
	"database/sql"
	"github.com/fpawel/eccco73/internal/flash"
)

type Product2 struct {
	Product
	Party Party
	Types []ProductType
}

func (x Product) Flash() flash.Bytes {
	return flash.Bytes(x.FlashBytes)
}

func (x Product) CurrentToCoefficientSensitivityPercent() DeltaCurrentToCoefficientSensitivityPercentFunc {
	if x.Fon20.Valid && x.Sens20.Valid && x.Fon20.Float64 != x.Sens20.Float64 {
		d := x.Sens20.Float64 - x.Fon20.Float64
		if d != 0 {
			return func (curFon,curSens float64) float64{
				return 100 * math.Abs ( (curSens-curFon)/d )
			}
		}
	}
	return nil
}

func (x Product) OriginalTemperatureToCoefficientSensitivityPercent() (r X2Y) {
	f := x.CurrentToCoefficientSensitivityPercent()
	if f == nil {
		return
	}
	r = make(map[float64]float64)

	ts := []float64{-20,20,50}
	cs := []sql.NullFloat64{ x.SensMinus20, x.Sens20, x.Sens50 }
	cf := []sql.NullFloat64{ x.FonMinus20, x.Fon20, x.Fon50 }
	for i:=0; i<2; i++{
		if cs[i].Valid && cf[i].Valid {
			r[ts[i]] = f(cf[i].Float64, cs[i].Float64)
		}
	}
	return
}

func (x Product) CoefficientSensitivity50() sql.NullFloat64 {

	if x.Fon50.Valid && x.Sens50.Valid {
		if f := x.CurrentToCoefficientSensitivityPercent(); f != nil {
			return sql.NullFloat64{
				Valid:true,
				Float64:f(x.Fon50.Float64, x.Sens50.Float64),
			}
		}
	}
	return sql.NullFloat64{}
}

func (x Product) CoefficientSensitivity20(party Party) sql.NullFloat64 {
	d := party.Gas3 - party.Gas1
	if d > 0 && x.Fon20.Valid && x.Sens20.Valid && x.Fon20.Float64 != x.Sens20.Float64 {
		return sql.NullFloat64{
			Float64: (x.Sens20.Float64 - x.Fon20.Float64) / d,
			Valid:   true,
		}
	}
	return sql.NullFloat64{}
}
