// Copyright 2025 Clivern. All rights reserved.
// Use of this code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

// ID is a primary or foreign key that can be either an integer (SQLite) or UUID (PostgreSQL).
// It implements sql.Scanner and driver.Valuer so the same structs work with both drivers.
type ID string

// Scan implements sql.Scanner. Accepts int64 (SQLite), []byte, or string (PostgreSQL UUID).
func (id *ID) Scan(value interface{}) error {
	if value == nil {
		*id = ""
		return nil
	}
	switch v := value.(type) {
	case int64:
		*id = ID(strconv.FormatInt(v, 10))
		return nil
	case []byte:
		*id = ID(string(v))
		return nil
	case string:
		*id = ID(v)
		return nil
	default:
		return fmt.Errorf("db: cannot scan type %T into ID", value)
	}
}

// Value implements driver.Valuer for INSERT/UPDATE.
func (id ID) Value() (driver.Value, error) {
	if id == "" {
		return nil, nil
	}
	return string(id), nil
}

// String returns the ID as string.
func (id ID) String() string {
	return string(id)
}

// ValueOrNil returns the ID as interface{} for use in queries; returns nil if id is empty (for nullable columns).
func (id ID) ValueOrNil() interface{} {
	if id == "" {
		return nil
	}
	return string(id)
}
