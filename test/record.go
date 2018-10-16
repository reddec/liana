package dbt

import (
	"database/sql"
	"github.com/shopspring/decimal"
	"time"
)

// Advertisement item
type Ad struct {
	ID          int64  // Unique ID of ad
	Location    string // Post address of AD
	Description string // Custom description
}

// Advertisement manager
type AdService interface {
	// Simple check availablility
	Ping()
	ErrorWithoutArgs() error
	ResultWithoutArgs() (int64, error)

	ArgsWithoutResult(x, y, z int64)
	ArgsWithError(x, y, z int64, ad Ad, stamp time.Time, duration time.Duration, value decimal.Decimal, data []byte) error
	ArgsWithResult(x, y, z int64, val sql.NullInt64) (int64, error)
}
