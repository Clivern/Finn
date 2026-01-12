// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package util

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("secret data and more")
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(encrypted) <= len(plaintext) {
		t.Errorf("expected ciphertext longer than plaintext")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptWrongKeySize(t *testing.T) {
	_, err := Encrypt([]byte("x"), []byte("short"))
	if err != ErrKeySize {
		t.Errorf("Encrypt with short key: got %v, want ErrKeySize", err)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key := make([]byte, 32)
	encrypted, _ := Encrypt([]byte("secret"), key)
	wrongKey := make([]byte, 32)
	wrongKey[0] = 1

	_, err := Decrypt(encrypted, wrongKey)
	if err != ErrDecrypt {
		t.Errorf("Decrypt with wrong key: got %v, want ErrDecrypt", err)
	}
}
