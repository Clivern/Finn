// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package service

import (
	"crypto/tls"
	"fmt"
	"html"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/clivern/tivo/db"
	"github.com/rs/zerolog/log"
)

// Mailer sends email using SMTP configuration from options.
type Mailer struct {
	opts db.OptionRepository
}

// NewMailer returns a Mailer that reads SMTP config from the option repository.
func NewMailer(opts db.OptionRepository) *Mailer {
	return &Mailer{opts: opts}
}

// getOption returns the option value or empty string.
func (m *Mailer) getOption(key string) string {
	o, err := m.opts.Get(key)
	if err != nil || o == nil {
		return ""
	}
	return strings.TrimSpace(o.Value)
}

// Config returns whether SMTP is configured (server non-empty) and the values needed to send.
func (m *Mailer) Config() (server, from, username, password string, port int, useTLS bool, configured bool) {
	server = m.getOption("smtp_server")
	from = m.getOption("smtp_from_email")
	username = m.getOption("smtp_username")
	password = m.getOption("smtp_password")
	port, _ = strconv.Atoi(m.getOption("smtp_port"))
	useTLS = m.getOption("smtp_use_tls") == "1"
	return server, from, username, password, port, useTLS, true
}

// Send sends a plain-text email. No-op if SMTP is not configured.
func (m *Mailer) Send(to, subject, body string) error {
	server, from, username, password, port, useTLS, configured := m.Config()
	if !configured {
		log.Error().Str("to", to).Msg("SMTP not configured, skipping send")
		return nil
	}
	addr := net.JoinHostPort(server, strconv.Itoa(port))
	msg := buildMessage(to, from, subject, "text/plain", body)
	err := m.sendClient(server, addr, from, to, msg, username, password, port, useTLS)
	if err != nil {
		return err
	}
	log.Info().Str("to", to).Str("subject", subject).Msg("Email sent successfully")
	return nil
}

// buildMessage builds the raw email bytes (headers + body).
func buildMessage(to, from, subject, contentType, body string) []byte {
	return []byte(
		"To: " + to + "\r\n" +
			"From: " + from + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Content-Type: " + contentType + "; charset=UTF-8\r\n" +
			"MIME-Version: 1.0\r\n" +
			"\r\n" +
			body + "\r\n",
	)
}

// SendHTML sends an HTML email. No-op if SMTP is not configured.
func (m *Mailer) SendHTML(to, subject, htmlBody string) error {
	server, from, username, password, port, useTLS, configured := m.Config()
	if !configured {
		log.Debug().Str("to", to).Msg("SMTP not configured, skipping send")
		return nil
	}
	if from == "" {
		log.Warn().Msg("SMTP from address not set, skipping send")
		return nil
	}
	addr := net.JoinHostPort(server, strconv.Itoa(port))
	msg := buildMessage(to, from, subject, "text/html", htmlBody)
	err := m.sendClient(server, addr, from, to, msg, username, password, port, useTLS)
	if err != nil {
		return err
	}
	log.Info().Str("to", to).Str("subject", subject).Msg("Email sent successfully")
	return nil
}

// sendClient sends the email using the SMTP client.
func (m *Mailer) sendClient(server, addr, from, to string, msg []byte, username, password string, port int, useTLS bool) error {
	var client *smtp.Client
	if port == 465 {
		tlsConfig := &tls.Config{ServerName: server}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			log.Error().Err(err).Str("addr", addr).Msg("SMTP TLS dial failed")
			return fmt.Errorf("smtp dial: %w", err)
		}
		defer conn.Close()
		client, err = smtp.NewClient(conn, server)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()
	} else {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Error().Err(err).Str("addr", addr).Msg("SMTP dial failed")
			return fmt.Errorf("smtp dial: %w", err)
		}
		defer conn.Close()
		client, err = smtp.NewClient(conn, server)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()
		if useTLS {
			tlsConfig := &tls.Config{ServerName: server}
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("smtp starttls: %w", err)
			}
		}
	}
	if username != "" && password != "" {
		auth := smtp.PlainAuth("", username, password, server)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	if err := client.Quit(); err != nil {
		log.Debug().Err(err).Msg("SMTP quit error")
	}
	return nil
}

// inviteEmailHTML returns a simple styled HTML body for the invite email.
func inviteEmailHTML(platformName, inviteLink string) string {
	platformName = html.EscapeString(platformName)
	inviteLinkEsc := html.EscapeString(inviteLink)
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin:0; padding:0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #F7F6F3;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color: #F7F6F3; padding: 24px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width: 560px; background-color: #ffffff; border: 1px solid #E3E2E0; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06);">
          <tr>
            <td style="padding: 32px 32px 24px 32px;">
              <h1 style="margin: 0 0 8px 0; font-size: 22px; font-weight: 600; color: #37352F;">You're invited to %s</h1>
              <p style="margin: 0; font-size: 15px; line-height: 1.5; color: #6B6966;">You have been invited to join <strong style="color: #37352F;">%s</strong>. Click the button below to create your account. This link expires in 7 days.</p>
            </td>
          </tr>
          <tr>
            <td style="padding: 8px 32px 24px 32px;">
              <a href="%s" style="display: inline-block; padding: 12px 24px; font-size: 15px; font-weight: 500; color: #ffffff; background-color: #37352F; text-decoration: none; border-radius: 6px;">Accept invite</a>
            </td>
          </tr>
          <tr>
            <td style="padding: 0 32px 32px 32px; font-size: 13px; line-height: 1.5; color: #6B6966;">
              <p style="margin: 0;">If the button doesn't work, copy and paste this link into your browser:</p>
              <p style="margin: 8px 0 0 0; word-break: break-all;"><a href="%s" style="color: #37352F;">%s</a></p>
              <p style="margin: 24px 0 0 0;">If you didn't expect this email, you can ignore it.</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, platformName, platformName, inviteLinkEsc, inviteLinkEsc, inviteLinkEsc)
}

// SendInviteEmail sends the user invite email with the sign-up link (HTML).
func (m *Mailer) SendInviteEmail(to, inviteLink, platformName string) error {
	subject := "You're invited to " + platformName
	htmlBody := inviteEmailHTML(platformName, inviteLink)
	return m.SendHTML(to, subject, htmlBody)
}

// resetPasswordEmailHTML returns the HTML body for the password reset email.
func resetPasswordEmailHTML(platformName, resetLink string) string {
	platformName = html.EscapeString(platformName)
	resetLinkEsc := html.EscapeString(resetLink)
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin:0; padding:0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #F7F6F3;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color: #F7F6F3; padding: 24px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width: 560px; background-color: #ffffff; border: 1px solid #E3E2E0; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06);">
          <tr>
            <td style="padding: 32px 32px 24px 32px;">
              <h1 style="margin: 0 0 8px 0; font-size: 22px; font-weight: 600; color: #37352F;">Reset your password</h1>
              <p style="margin: 0; font-size: 15px; line-height: 1.5; color: #6B6966;">You requested a password reset for <strong style="color: #37352F;">%s</strong>. Click the button below to set a new password. This link expires in 24 hours.</p>
            </td>
          </tr>
          <tr>
            <td style="padding: 8px 32px 24px 32px;">
              <a href="%s" style="display: inline-block; padding: 12px 24px; font-size: 15px; font-weight: 500; color: #ffffff; background-color: #37352F; text-decoration: none; border-radius: 6px;">Reset password</a>
            </td>
          </tr>
          <tr>
            <td style="padding: 0 32px 32px 32px; font-size: 13px; line-height: 1.5; color: #6B6966;">
              <p style="margin: 0;">If the button doesn't work, copy and paste this link into your browser:</p>
              <p style="margin: 8px 0 0 0; word-break: break-all;"><a href="%s" style="color: #37352F;">%s</a></p>
              <p style="margin: 24px 0 0 0;">If you didn't request this, you can ignore this email. Your password will not be changed.</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, platformName, resetLinkEsc, resetLinkEsc, resetLinkEsc)
}

// SendPasswordResetEmail sends the password reset email with the reset link (HTML).
func (m *Mailer) SendPasswordResetEmail(to, resetLink, platformName string) error {
	subject := "Reset your password - " + platformName
	htmlBody := resetPasswordEmailHTML(platformName, resetLink)
	return m.SendHTML(to, subject, htmlBody)
}
