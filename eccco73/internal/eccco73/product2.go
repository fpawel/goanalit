package eccco73

import (
	"math"
	"database/sql"
)

func (x Product2) ProductType() ProductType {
	if x.ProductTypeID.Valid {
		for _, t := range x.Types {
			if t.ProductTypeID == ProductTypeID(x.ProductTypeID.Int64) {
				return t
			}
		}
	}
	return x.Party.ProductType(x.Types)
}

func (x Product2) Ok() bool {
	return x.OkFon20() && x.OkDeltaFon20() && x.OkDeltaFonTemperature() && x.OkDeltaFonN() &&
		x.OkCoefficientSensitivity20() && x.OkCoefficientSensitivity50()

}

func (x Product2) OkFon20() bool {
	t := x.ProductType()
	return !( t.MaxFonCurrent.Valid && x.Fon20.Valid &&
		math.Abs(x.Fon20.Float64) >= math.Abs(t.MaxFonCurrent.Float64) )
}

func (x Product2) OkDeltaFon20() bool {
	t := x.ProductType()
	d := x.DeltaFon20()
	return !( t.MaxDeltaFonCurrent.Valid && d.Valid &&
		math.Abs(d.Float64) >= math.Abs(t.MaxDeltaFonCurrent.Float64) )
}

func (x Product2) OkDeltaFonTemperature() bool {
	t := x.ProductType()
	d := x.DeltaFonTemperature()
	if t.MinDeltaTemperature.Valid && t.MaxDeltaTemperature.Valid && d.Valid {
		da := math.Abs(d.Float64)
		return da < t.MaxDeltaTemperature.Float64 && da > t.MinDeltaTemperature.Float64
	}
	return true
}

func (x Product2) OkDeltaFonN() bool {
	t := x.ProductType()
	d := x.DeltaFonN()
	return !( t.MaxDeltaNotMeasured.Valid && d.Valid &&
		math.Abs(d.Float64) >= t.MaxDeltaNotMeasured.Float64 )
}

func (x Product2) OkCoefficientSensitivity20() bool {
	t := x.ProductType()
	kv := x.CoefficientSensitivity20()
	if t.MaxCoefficientSensitivity.Valid && t.MinCoefficientSensitivity.Valid && kv.Valid {
		k := math.Abs(kv.Float64)
		return k < t.MaxCoefficientSensitivity.Float64 && k > t.MinCoefficientSensitivity.Float64
	}
	return true
}

func (x Product2) OkCoefficientSensitivity50() bool {
	t := x.ProductType()
	kv := x.CoefficientSensitivity50()
	if t.MaxCoefficientSensitivity50.Valid && t.MinCoefficientSensitivity50.Valid && kv.Valid {
		k := math.Abs(kv.Float64)
		return k < t.MaxCoefficientSensitivity50.Float64 && k > t.MinCoefficientSensitivity50.Float64
	}
	return true
}

func (x Product2) CoefficientSensitivity20() sql.NullFloat64 {
	return x.Product.CoefficientSensitivity20(x.Party)
}


func (x Product2) DeltaFon20() (r sql.NullFloat64) {
	r.Valid = x.Fon20.Valid && x.I13.Valid
	if !r.Valid {
		return
	}
	r.Float64 = x.I13.Float64 - x.Fon20.Float64
	return
}

func (x Product2) DeltaFonTemperature() (r sql.NullFloat64) {
	r.Valid = x.Fon20.Valid && x.Fon50.Valid
	if !r.Valid {
		return
	}
	r.Float64 = x.Fon50.Float64 - x.Fon20.Float64
	return
}

func (x Product2) DeltaFonN() (r sql.NullFloat64) {
	k := x.CoefficientSensitivity20()
	r.Valid = x.Fon20.Valid && x.In.Valid && k.Valid && k.Float64 != 0
	if !r.Valid {
		return
	}
	r.Float64 = (x.In.Float64 - x.Fon20.Float64) / k.Float64
	return
}
