package dbt

import (
	"github.com/shopspring/decimal"
	"time"
)

type Ad struct {
	ID          int64
	Location    string
	Description string
}

type AdService interface {
	// Simple check availablility
	Ping()
	ErrorWithoutArgs() error
	ResultWithoutArgs() (int64, error)

	ArgsWithoutResult(x, y, z int64)
	ArgsWithError(x, y, z int64, ad Ad, stamp time.Time, duration time.Duration, value decimal.Decimal, data []byte) error
	ArgsWithResult(x, y, z int64) (int64, error)
}
