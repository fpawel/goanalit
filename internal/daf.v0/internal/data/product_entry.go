package data

import (
	"time"
)

//go:generate reform

// ProductEntry represents a row in product_entry table.
//reform:product_entry
type ProductEntry struct {
	ProductEntryID int64     `reform:"product_entry_id,pk"`
	ProductID      int64     `reform:"product_id"`
	CreatedAt      time.Time `reform:"created_at"`
	WorkName       string    `reform:"work_name"`
	Ok             bool      `reform:"ok"`
	Message        string    `reform:"message"`
}
