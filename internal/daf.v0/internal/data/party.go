package data

import (
	"time"
)

//go:generate reform

// Party represents a row in party table.
//reform:party
type Party struct {
	PartyID              int64     `reform:"party_id,pk"`
	CreatedAt            time.Time `reform:"created_at"`
	Type                 int       `reform:"type"`
	Component            string    `reform:"component"`
	Scale                float64   `reform:"scale"`
	AbsoluteErrorRange   float64   `reform:"absolute_error_range"`
	AbsoluteErrorLimit   float64   `reform:"absolute_error_limit"`
	RelativeErrorLimit   float64   `reform:"relative_error_limit"`
	Threshold1Production float64   `reform:"threshold1_production"`
	Threshold2Production float64   `reform:"threshold2_production"`
	Threshold1Test       float64   `reform:"threshold1_test"`
	Threshold2Test       float64   `reform:"threshold2_test"`
	Pgs1                 float64   `reform:"pgs1"`
	Pgs2                 float64   `reform:"pgs2"`
	Pgs3                 float64   `reform:"pgs3"`
	Pgs4                 float64   `reform:"pgs4"`
}
