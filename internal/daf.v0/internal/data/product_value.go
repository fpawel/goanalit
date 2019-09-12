package data

import (
	"time"
)

//go:generate reform

// ProductValue represents a row in product_value table.
//reform:product_value
type ProductValue struct {
	ProductValueID int64     `reform:"product_value_id,pk"`
	ProductID      int64     `reform:"product_id"`
	CreatedAt      time.Time `reform:"created_at"`
	WorkIndex      int       `reform:"work_index"`
	Gas            Gas       `reform:"gas"`
	Concentration  float64   `reform:"concentration"`
	Current        float64   `reform:"current"`
	Threshold1     bool      `reform:"threshold1"`
	Threshold2     bool      `reform:"threshold2"`
	Mode           uint16    `reform:"mode"`
	FailureCode    float64   `reform:"failure_code"`
}
