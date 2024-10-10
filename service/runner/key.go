// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/google/uuid"
)

// Key generates public/private key pairs and random UUIDs.
type Key struct{}

// Pair holds a PEM-encoded public and private key (RSA).
type Pair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// GenerateKeyPair creates a new RSA public/private key pair (PEM-encoded).
func (k *Key) GenerateKeyPair() (*Pair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("key: generate key pair: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("key: marshal private key: %w", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("key: marshal public key: %w", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return &Pair{
		PublicKey:  publicPEM,
		PrivateKey: privatePEM,
	}, nil
}

// NewUUID returns a new random UUID string
func (k *Key) NewUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("key: generate uuid: %w", err)
	}
	return id.String(), nil
}
