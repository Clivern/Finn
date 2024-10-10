// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// Encrypter encrypts and decrypts data using the runner's RSA public and private keys
type Encrypter struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewEncrypter builds an Encrypter from the runner
func NewEncrypter(identity *Runner) (*Encrypter, error) {
	if identity == nil {
		return nil, fmt.Errorf("encrypt: identity is required")
	}
	if len(identity.PublicKey) == 0 || len(identity.PrivateKey) == 0 {
		return nil, fmt.Errorf("encrypt: identity has no keys")
	}

	pubBlock, _ := pem.Decode(identity.PublicKey)
	if pubBlock == nil {
		return nil, fmt.Errorf("encrypt: invalid public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("encrypt: public key is not RSA")
	}

	privBlock, _ := pem.Decode(identity.PrivateKey)
	if privBlock == nil {
		return nil, fmt.Errorf("encrypt: invalid private key PEM")
	}
	priv, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse private key: %w", err)
	}
	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("encrypt: private key is not RSA")
	}

	return &Encrypter{
		publicKey:  rsaPub,
		privateKey: rsaPriv,
	}, nil
}

// Encrypt encrypts plaintext with the runner's public key
func (e *Encrypter) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, plaintext, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}
	return ciphertext, nil
}

// Decrypt decrypts ciphertext with the runner's private key
func (e *Encrypter) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, e.privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

// EncryptString encrypts a UTF-8 string and returns the ciphertext
func (e *Encrypter) EncryptString(plaintext string) ([]byte, error) {
	return e.Encrypt([]byte(plaintext))
}

// DecryptString decrypts ciphertext and returns the UTF-8 string
func (e *Encrypter) DecryptString(ciphertext []byte) (string, error) {
	b, err := e.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
