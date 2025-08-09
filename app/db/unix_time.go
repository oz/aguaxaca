// This file is part of Aguaxaca.
// Copyright (C) 2025 Arnaud Berthomier.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
		ut.Time = time.Unix(v, 0).UTC()
	default:
		return fmt.Errorf("unsupported type for UnixTime: %T, expected int64", src)
	}
	return nil
}

func (ut UnixTime) Value() (driver.Value, error) {
	return ut.Time.Unix(), nil
}

func Now() UnixTime {
	return UnixTime{Time: time.Now().UTC()}
}
