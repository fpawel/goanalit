package main

import (
	"github.com/fpawel/gohelp"
	"github.com/powerman/structlog"
	"time"
)

const (
	KeyAddr   = "_addr"
	KeyPlace  = "_place"
	KeyEN6408 = "_en6408"
)

func now() string {
	return time.Now().Format("15:04:05")
}

func withProductAtPlace(log *structlog.Logger, place int) *structlog.Logger {
	product := prodsMdl.ProductAt(place)
	return gohelp.LogPrependSuffixKeys(log, "место", place+1, "заводской_номер", product.ProductID,
		"product_id", product.ProductID)
}
