// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// PasswordResetToken represents a password reset token in the database
type PasswordResetToken struct {
	ID        ID
	UserID    ID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// PasswordResetTokenRepository defines password reset token CRUD
type PasswordResetTokenRepository interface {
	Create(t *PasswordResetToken) error
	GetByToken(token string) (*PasswordResetToken, error)
	Delete(id ID) error
	DeleteByToken(token string) error
	DeleteExpired() (int64, error)
}

// NewPasswordResetTokenRepository returns the repository for the current driver
func NewPasswordResetTokenRepository(db *sql.DB) PasswordResetTokenRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &passwordResetTokenRepoPostgres{db: db}
	}
	return &passwordResetTokenRepoSQLite{db: db}
}

// --- SQLite ---

type passwordResetTokenRepoSQLite struct {
	db *sql.DB
}

// Create inserts a password reset token row.
func (r *passwordResetTokenRepoSQLite) Create(t *PasswordResetToken) error {
	result, err := r.db.Exec(
		`INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES (?, ?, ?)`,
		t.UserID.String(), t.Token, t.ExpiresAt,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = ID(strconv.FormatInt(lid, 10))
	return nil
}

// GetByToken returns a token by its string value.
func (r *passwordResetTokenRepoSQLite) GetByToken(token string) (*PasswordResetToken, error) {
	tok := &PasswordResetToken{}
	err := r.db.QueryRow(
		`SELECT id, user_id, token, expires_at, created_at FROM password_reset_tokens WHERE token = ?`,
		token,
	).Scan(&tok.ID, &tok.UserID, &tok.Token, &tok.ExpiresAt, &tok.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// Delete removes a password reset token by ID.
func (r *passwordResetTokenRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE id = ?`, id.String())
	return err
}

// DeleteByToken removes a password reset token by token string.
func (r *passwordResetTokenRepoSQLite) DeleteByToken(token string) error {
	_, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE token = ?`, token)
	return err
}

// DeleteExpired removes expired tokens and returns the count deleted.
func (r *passwordResetTokenRepoSQLite) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE expires_at < ?`, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// --- Postgres ---

type passwordResetTokenRepoPostgres struct {
	db *sql.DB
}

// Create inserts a password reset token row.
func (r *passwordResetTokenRepoPostgres) Create(t *PasswordResetToken) error {
	return r.db.QueryRow(
		`INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3) RETURNING id`,
		t.UserID.String(), t.Token, t.ExpiresAt,
	).Scan(&t.ID)
}

// GetByToken returns a token by its string value.
func (r *passwordResetTokenRepoPostgres) GetByToken(token string) (*PasswordResetToken, error) {
	tok := &PasswordResetToken{}
	err := r.db.QueryRow(
		`SELECT id, user_id, token, expires_at, created_at FROM password_reset_tokens WHERE token = $1`,
		token,
	).Scan(&tok.ID, &tok.UserID, &tok.Token, &tok.ExpiresAt, &tok.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// Delete removes a password reset token by ID.
func (r *passwordResetTokenRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE id = $1`, id.String())
	return err
}

// DeleteByToken removes a password reset token by token string.
func (r *passwordResetTokenRepoPostgres) DeleteByToken(token string) error {
	_, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE token = $1`, token)
	return err
}

// DeleteExpired removes expired tokens and returns the count deleted.
func (r *passwordResetTokenRepoPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE expires_at < $1`, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
