// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// Session represents a user session in the database.
type Session struct {
	ID        ID
	Token     string
	UserID    ID
	IPAddress *string
	UserAgent *string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SessionRepository handles database operations for sessions.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create inserts a new session into the database.
func (r *SessionRepository) Create(session *Session) error {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		q := `INSERT INTO sessions (token, user_id, ip_address, user_agent, expires_at)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`
		return r.db.QueryRow(q,
			session.Token,
			session.UserID.String(),
			session.IPAddress,
			session.UserAgent,
			session.ExpiresAt,
		).Scan(&session.ID)
	}
	q := Placeholder(`INSERT INTO sessions (token, user_id, ip_address, user_agent, expires_at)
		VALUES (?, ?, ?, ?, ?)`, driver)
	result, err := r.db.Exec(q,
		session.Token,
		session.UserID.String(),
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	session.ID = ID(strconv.FormatInt(lid, 10))
	return nil
}

// GetByToken retrieves a session by token.
func (r *SessionRepository) GetByToken(token string) (*Session, error) {
	session := &Session{}
	q := Placeholder(`SELECT id, token, user_id, ip_address, user_agent, expires_at, created_at, updated_at
		FROM sessions
		WHERE token = ?`, GetDriver())
	err := r.db.QueryRow(q, token).Scan(
		&session.ID,
		&session.Token,
		&session.UserID,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetByID retrieves a session by ID.
func (r *SessionRepository) GetByID(id ID) (*Session, error) {
	session := &Session{}
	q := Placeholder(`SELECT id, token, user_id, ip_address, user_agent, expires_at, created_at, updated_at
		FROM sessions
		WHERE id = ?`, GetDriver())
	err := r.db.QueryRow(q, id.String()).Scan(
		&session.ID,
		&session.Token,
		&session.UserID,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetByUserID retrieves all sessions for a user.
func (r *SessionRepository) GetByUserID(userID ID) ([]*Session, error) {
	q := Placeholder(`SELECT id, token, user_id, ip_address, user_agent, expires_at, created_at, updated_at
		FROM sessions
		WHERE user_id = ?
		ORDER BY created_at DESC`, GetDriver())
	rows, err := r.db.Query(q, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
		if err := rows.Scan(
			&session.ID,
			&session.Token,
			&session.UserID,
			&session.IPAddress,
			&session.UserAgent,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// Delete removes a session from the database.
func (r *SessionRepository) Delete(id ID) error {
	q := Placeholder("DELETE FROM sessions WHERE id = ?", GetDriver())
	_, err := r.db.Exec(q, id.String())
	return err
}

// DeleteByToken removes a session by token.
func (r *SessionRepository) DeleteByToken(token string) error {
	q := Placeholder("DELETE FROM sessions WHERE token = ?", GetDriver())
	_, err := r.db.Exec(q, token)
	return err
}

// DeleteByUserID removes all sessions for a user.
func (r *SessionRepository) DeleteByUserID(userID ID) error {
	q := Placeholder("DELETE FROM sessions WHERE user_id = ?", GetDriver())
	_, err := r.db.Exec(q, userID.String())
	return err
}

// DeleteExpired removes all expired sessions.
func (r *SessionRepository) DeleteExpired() (int64, error) {
	q := Placeholder("DELETE FROM sessions WHERE expires_at < ?", GetDriver())
	result, err := r.db.Exec(q, time.Now().UTC())
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// IsValid checks if a session exists and is not expired.
func (r *SessionRepository) IsValid(token string) (bool, error) {
	session, err := r.GetByToken(token)
	if err != nil {
		return false, err
	}

	if session == nil {
		return false, nil
	}

	return session.ExpiresAt.After(time.Now().UTC()), nil
}

// UpdateExpiration updates the expiration time of a session.
func (r *SessionRepository) UpdateExpiration(id ID, expiresAt time.Time) error {
	q := Placeholder(`UPDATE sessions SET
			expires_at = ?, updated_at = ?
		WHERE id = ?`, GetDriver())
	_, err := r.db.Exec(q, expiresAt, time.Now().UTC(), id.String())
	return err
}

// Count returns the total number of active (non-expired) sessions.
func (r *SessionRepository) Count() (int64, error) {
	var count int64
	q := Placeholder("SELECT COUNT(*) FROM sessions WHERE expires_at > ?", GetDriver())
	err := r.db.QueryRow(q, time.Now().UTC()).Scan(&count)
	return count, err
}

// CountByUserID returns the number of active sessions for a user.
func (r *SessionRepository) CountByUserID(userID ID) (int64, error) {
	var count int64
	q := Placeholder(`SELECT COUNT(*) FROM sessions WHERE user_id = ? AND expires_at > ?`, GetDriver())
	err := r.db.QueryRow(q, userID.String(), time.Now().UTC()).Scan(&count)
	return count, err
}
