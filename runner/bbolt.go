// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"fmt"
	"os"

	bbolt "go.etcd.io/bbolt"
)

const defaultBoltMode = 0o600

// Bolt wraps go.etcd.io/bbolt for opening and accessing a Bolt DB file
type Bolt struct {
	Path string
	db   *bbolt.DB
}

// Open opens the Bolt database at Path
func (b *Bolt) Open() error {
	if b.Path == "" {
		return fmt.Errorf("bbolt: path is required")
	}
	db, err := bbolt.Open(b.Path, os.FileMode(defaultBoltMode), nil)
	if err != nil {
		return fmt.Errorf("bbolt open: %w", err)
	}
	b.db = db
	return nil
}

// Close closes the database
func (b *Bolt) Close() error {
	if b.db == nil {
		return nil
	}
	err := b.db.Close()
	b.db = nil
	if err != nil {
		return fmt.Errorf("bbolt close: %w", err)
	}
	return nil
}

// DB returns the underlying *bbolt.DB
func (b *Bolt) DB() *bbolt.DB {
	return b.db
}

// Get reads the value for key in the given bucket
func (b *Bolt) Get(bucket, key []byte) ([]byte, error) {
	if b.db == nil {
		return nil, fmt.Errorf("bbolt: database not open")
	}
	var out []byte
	err := b.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(bucket)
		if bkt == nil {
			return nil
		}
		v := bkt.Get(key)
		if v != nil {
			out = make([]byte, len(v))
			copy(out, v)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("bbolt get: %w", err)
	}
	return out, nil
}

// Put writes key/value into the given bucket
func (b *Bolt) Put(bucket, key, value []byte) error {
	if b.db == nil {
		return fmt.Errorf("bbolt: database not open")
	}
	err := b.db.Update(func(tx *bbolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return bkt.Put(key, value)
	})
	if err != nil {
		return fmt.Errorf("bbolt put: %w", err)
	}
	return nil
}

// Delete removes key from the given bucket
func (b *Bolt) Delete(bucket, key []byte) error {
	if b.db == nil {
		return fmt.Errorf("bbolt: database not open")
	}
	err := b.db.Update(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(bucket)
		if bkt == nil {
			return nil
		}
		return bkt.Delete(key)
	})
	if err != nil {
		return fmt.Errorf("bbolt delete: %w", err)
	}
	return nil
}

// CreateBucket creates a bucket with the given name
func (b *Bolt) CreateBucket(name []byte) error {
	if b.db == nil {
		return fmt.Errorf("bbolt: database not open")
	}
	err := b.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(name)
		return err
	})
	if err != nil {
		return fmt.Errorf("bbolt create bucket: %w", err)
	}
	return nil
}

// DeleteBucket deletes the bucket and all its keys
func (b *Bolt) DeleteBucket(name []byte) error {
	if b.db == nil {
		return fmt.Errorf("bbolt: database not open")
	}
	err := b.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket(name)
	})
	if err != nil {
		return fmt.Errorf("bbolt delete bucket: %w", err)
	}
	return nil
}

// ListKeys returns all keys in the given bucket in order
func (b *Bolt) ListKeys(bucket []byte) ([][]byte, error) {
	if b.db == nil {
		return nil, fmt.Errorf("bbolt: database not open")
	}
	var keys [][]byte
	err := b.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(bucket)
		if bkt == nil {
			return nil
		}
		return bkt.ForEach(func(k, _ []byte) error {
			cp := make([]byte, len(k))
			copy(cp, k)
			keys = append(keys, cp)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("bbolt list keys: %w", err)
	}
	return keys, nil
}

// Update runs fn inside a read-write transaction
func (b *Bolt) Update(fn func(tx *bbolt.Tx) error) error {
	if b.db == nil {
		return fmt.Errorf("bbolt: database not open")
	}
	err := b.db.Update(fn)
	if err != nil {
		return fmt.Errorf("bbolt update: %w", err)
	}
	return nil
}

// View runs fn inside a read-only transaction
func (b *Bolt) View(fn func(tx *bbolt.Tx) error) error {
	if b.db == nil {
		return fmt.Errorf("bbolt: database not open")
	}
	err := b.db.View(fn)
	if err != nil {
		return fmt.Errorf("bbolt view: %w", err)
	}
	return nil
}
