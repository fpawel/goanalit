package sqlite

import (
	"github.com/fpawel/eccco73/internal/eccco73"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	)



type DB struct {
	conn *sqlx.DB
}

func (x DB) Close() error {
	return x.conn.Close()
}

func MustConnect(filename string) (x DB) {
	x.conn = sqlx.MustConnect("sqlite3", filename)
	x.conn.MustExec(`
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';`)
	return
}

func (x DB) GetProductTypes() (r []eccco73.ProductType) {
	rows, err := x.conn.Queryx(`SELECT * FROM product_types;`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var p eccco73.ProductType
		if err = rows.StructScan(&p); err != nil {
			log.Fatal(err)
		}
		r = append(r, p)
	}
	return
}

func (x DB) GetLastParty() (r eccco73.Party) {
	err := x.conn.Get(&r, `SELECT * FROM parties ORDER BY created_at DESC LIMIT 1;`)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) GetLastPartyID() (r eccco73.PartyID) {
	err := x.conn.Get(&r, `SELECT party_id FROM parties ORDER BY created_at DESC LIMIT 1;`)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) GetLastPartyProducts() (xs eccco73.Products) {
	err := x.conn.Select(&xs, `
SELECT * FROM products
WHERE party_id IN (
  SELECT party_id FROM parties ORDER BY created_at DESC LIMIT 1
);`)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) UpdateProductProduction(productID eccco73.ProductID, production bool) {
	_, err := x.conn.Exec(`UPDATE products SET production = $1, updated_at = current_timestamp WHERE product_id = $2`, production, productID)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) UpdateProduct(p eccco73.Product) {
	_, err := x.conn.Exec(`
UPDATE products SET product_type_id = $1, serial_number = $2, note = $3, updated_at = current_timestamp 
WHERE product_id = $4`, p.ProductTypeID, p.Serial, p.Note, p.ProductID)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) UpdateParty(p eccco73.Party) {
	_, err := x.conn.Exec(`
UPDATE parties 
SET product_type_id = $1, note = $2, gas1 = $3, gas2 = $4, gas3 = $5     
WHERE party_id = $6`, p.ProductTypeID, p.Note, p.Gas1, p.Gas2, p.Gas3, p.PartyID)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) GetYears() (xs []int64) {
	err := x.conn.Select(&xs, `
SELECT cast(strftime('%Y', created_at) AS INT) AS year FROM parties GROUP BY year;`)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) GetMonthsOfYear(year int64) (xs []int64) {
	err := x.conn.Select(&xs, `
SELECT cast( strftime('%m', created_at) AS INT) AS month FROM parties
WHERE cast(strftime('%Y', created_at) AS INT) = $1
GROUP BY month;
`, year)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) GetDaysOfYearMonth(year, month int64) (xs []int64) {
	err := x.conn.Select(&xs, `
SELECT cast( strftime('%d', created_at) AS INT) AS day FROM parties
WHERE  cast(strftime('%Y', created_at) AS INT) = $1 AND cast(strftime('%m', created_at) AS INT) = $2
GROUP BY day;
`, year, month)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) GetPartiesOfYearMonthDay(year, month, day int64) (xs []eccco73.Party1) {
	err := x.conn.Select(&xs, `
SELECT party_id, created_at, note, product_type_name FROM parties
INNER JOIN product_types pt on parties.product_type_id = pt.product_type_id
WHERE
  cast(strftime('%Y', created_at) AS INT) = $1 AND
  cast(strftime('%m', created_at) AS INT) = $2 AND
  cast(strftime('%d', created_at) AS INT) = $3
ORDER BY created_at;
`, year, month, day)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (x DB) GetPartyProductByPartyID(partyID eccco73.PartyID) (party eccco73.Party, products eccco73.Products) {
	err := x.conn.Get(&party, `SELECT * FROM parties WHERE party_id = $1;`, partyID)
	if err != nil {
		log.Panicf("GetPartyProductByPartyID: %d: %v", partyID, err)
	}
	err = x.conn.Select(&products, `SELECT * FROM products WHERE party_id = $1 ORDER BY order_in_party ASC;`, partyID)
	if err != nil {
		log.Panic(err)
	}
	return
}
func (x DB) CreateNewParty(p eccco73.NewParty) (newPartyID eccco73.PartyID, newProductsID [96]eccco73.ProductID) {
	r, err := x.conn.Exec(`
BEGIN TRANSACTION;
INSERT INTO parties (product_type_id, note, gas1, gas2, gas3)  VALUES($1, $2, $3, $4, $5);
SELECT last_insert_rowid();`, p.ProductType.ProductTypeID, p.Note, p.Gas1, p.Gas2, p.Gas3)
	if err != nil {
		log.Panic(err)
	}
	{
		v, err := r.LastInsertId()
		if err != nil {
			log.Panic(err)
		}
		newPartyID = eccco73.PartyID(v)
	}

	for i, serial := range p.Serials {
		if serial < 1 {
			continue
		}
		r, err := x.conn.Exec(`
INSERT INTO products (party_id, order_in_party, serial_number, updated_at, production) VALUES ($1,$2,$3, current_timestamp, 0 );
SELECT last_insert_rowid();`, newPartyID, i, serial)

		if err != nil {
			log.Panic(err)
		}

		v, err := r.LastInsertId()
		if err != nil {
			log.Panic(err)
		}
		newProductsID[i] = eccco73.ProductID(v)
	}
	x.conn.MustExec("COMMIT;")

	return
}

func (x DB) DeletePartyByID(partyID eccco73.PartyID) {
	x.conn.MustExec(`
DELETE FROM import_parties WHERE party_id=$1;
DELETE FROM parties WHERE party_id=$2;`, partyID, partyID)
}

func (x DB) FindProductsBySerial(serial int64) (xs []eccco73.FoundProduct) {
	const query = `
SELECT pr.created_at, pt.product_type_name party_product_type_name,
  ptt.product_type_name,
  p.product_id, pr.party_id, p.note, pr.note as party_note, 
  	CASE WHEN length(flash) > 0 THEN 
  		1 
  	ELSE 
  		0 
  	END AS flash
FROM products p
  INNER JOIN parties pr ON p.party_id = pr.party_id
  INNER JOIN product_types pt on pr.product_type_id = pt.product_type_id
  LEFT JOIN product_types ptt on p.product_type_id = ptt.product_type_id
WHERE p.serial_number = $1;`
	err := x.conn.Select(&xs, query, serial)
	if err != nil {
		log.Panic(err)
	}
	return
}


