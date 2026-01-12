// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"

	"github.com/spf13/viper"
)

var (
	// ErrKeySize is returned when the key length is not 16, 24, or 32 bytes.
	// Use: openssl rand -hex 32
	ErrKeySize = errors.New("encryption key must be 16, 24, or 32 bytes")
	// ErrDecrypt is returned when decryption fails.
	ErrDecrypt = errors.New("decryption failed")
)

// Encrypt encrypts plaintext with AES-GCM using the given key (16, 24, or 32 bytes).
func Encrypt(plaintext, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrDecrypt
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrDecrypt
	}
	return plaintext, nil
}

// GetEncryptionKey returns the app encryption key from config (16, 24, or 32 bytes).
func GetEncryptionKey() []byte {
	s := viper.GetString("app.encryption_key")
	if s == "" {
		return nil
	}
	decoded, err := hex.DecodeString(s)
	if err == nil && (len(decoded) == 16 || len(decoded) == 24 || len(decoded) == 32) {
		return decoded
	}
	return []byte(s)
}

// EncryptToBase64String encrypts plaintext and returns base64-encoded ciphertext.
func EncryptToBase64String(plain string, key []byte) (string, error) {
	if plain == "" {
		return "", nil
	}

	out, err := Encrypt([]byte(plain), key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptFromBase64String decodes base64 and decrypts. Returns the decrypted string.
func DecryptFromBase64String(b64 string, key []byte) string {
	if b64 == "" || len(key) == 0 {
		return b64
	}

	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return b64
	}

	out, err := Decrypt(raw, key)
	if err != nil {
		return b64
	}
	return string(out)
}
