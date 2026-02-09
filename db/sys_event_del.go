// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// SystemEventDeliverable is one row in system_events_deliverables (event to deliver to a system integration).
type SystemEventDeliverable struct {
	ID            ID
	EventID       ID
	IntegrationID ID
	Status        string
	DeliveredAt   *time.Time
	ErrorMessage  *string
	RetryCount    int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SystemEventsDeliverableRepository is the interface for system_events_deliverables persistence.
type SystemEventDeliverableRepository interface {
	Create(eventID, integrationID ID) error
	GetByID(id ID) (*SystemEventDeliverable, error)
	ListPending(limit int) ([]*SystemEventDeliverable, error)
	ListByEventID(eventID ID, limit, offset int) ([]*SystemEventDeliverable, error)
	ListByIntegrationID(integrationID ID, limit, offset int) ([]*SystemEventDeliverable, error)
	UpdateStatus(id ID, status string, deliveredAt *time.Time, errorMessage *string, retryCount int) error
}

// NewSystemEventsDeliverableRepository returns the repository for the current driver.
func NewSystemEventsDeliverableRepository(db *sql.DB) SystemEventDeliverableRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &systemEventDeliverableRepoPostgres{db: db}
	}
	return &systemEventDeliverableRepoSQLite{db: db}
}

func scanSystemEventDeliverableRows(rows *sql.Rows) ([]*SystemEventDeliverable, error) {
	var list []*SystemEventDeliverable
	for rows.Next() {
		d := &SystemEventDeliverable{}
		var deliveredAt sql.NullTime
		var errorMessage sql.NullString
		if err := rows.Scan(
			&d.ID,
			&d.EventID,
			&d.IntegrationID,
			&d.Status,
			&deliveredAt,
			&errorMessage,
			&d.RetryCount,
			&d.CreatedAt,
			&d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if deliveredAt.Valid {
			d.DeliveredAt = &deliveredAt.Time
		}
		if errorMessage.Valid {
			d.ErrorMessage = &errorMessage.String
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

// --- SQLite ---

type systemEventDeliverableRepoSQLite struct {
	db *sql.DB
}

func (r *systemEventDeliverableRepoSQLite) Create(eventID, integrationID ID) error {
	_, err := r.db.Exec(
		`INSERT INTO system_events_deliverables (event_id, integration_id, status)
		VALUES (?, ?, 'pending')`,
		eventID.String(),
		integrationID.String(),
	)
	return err
}

func (r *systemEventDeliverableRepoSQLite) GetByID(id ID) (*SystemEventDeliverable, error) {
	d := &SystemEventDeliverable{}
	var deliveredAt sql.NullTime
	var errorMessage sql.NullString
	err := r.db.QueryRow(
		`SELECT
			id, event_id, integration_id, status, delivered_at,
			error_message, retry_count, created_at, updated_at
		FROM system_events_deliverables
		WHERE id = ?`,
		id.String(),
	).Scan(
		&d.ID,
		&d.EventID,
		&d.IntegrationID,
		&d.Status,
		&deliveredAt,
		&errorMessage,
		&d.RetryCount,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if deliveredAt.Valid {
		d.DeliveredAt = &deliveredAt.Time
	}
	if errorMessage.Valid {
		d.ErrorMessage = &errorMessage.String
	}
	return d, nil
}

func (r *systemEventDeliverableRepoSQLite) ListPending(limit int) ([]*SystemEventDeliverable, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(
		`SELECT
			id, event_id, integration_id, status, delivered_at,
			error_message, retry_count, created_at, updated_at
		FROM system_events_deliverables
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventDeliverableRows(rows)
}

func (r *systemEventDeliverableRepoSQLite) ListByEventID(eventID ID, limit, offset int) ([]*SystemEventDeliverable, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT
			id, event_id, integration_id, status, delivered_at,
			error_message, retry_count, created_at, updated_at
		FROM system_events_deliverables
		WHERE event_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		eventID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventDeliverableRows(rows)
}

func (r *systemEventDeliverableRepoSQLite) ListByIntegrationID(integrationID ID, limit, offset int) ([]*SystemEventDeliverable, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT
			id, event_id, integration_id, status, delivered_at,
			error_message, retry_count, created_at, updated_at
		FROM system_events_deliverables
		WHERE integration_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		integrationID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventDeliverableRows(rows)
}

func (r *systemEventDeliverableRepoSQLite) UpdateStatus(id ID, status string, deliveredAt *time.Time, errorMessage *string, retryCount int) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE system_events_deliverables
		SET
			status = ?,
			delivered_at = ?,
			error_message = ?,
			retry_count = ?,
			updated_at = ?
		WHERE id = ?`,
		status,
		deliveredAt,
		nullStrPtr(errorMessage),
		retryCount,
		now,
		id.String(),
	)
	return err
}

// --- Postgres ---

type systemEventDeliverableRepoPostgres struct {
	db *sql.DB
}

func (r *systemEventDeliverableRepoPostgres) Create(eventID, integrationID ID) error {
	_, err := r.db.Exec(
		`INSERT INTO system_events_deliverables (event_id, integration_id, status)
		VALUES ($1, $2, 'pending')`,
		eventID.String(),
		integrationID.String(),
	)
	return err
}

func (r *systemEventDeliverableRepoPostgres) GetByID(id ID) (*SystemEventDeliverable, error) {
	d := &SystemEventDeliverable{}
	var deliveredAt sql.NullTime
	var errorMessage sql.NullString
	err := r.db.QueryRow(
		`SELECT
			id, event_id, integration_id, status, delivered_at,
			error_message, retry_count, created_at, updated_at
		FROM system_events_deliverables
		WHERE id = $1`,
		id.String(),
	).Scan(
		&d.ID,
		&d.EventID,
		&d.IntegrationID,
		&d.Status,
		&deliveredAt,
		&errorMessage,
		&d.RetryCount,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if deliveredAt.Valid {
		d.DeliveredAt = &deliveredAt.Time
	}
	if errorMessage.Valid {
		d.ErrorMessage = &errorMessage.String
	}
	return d, nil
}

func (r *systemEventDeliverableRepoPostgres) ListPending(limit int) ([]*SystemEventDeliverable, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(
		`SELECT
			id, event_id, integration_id, status, delivered_at,
			error_message, retry_count, created_at, updated_at
		FROM system_events_deliverables
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventDeliverableRows(rows)
}

func (r *systemEventDeliverableRepoPostgres) ListByEventID(eventID ID, limit, offset int) ([]*SystemEventDeliverable, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT
			id, event_id, integration_id, status, delivered_at,
			error_message, retry_count, created_at, updated_at
		FROM system_events_deliverables
		WHERE event_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		eventID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventDeliverableRows(rows)
}

func (r *systemEventDeliverableRepoPostgres) ListByIntegrationID(integrationID ID, limit, offset int) ([]*SystemEventDeliverable, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT
			id, event_id, integration_id, status, delivered_at,
			error_message, retry_count, created_at, updated_at
		FROM system_events_deliverables
		WHERE integration_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		integrationID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSystemEventDeliverableRows(rows)
}

func (r *systemEventDeliverableRepoPostgres) UpdateStatus(id ID, status string, deliveredAt *time.Time, errorMessage *string, retryCount int) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE system_events_deliverables
		SET
			status = $1,
			delivered_at = $2,
			error_message = $3,
			retry_count = $4,
			updated_at = $5
		WHERE id = $6`,
		status,
		deliveredAt,
		nullStrPtr(errorMessage),
		retryCount,
		now,
		id.String(),
	)
	return err
}
