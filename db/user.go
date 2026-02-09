// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"strconv"
	"time"
)

// User role constants
const (
	UserRoleAdmin    = "admin"
	UserRoleRegular  = "regular"
	UserRoleOwner    = "owner"
	UserRoleReadOnly = "readonly"
)

// User is the DB row for a user.
type User struct {
	ID          ID
	Name        string
	Email       string
	Password    string
	Role        string
	IsActive    bool
	LastLoginAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserRepository is the interface for user CRUD.
type UserRepository interface {
	Create(user *User) error
	GetByID(id ID) (*User, error)
	GetByAPIKey(apiKey string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	UpdateLastLogin(id ID) error
	Delete(id ID) error
	List(limit, offset int) ([]*User, error)
	Count() (int64, error)
}

// NewUserRepository returns the repository for the current driver
func NewUserRepository(db *sql.DB) UserRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &userRepoPostgres{db: db}
	}
	return &userRepoSQLite{db: db}
}

// scanUserRows builds a slice of User from query rows.
func scanUserRows(rows *sql.Rows) ([]*User, error) {
	var list []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.Password, &u.Role,
			&u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

// --- SQLite ---

// userRepoSQLite implements UserRepository for SQLite.
type userRepoSQLite struct {
	db *sql.DB
}

// Create inserts a row
func (r *userRepoSQLite) Create(user *User) error {
	result, err := r.db.Exec(
		`INSERT INTO users
		(name, email, password_hash, role, is_active, last_login_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		user.Name, user.Email, user.Password, user.Role, user.IsActive, user.LastLoginAt,
	)
	if err != nil {
		return err
	}
	lid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = ID(strconv.FormatInt(lid, 10))
	return nil
}

// GetByID returns a user by ID
func (r *userRepoSQLite) GetByID(id ID) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT
			id, name, email, password_hash, role, is_active,
			last_login_at, created_at, updated_at
		FROM users
		WHERE id = ?`,
		id.String(),
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByAPIKey returns a user by API key (raw key lookup)
func (r *userRepoSQLite) GetByAPIKey(apiKey string) (*User, error) {
	if apiKey == "" {
		return nil, nil
	}
	user := &User{}
	err := r.db.QueryRow(
		`SELECT
			u.id, u.name, u.email, u.password_hash, u.role,
			u.is_active, u.last_login_at, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_api_keys k ON u.id = k.user_id
		WHERE k.key = ? AND (k.expires_at IS NULL OR k.expires_at > CURRENT_TIMESTAMP)`,
		apiKey,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail returns a user by email
func (r *userRepoSQLite) GetByEmail(email string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT
			id, name, email, password_hash, role, is_active,
			last_login_at, created_at, updated_at
		FROM users
		WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates a user
func (r *userRepoSQLite) Update(user *User) error {
	_, err := r.db.Exec(
		`UPDATE users
		SET
			name = ?,
			email = ?,
			password_hash = ?,
			role = ?,
			is_active = ?,
			last_login_at = ?,
			updated_at = ?
		WHERE id = ?`,
		user.Name, user.Email, user.Password, user.Role, user.IsActive,
		user.LastLoginAt, time.Now().UTC(), user.ID.String(),
	)
	return err
}

// UpdateLastLogin updates the last login timestamp
func (r *userRepoSQLite) UpdateLastLogin(id ID) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE users
		SET
			last_login_at = ?,
			updated_at = ?
		WHERE id = ?`,
		now, now, id.String(),
	)
	return err
}

// Delete removes a user
func (r *userRepoSQLite) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM users WHERE id = ?`,
		id.String(),
	)
	return err
}

// List returns a list of users
func (r *userRepoSQLite) List(limit, offset int) ([]*User, error) {
	rows, err := r.db.Query(
		`SELECT
			id, name, email, password_hash, role, is_active,
			last_login_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRows(rows)
}

// Count returns the total number of users
func (r *userRepoSQLite) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM users`,
	).Scan(&count)
	return count, err
}

// --- Postgres ---

// userRepoPostgres implements UserRepository for PostgreSQL
type userRepoPostgres struct {
	db *sql.DB
}

// Create inserts a row
func (r *userRepoPostgres) Create(user *User) error {
	err := r.db.QueryRow(
		`INSERT INTO users
		(name, email, password_hash, role, is_active, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		user.Name, user.Email, user.Password, user.Role, user.IsActive, user.LastLoginAt,
	).Scan(&user.ID)
	return err
}

// GetByID returns a user by ID
func (r *userRepoPostgres) GetByID(id ID) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT
			id, name, email, password_hash, role, is_active,
			last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1`,
		id.String(),
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByAPIKey returns a user by API key (raw key lookup)
func (r *userRepoPostgres) GetByAPIKey(apiKey string) (*User, error) {
	if apiKey == "" {
		return nil, nil
	}
	user := &User{}
	err := r.db.QueryRow(
		`SELECT
			u.id, u.name, u.email, u.password_hash, u.role,
			u.is_active, u.last_login_at, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_api_keys k ON u.id = k.user_id
		WHERE k.key = $1 AND (k.expires_at IS NULL OR k.expires_at > CURRENT_TIMESTAMP)`,
		apiKey,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail returns a user by email
func (r *userRepoPostgres) GetByEmail(email string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT
			id, name, email, password_hash, role, is_active,
			last_login_at, created_at, updated_at
		FROM users
		WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates a user
func (r *userRepoPostgres) Update(user *User) error {
	_, err := r.db.Exec(
		`UPDATE users
		SET
			name = $1,
			email = $2,
			password_hash = $3,
			role = $4,
			is_active = $5,
			last_login_at = $6,
			updated_at = $7
		WHERE id = $8`,
		user.Name, user.Email, user.Password, user.Role, user.IsActive,
		user.LastLoginAt, time.Now().UTC(), user.ID.String(),
	)
	return err
}

// UpdateLastLogin updates the last login timestamp
func (r *userRepoPostgres) UpdateLastLogin(id ID) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE users
		SET
			last_login_at = $1,
			updated_at = $2
		WHERE id = $3`,
		now, now, id.String(),
	)
	return err
}

// Delete removes a user
func (r *userRepoPostgres) Delete(id ID) error {
	_, err := r.db.Exec(
		`DELETE FROM users WHERE id = $1`,
		id.String(),
	)
	return err
}

// List returns a list of users
func (r *userRepoPostgres) List(limit, offset int) ([]*User, error) {
	rows, err := r.db.Query(
		`SELECT
			id, name, email, password_hash, role, is_active,
			last_login_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRows(rows)
}

// Count returns the total number of users
func (r *userRepoPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM users`,
	).Scan(&count)
	return count, err
}

// UserMeta is extra metadata for a user.
type UserMeta struct {
	ID        ID
	Key       string
	Value     string
	UserID    ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserMetaRepository is the interface for user metadata CRUD.
type UserMetaRepository interface {
	Create(userID ID, key, value string) error
	Get(userID ID, key string) (*UserMeta, error)
	Update(userID ID, key, value string) error
	Delete(userID ID, key string) error
	ListByUser(userID ID) ([]*UserMeta, error)
	Upsert(userID ID, key, value string) error
}

// NewUserMetaRepository returns the repository for the current driver
func NewUserMetaRepository(db *sql.DB) UserMetaRepository {
	driver := GetDriver()
	if driver == "postgres" || driver == "postgresql" {
		return &userMetaRepoPostgres{db: db}
	}
	return &userMetaRepoSQLite{db: db}
}

// scanUserMetaRows builds a slice of UserMeta from query rows.
func scanUserMetaRows(rows *sql.Rows) ([]*UserMeta, error) {
	var list []*UserMeta
	for rows.Next() {
		m := &UserMeta{}
		if err := rows.Scan(&m.ID, &m.Key, &m.Value, &m.UserID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

// --- UserMeta SQLite ---

// userMetaRepoSQLite implements UserMetaRepository for SQLite.
type userMetaRepoSQLite struct {
	db *sql.DB
}

// Create inserts a user meta
func (r *userMetaRepoSQLite) Create(userID ID, key, value string) error {
	_, err := r.db.Exec(
		`INSERT INTO users_meta (user_id, key, value)
		VALUES (?, ?, ?)`,
		userID.String(), key, value,
	)
	return err
}

// Get returns a user meta
func (r *userMetaRepoSQLite) Get(userID ID, key string) (*UserMeta, error) {
	meta := &UserMeta{}
	err := r.db.QueryRow(
		`SELECT id, key, value, user_id, created_at, updated_at
		FROM users_meta
		WHERE user_id = ? AND key = ?`,
		userID.String(), key,
	).Scan(&meta.ID, &meta.Key, &meta.Value, &meta.UserID, &meta.CreatedAt, &meta.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// Update updates a user meta
func (r *userMetaRepoSQLite) Update(userID ID, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE users_meta
		SET
			value = ?,
			updated_at = ?
		WHERE user_id = ? AND key = ?`,
		value, time.Now().UTC(), userID.String(), key,
	)
	return err
}

// Delete removes a user meta
func (r *userMetaRepoSQLite) Delete(userID ID, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM users_meta WHERE user_id = ? AND key = ?`,
		userID.String(), key,
	)
	return err
}

// ListByUser returns a list of user meta
func (r *userMetaRepoSQLite) ListByUser(userID ID) ([]*UserMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, key, value, user_id, created_at, updated_at
		FROM users_meta
		WHERE user_id = ?
		ORDER BY key`,
		userID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserMetaRows(rows)
}

// Upsert upserts a user meta
func (r *userMetaRepoSQLite) Upsert(userID ID, key, value string) error {
	existing, err := r.Get(userID, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(userID, key, value)
	}
	return r.Update(userID, key, value)
}

// --- UserMeta Postgres ---

// userMetaRepoPostgres implements UserMetaRepository for PostgreSQL
type userMetaRepoPostgres struct {
	db *sql.DB
}

// Create inserts a user meta
func (r *userMetaRepoPostgres) Create(userID ID, key, value string) error {
	_, err := r.db.Exec(
		`INSERT INTO users_meta (user_id, key, value)
		VALUES ($1, $2, $3)`,
		userID.String(), key, value,
	)
	return err
}

// Get returns a user meta
func (r *userMetaRepoPostgres) Get(userID ID, key string) (*UserMeta, error) {
	meta := &UserMeta{}
	err := r.db.QueryRow(
		`SELECT id, key, value, user_id, created_at, updated_at
		FROM users_meta
		WHERE user_id = $1 AND key = $2`,
		userID.String(), key,
	).Scan(&meta.ID, &meta.Key, &meta.Value, &meta.UserID, &meta.CreatedAt, &meta.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// Update updates a user meta
func (r *userMetaRepoPostgres) Update(userID ID, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE users_meta
		SET
			value = $1,
			updated_at = $2
		WHERE user_id = $3 AND key = $4`,
		value, time.Now().UTC(), userID.String(), key,
	)
	return err
}

// Delete removes a user meta
func (r *userMetaRepoPostgres) Delete(userID ID, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM users_meta WHERE user_id = $1 AND key = $2`,
		userID.String(), key,
	)
	return err
}

// ListByUser returns a list of user meta
func (r *userMetaRepoPostgres) ListByUser(userID ID) ([]*UserMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, key, value, user_id, created_at, updated_at
		FROM users_meta
		WHERE user_id = $1
		ORDER BY key`,
		userID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserMetaRows(rows)
}

// Upsert upserts a user meta
func (r *userMetaRepoPostgres) Upsert(userID ID, key, value string) error {
	existing, err := r.Get(userID, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(userID, key, value)
	}
	return r.Update(userID, key, value)
}
