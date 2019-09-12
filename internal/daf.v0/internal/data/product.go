package data

import (
	"github.com/fpawel/comm/modbus"
	"time"
)

//go:generate reform

// Product represents a row in product table.
//reform:product
type Product struct {
	ProductID int64       `reform:"product_id,pk"`
	PartyID   int64       `reform:"party_id"`
	CreatedAt time.Time   `reform:"created_at"`
	Serial    int64       `reform:"serial"`
	Addr      modbus.Addr `reform:"addr"`
}
