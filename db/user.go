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
	UserRoleUser     = "user"
	UserRoleReadonly = "readonly"
)

// User represents a user in the database.
type User struct {
	ID          ID
	Email       string
	Password    string
	Role        string
	APIKey      string
	IsActive    bool
	LastLoginAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserRepository handles database operations for users.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(user *User) error {
	driver := GetDriver()
	q := Placeholder(`INSERT INTO users (email, password, role, api_key, is_active, last_login_at)
		VALUES (?, ?, ?, ?, ?, ?)`, driver)
	if driver == "postgres" || driver == "postgresql" {
		q = `INSERT INTO users (email, password, role, api_key, is_active, last_login_at)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
		err := r.db.QueryRow(q,
			user.Email,
			user.Password,
			user.Role,
			user.APIKey,
			user.IsActive,
			user.LastLoginAt,
		).Scan(&user.ID)
		return err
	}
	result, err := r.db.Exec(q,
		user.Email,
		user.Password,
		user.Role,
		user.APIKey,
		user.IsActive,
		user.LastLoginAt,
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

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(id ID) (*User, error) {
	user := &User{}
	q := Placeholder(`SELECT id, email, password, role, api_key, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE id = ?`, GetDriver())
	err := r.db.QueryRow(q, id.String()).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.APIKey,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(email string) (*User, error) {
	user := &User{}
	q := Placeholder(`SELECT id, email, password, role, api_key, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE email = ?`, GetDriver())
	err := r.db.QueryRow(q, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.APIKey,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByAPIKey retrieves a user by API key.
func (r *UserRepository) GetByAPIKey(apiKey string) (*User, error) {
	user := &User{}
	q := Placeholder(`SELECT id, email, password, role, api_key, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE api_key = ?`, GetDriver())
	err := r.db.QueryRow(q, apiKey).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.APIKey,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates a user's information.
func (r *UserRepository) Update(user *User) error {
	q := Placeholder(`UPDATE users SET
			email = ?, password = ?, role = ?, api_key = ?, is_active = ?,
			last_login_at = ?, updated_at = ?
		WHERE id = ?`, GetDriver())
	_, err := r.db.Exec(q,
		user.Email,
		user.Password,
		user.Role,
		user.APIKey,
		user.IsActive,
		user.LastLoginAt,
		time.Now().UTC(),
		user.ID.String(),
	)
	return err
}

// UpdateLastLogin updates the user's last login timestamp.
func (r *UserRepository) UpdateLastLogin(id ID) error {
	now := time.Now().UTC()
	q := Placeholder(`UPDATE users SET
			last_login_at = ?, updated_at = ?
		WHERE id = ?`, GetDriver())
	_, err := r.db.Exec(q, now, now, id.String())
	return err
}

// Delete removes a user from the database.
func (r *UserRepository) Delete(id ID) error {
	q := Placeholder("DELETE FROM users WHERE id = ?", GetDriver())
	_, err := r.db.Exec(q, id.String())
	return err
}

// List retrieves all users with pagination.
func (r *UserRepository) List(limit, offset int) ([]*User, error) {
	q := Placeholder(`SELECT id, email, password, role, api_key, is_active, last_login_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, GetDriver())
	rows, err := r.db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.APIKey,
			&user.IsActive,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// Count returns the total number of users.
func (r *UserRepository) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// UserMeta represents metadata associated with a user.
type UserMeta struct {
	ID        ID
	Key       string
	Value     string
	UserID    ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserMetaRepository handles database operations for user metadata.
type UserMetaRepository struct {
	db *sql.DB
}

// NewUserMetaRepository creates a new user meta repository.
func NewUserMetaRepository(db *sql.DB) *UserMetaRepository {
	return &UserMetaRepository{db: db}
}

// Create inserts new metadata for a user.
func (r *UserMetaRepository) Create(userID ID, key, value string) error {
	q := Placeholder("INSERT INTO users_meta (user_id, key, value) VALUES (?, ?, ?)", GetDriver())
	_, err := r.db.Exec(q, userID.String(), key, value)
	return err
}

// Get retrieves metadata for a user by key.
func (r *UserMetaRepository) Get(userID ID, key string) (*UserMeta, error) {
	meta := &UserMeta{}
	q := Placeholder(`SELECT id, key, value, user_id, created_at, updated_at
		FROM users_meta
		WHERE user_id = ? AND key = ?`, GetDriver())
	err := r.db.QueryRow(q, userID.String(), key).Scan(
		&meta.ID,
		&meta.Key,
		&meta.Value,
		&meta.UserID,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// Update updates metadata for a user.
func (r *UserMetaRepository) Update(userID ID, key, value string) error {
	q := Placeholder(`UPDATE users_meta SET
			value = ?, updated_at = ?
		WHERE user_id = ? AND key = ?`, GetDriver())
	_, err := r.db.Exec(q, value, time.Now().UTC(), userID.String(), key)
	return err
}

// Delete removes metadata for a user.
func (r *UserMetaRepository) Delete(userID ID, key string) error {
	q := Placeholder("DELETE FROM users_meta WHERE user_id = ? AND key = ?", GetDriver())
	_, err := r.db.Exec(q, userID.String(), key)
	return err
}

// ListByUser retrieves all metadata for a user.
func (r *UserMetaRepository) ListByUser(userID ID) ([]*UserMeta, error) {
	q := Placeholder(`SELECT id, key, value, user_id, created_at, updated_at
		FROM users_meta
		WHERE user_id = ?
		ORDER BY key`, GetDriver())
	rows, err := r.db.Query(q, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metadata []*UserMeta
	for rows.Next() {
		meta := &UserMeta{}
		if err := rows.Scan(
			&meta.ID,
			&meta.Key,
			&meta.Value,
			&meta.UserID,
			&meta.CreatedAt,
			&meta.UpdatedAt,
		); err != nil {
			return nil, err
		}
		metadata = append(metadata, meta)
	}

	return metadata, rows.Err()
}

// Upsert inserts or updates metadata for a user.
func (r *UserMetaRepository) Upsert(userID ID, key, value string) error {
	existing, err := r.Get(userID, key)
	if err != nil {
		return err
	}

	if existing == nil {
		return r.Create(userID, key, value)
	}

	return r.Update(userID, key, value)
}
