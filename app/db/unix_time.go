package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// UnixTime is a wrapper around time.Time to work with sqlite's
// timestamps (using int64).
type UnixTime struct {
	Time time.Time
}

func (ut *UnixTime) Scan(src any) error {
	if src == nil {
		ut.Time = time.Time{}
		return nil
	}

	switch v := src.(type) {
	case int64:
		ut.Time = time.Unix(v, 0)
	default:
		return fmt.Errorf("unsupported type for UnixTime: %T, expected int64", src)
	}
	return nil
}

func (ut UnixTime) Value() (driver.Value, error) {
	return ut.Time.Unix(), nil
}

func Now() UnixTime {
	return UnixTime{Time: time.Now()}
}
