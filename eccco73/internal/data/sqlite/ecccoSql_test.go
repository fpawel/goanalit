package sqlite

import (
	"fmt"
	"github.com/jmoiron/sqlx"

	"testing"
	"github.com/fpawel/eccco73/internal/data/old"
	"github.com/fpawel/goanalit/eccco73"
	"strings"
)

func TestRecords(t *testing.T) {
	db := MustConnect()
	types := db.GetProductTypes()
	party := db.GetLastParty()
	products := db.GetLastPartyProducts()
	fmt.Println(party)
	fmt.Println(products)
	fmt.Println(types)
	if err := db.conn.Close(); err != nil {
		panic(err)
	}
}

func TestImportProductTypes(t *testing.T) {
	fmt.Println(DBFilePath())
	conn := sqlx.MustConnect("sqlite3", DBFilePath())
	conn.MustExec(`
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';`)

	bs := old.ReadProductTypesFromFile()
	for _, x := range bs {
		pointsMethod := eccco73.PointsMethod3
		if x.PointsMethod == "Pt2" {
			pointsMethod = eccco73.PointsMethod2
		}
		conn.MustExec(`
INSERT INTO product_types
  ( product_type_name, gas, units, scale, noble_metal_content, lifetime_months, lc64, points_method, 
  	max_fon_curr, max_delta_fon_curr, min_coefficient_sens, max_coefficient_sens, min_delta_temperature, 
  	max_delta_temperature, min_coefficient_sens40, max_coefficient_sens40, max_delta_not_measured)
VALUES
      ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);
`, x.Name, x.Gas, x.Units, x.Scale, x.NobleMetalContent, x.LifetimeMonths, x.LC64, pointsMethod,
	x.MaxFonCurrent, x.MaxDeltaFonCurrent, x.MinCoefficientSensitivity, x.MaxCoefficientSensitivity, x.MinDeltaTemperature,
		x.MaxDeltaTemperature, x.MinCoefficientSensitivity40, x.MaxCoefficientSensitivity40, x.MaxDeltaNotMeasured)
	}
	conn.Close()
}

func TestPrintProductTypesSQL(t *testing.T) {

	for _, x := range old.ReadProductTypesFromFile() {
		pointsMethod := eccco73.PointsMethod3
		if x.PointsMethod == "Pt2" {
			pointsMethod = eccco73.PointsMethod2
		}
		lc64 := 0
		if x.LC64 {
			lc64 = 1
		}
		f := func(x *float64) string{
			if x == nil {
				return "NULL"
			}
			return fmt.Sprintf("%g", *x)
		}
		fmt.Printf("('%s', '%s', '%s', %g, %g, %d, '%d', %d, %s, %s, %s, %s, %s, %s, %s, %s, %s),\n",

			x.Name, x.Gas, x.Units, x.Scale, x.NobleMetalContent, x.LifetimeMonths, lc64, pointsMethod,
			f(x.MaxFonCurrent), f(x.MaxDeltaFonCurrent), f(x.MinCoefficientSensitivity), f(x.MaxCoefficientSensitivity),
			f(x.MinDeltaTemperature),
			f(x.MaxDeltaTemperature), f(x.MinCoefficientSensitivity40), f(x.MaxCoefficientSensitivity40), f(x.MaxDeltaNotMeasured),
		)
	}

}

func TestWorks(t *testing.T) {

	db := MustConnect()
	db.AddNewRootWork("Настройка ЭХЯ")
	db.AddNewWork(1, 0, "Продувка воздухом")
	db.AddNewWork(2, 1,  "Подать воздух")
	db.AddNewWork(3, 1,  "Задержка")
	db.AddNewWork(4, 1,  "Отключить пневмоблок")

	db.AddNewWork(5,0, "Перевод термокамеры")
	db.AddNewWork(6,5,"стоп")
	db.AddNewWork(7,5,  "уставка")
	db.AddNewWork(8,5,"старт")

	db.AddNewRootWork("Настройка ЭХЯ")
	db.AddNewWork(1, 0, "Продувка воздухом")
	db.AddNewWork(2, 1,  "Подать воздух")
	db.AddNewWork(3, 1,  "Задержка")
	db.AddNewWork(4, 1,  "Отключить пневмоблок")

	db.AddNewWork(5,0, "Перевод термокамеры")
	db.AddNewWork(6,5,"стоп")
	db.AddNewWork(7,5,  "уставка")
	db.AddNewWork(8,5,"старт")


	db.AddWorkMessage(2, 0, "пневмоблок: подать воздух")
	db.AddWorkMessage(3, 0, "пневмоблок: задержка")
	db.AddWorkMessage(4, 0, "пневмоблок: отключить")

	db.AddWorkMessage(6, 0, "термокамера: стоп")
	db.AddWorkMessage(7, 0, "термокамера: уставка")
	db.AddWorkMessage(8, 0, "термокамера: старт")
}

func Test1(t *testing.T) {

	db := MustConnect()

	fmt.Println(db.GetWorkIDByOrder(100))

	for i,x := range db.GetWorkMessages(1){
		fmt.Println(i, x.Text)
	}
	fmt.Println(strings.Repeat("-",20))
	for i,x := range db.GetWorkMessages(5){
		fmt.Println(i, x.Text)
	}
	fmt.Println(strings.Repeat("-",20))
	for i,x := range db.GetWorkMessages(0){
		fmt.Println(i, x.Text)
	}
}