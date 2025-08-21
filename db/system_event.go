// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// SystemEvent is a single row in the system_events table (runner_added, user_invited, etc.)
type SystemEvent struct {
	ID        ID
	Type      string
	Data      json.RawMessage
	CreatedAt time.Time
}

// SystemEventPlatformInstalled is the event data for "platform_installed".
type SystemEventPlatformInstalled struct {
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email"`
}

// SystemEventRepository is the interface for system event persistence.
type SystemEventRepository interface {
	Create(eventType string, data interface{}) error
	GetByID(id ID) (*SystemEvent, error)
	List(limit, offset int) ([]*SystemEvent, error)
	ListByType(eventType string, limit, offset int) ([]*SystemEvent, error)
	Count() (int64, error)
	CountByType(eventType string) (int64, error)
}

// NewSystemEventRepository returns the repository for the current driver
func NewSystemEventRepository(db *sql.DB) SystemEventRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &systemEventRepoPostgres{db: db}
	}
	return &systemEventRepoSQLite{db: db}
}

func scanSystemEventRows(rows *sql.Rows) ([]*SystemEvent, error) {
	var list []*SystemEvent
	for rows.Next() {
		e := &SystemEvent{}
		var data []byte
		if err := rows.Scan(&e.ID, &e.Type, &data, &e.CreatedAt); err != nil {
			return nil, err
		}
		if len(data) > 0 {
			e.Data = data
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

// --- SQLite ---

type systemEventRepoSQLite struct {
	db *sql.DB
}

func (r *systemEventRepoSQLite) Create(eventType string, data interface{}) error {
	var dataJSON []byte
	if data != nil {
		var err error
		dataJSON, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	_, err := r.db.Exec(
		`INSERT INTO system_events (type, data) VALUES (?, ?)`,
		eventType, dataJSON,
	)
	return err
}

func (r *systemEventRepoSQLite) GetByID(id ID) (*SystemEvent, error) {
	e := &SystemEvent{}
	var data []byte
	err := r.db.QueryRow(
		`SELECT id, type, data, created_at FROM system_events WHERE id = ?`,
		id.String(),
	).Scan(&e.ID, &e.Type, &data, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		e.Data = data
	}
	return e, nil
}

func (r *systemEventRepoSQLite) List(limit, offset int) ([]*SystemEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, type, data, created_at FROM system_events ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventRows(rows)
}

func (r *systemEventRepoSQLite) ListByType(eventType string, limit, offset int) ([]*SystemEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, type, data, created_at FROM system_events WHERE type = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		eventType, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventRows(rows)
}

func (r *systemEventRepoSQLite) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM system_events`).Scan(&count)
	return count, err
}

func (r *systemEventRepoSQLite) CountByType(eventType string) (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM system_events WHERE type = ?`, eventType).Scan(&count)
	return count, err
}

// --- Postgres ---

type systemEventRepoPostgres struct {
	db *sql.DB
}

func (r *systemEventRepoPostgres) Create(eventType string, data interface{}) error {
	var dataJSON []byte
	if data != nil {
		var err error
		dataJSON, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	_, err := r.db.Exec(
		`INSERT INTO system_events (type, data) VALUES ($1, $2)`,
		eventType, dataJSON,
	)
	return err
}

func (r *systemEventRepoPostgres) GetByID(id ID) (*SystemEvent, error) {
	e := &SystemEvent{}
	var data []byte
	err := r.db.QueryRow(
		`SELECT id, type, data, created_at FROM system_events WHERE id = $1`,
		id.String(),
	).Scan(&e.ID, &e.Type, &data, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		e.Data = data
	}
	return e, nil
}

func (r *systemEventRepoPostgres) List(limit, offset int) ([]*SystemEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, type, data, created_at FROM system_events ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventRows(rows)
}

func (r *systemEventRepoPostgres) ListByType(eventType string, limit, offset int) ([]*SystemEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, type, data, created_at FROM system_events WHERE type = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		eventType, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventRows(rows)
}

func (r *systemEventRepoPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM system_events`).Scan(&count)
	return count, err
}

func (r *systemEventRepoPostgres) CountByType(eventType string) (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM system_events WHERE type = $1`, eventType).Scan(&count)
	return count, err
}
