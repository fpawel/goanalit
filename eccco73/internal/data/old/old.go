package old

import (
	"io/ioutil"
	"log"
	"github.com/fpawel/goanalit/eccco73"
	"fmt"
	"encoding/json"
	"path/filepath"
)

type ProductTypeOld struct {
	Name                        string  `json:"Name"`
	Gas                         string  `json:"Gas"`
	Units                       string  `json:"Units"`
	Scale                       float64 `json:"Scale"`
	NobleMetalContent           float64 `json:"NobleMetalContent"`
	LifetimeMonths              int     `json:"LifetimeWarrianty"`
	LC64                        bool    `json:"Is64"`
	PointsMethod				string  `json:"CalculateTermoMethod"`
	MaxFonCurrent               *float64 `json:"Ifon_max,omitempty"`
	MaxDeltaFonCurrent          *float64 `json:"DeltaIfon_max,omitempty"`
	MinCoefficientSensitivity   *float64 `json:"Ksns_min,omitempty"`
	MaxCoefficientSensitivity   *float64 `json:"Ksns_max,omitempty"`
	MinDeltaTemperature         *float64 `json:"Delta_t_min,omitempty"`
	MaxDeltaTemperature         *float64 `json:"Delta_t_max,omitempty"`
	MinCoefficientSensitivity40 *float64 `json:"Ks40_min,omitempty"`
	MaxCoefficientSensitivity40 *float64 `json:"Ks40_max,omitempty"`
	MaxDeltaNotMeasured         *float64 `json:"Delta_nei_max,omitempty"`
}

func ReadProductTypesFromFile() (bs [] ProductTypeOld){


	b, err := ioutil.ReadFile(filepath.Join(eccco73.ProductName.Path(), "productTypes.config.json"))
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(b, &bs) ; err != nil {
		log.Fatal(err)
	}
	for _, x := range bs {
		pointsMethod := eccco73.PointsMethod3
		if x.PointsMethod == "Pt2" {
			pointsMethod = eccco73.PointsMethod2
		}
		lc64 := "FALSE"
		if x.LC64 {
			lc64 = "TRUE"
		}
		f := func(x *float64) string{
			if x == nil {
				return "NULL"
			}
			return fmt.Sprintf("%g", *x)
		}
		fmt.Printf("('%s', '%s', '%s', %g, %g, %d, '%s', %d, %s, %s, %s, %s, %s, %s, %s, %s, %s),\n",

			x.Name, x.Gas, x.Units, x.Scale, x.NobleMetalContent, x.LifetimeMonths, lc64, pointsMethod,
			f(x.MaxFonCurrent), f(x.MaxDeltaFonCurrent), f(x.MinCoefficientSensitivity), f(x.MaxCoefficientSensitivity),
			f(x.MinDeltaTemperature),
			f(x.MaxDeltaTemperature), f(x.MinCoefficientSensitivity40), f(x.MaxCoefficientSensitivity40), f(x.MaxDeltaNotMeasured),
		)
	}
	return
}
