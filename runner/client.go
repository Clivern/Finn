// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package runner

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client holds the runner base URL, public key and token for pinging.
type Client struct {
	BaseURL      string
	PublicKeyPEM []byte
	Token        string
}

// Ping sends a ping to the runner using the client's base URL, public key and token.
func (c *Client) Ping(ctx context.Context) error {
	if c.BaseURL == "" {
		return fmt.Errorf("ping: BaseURL is required")
	}
	if len(c.PublicKeyPEM) == 0 {
		return fmt.Errorf("ping: PublicKeyPEM is required")
	}
	if c.Token == "" {
		return fmt.Errorf("ping: Token is required")
	}

	pubBlock, _ := pem.Decode(c.PublicKeyPEM)
	if pubBlock == nil {
		return fmt.Errorf("ping: invalid public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return fmt.Errorf("ping: parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("ping: public key is not RSA")
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, []byte("PONG"), nil)
	if err != nil {
		return fmt.Errorf("ping: encrypt: %w", err)
	}
	body := base64.StdEncoding.EncodeToString(ciphertext)

	url := strings.TrimSuffix(c.BaseURL, "/") + "/_ping"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("ping: request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Runner-Token", c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ping: %s: %s", resp.Status, string(msg))
	}

	return nil
}
