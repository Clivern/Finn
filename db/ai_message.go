// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// AIMessage is the DB row for one AI chat message.
type AIMessage struct {
	ID        ID
	UserID    ID
	Role      string // e.g. "user", "assistant", "system"
	Content   string
	CreatedAt time.Time
}

// AIMessageRepository is the interface for AI message CRUD.
type AIMessageRepository interface {
	Create(msg *AIMessage) error
	GetByID(id ID) (*AIMessage, error)
	GetByUserID(userID ID, limit, offset int) ([]*AIMessage, error)
	Delete(id ID) error
	DeleteByUserID(userID ID) error
	CountByUserID(userID ID) (int64, error)
}

// NewAIMessageRepository returns the repository for the current driver.
func NewAIMessageRepository(db *sql.DB) AIMessageRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &aiMessageRepoPostgres{db: db}
	}
	return &aiMessageRepoSQLite{db: db}
}

func scanAIMessageRows(rows *sql.Rows) ([]*AIMessage, error) {
	var list []*AIMessage
	for rows.Next() {
		m := &AIMessage{}
		if err := rows.Scan(
			&m.ID,
			&m.UserID,
			&m.Role,
			&m.Content,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

// --- SQLite ---

type aiMessageRepoSQLite struct {
	db *sql.DB
}

func (r *aiMessageRepoSQLite) Create(msg *AIMessage) error {
	result, err := r.db.Exec(
		`INSERT INTO user_ai_messages (user_id, role, content)
		VALUES (?, ?, ?)`,
		msg.UserID.String(),
		msg.Role,
		msg.Content,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	msg.ID = ID(strconv.FormatInt(lid, 10))
	row := r.db.QueryRow(
		`SELECT created_at
		FROM user_ai_messages
		WHERE id = ?`,
		msg.ID.String(),
	)
	return row.Scan(&msg.CreatedAt)
}

func (r *aiMessageRepoSQLite) GetByID(id ID) (*AIMessage, error) {
	m := &AIMessage{}
	err := r.db.QueryRow(
		`SELECT
			id, user_id, role, content, created_at
		FROM user_ai_messages
		WHERE id = ?`,
		id.String(),
	).Scan(
		&m.ID,
		&m.UserID,
		&m.Role,
		&m.Content,
		&m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *aiMessageRepoSQLite) GetByUserID(userID ID, limit, offset int) ([]*AIMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(
		`SELECT
			id, user_id, role, content, created_at
		FROM user_ai_messages
		WHERE user_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?`,
		userID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAIMessageRows(rows)
}

func (r *aiMessageRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM user_ai_messages WHERE id = ?`,
		id.String(),
	)
	return err
}

func (r *aiMessageRepoSQLite) DeleteByUserID(userID ID) error {
	_, err := r.db.Exec(
		`DELETE FROM user_ai_messages WHERE user_id = ?`,
		userID.String(),
	)
	return err
}

func (r *aiMessageRepoSQLite) CountByUserID(userID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_ai_messages
		WHERE user_id = ?`,
		userID.String(),
	).Scan(&count)
	return count, err
}

// --- Postgres ---

type aiMessageRepoPostgres struct {
	db *sql.DB
}

func (r *aiMessageRepoPostgres) Create(msg *AIMessage) error {
	return r.db.QueryRow(
		`INSERT INTO user_ai_messages (user_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		msg.UserID.String(),
		msg.Role,
		msg.Content,
	).Scan(
		&msg.ID,
		&msg.CreatedAt,
	)
}

func (r *aiMessageRepoPostgres) GetByID(id ID) (*AIMessage, error) {
	m := &AIMessage{}
	err := r.db.QueryRow(
		`SELECT
			id, user_id, role, content, created_at
		FROM user_ai_messages
		WHERE id = $1`,
		id.String(),
	).Scan(
		&m.ID,
		&m.UserID,
		&m.Role,
		&m.Content,
		&m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *aiMessageRepoPostgres) GetByUserID(userID ID, limit, offset int) ([]*AIMessage, error) {
	rows, err := r.db.Query(
		`SELECT
			id, user_id, role, content, created_at
		FROM user_ai_messages
		WHERE user_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`,
		userID.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAIMessageRows(rows)
}

func (r *aiMessageRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM user_ai_messages WHERE id = $1`,
		id.String(),
	)
	return err
}

func (r *aiMessageRepoPostgres) DeleteByUserID(userID ID) error {
	_, err := r.db.Exec(
		`DELETE FROM user_ai_messages WHERE user_id = $1`,
		userID.String(),
	)
	return err
}

func (r *aiMessageRepoPostgres) CountByUserID(userID ID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_ai_messages
		WHERE user_id = $1`,
		userID.String(),
	).Scan(&count)
	return count, err
}
