package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
)

type EmailService struct {
	cfg    *config.Config
	client *http.Client
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

func (s *EmailService) SendVerificationEmail(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.FrontendURL, url.QueryEscape(token))
	subject := "Verify your ChaosPlane email"
	body := fmt.Sprintf(`<h2>Welcome to ChaosPlane</h2>
<p>Click the link below to verify your email address:</p>
<p><a href="%s">Verify Email</a></p>
<p>This link expires in 24 hours.</p>
<p>If you didn't create an account, you can safely ignore this email.</p>`, link)
	return s.send(ctx, to, subject, body)
}

func (s *EmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.FrontendURL, url.QueryEscape(token))
	subject := "Reset your ChaosPlane password"
	body := fmt.Sprintf(`<h2>Password Reset</h2>
<p>Click the link below to reset your password:</p>
<p><a href="%s">Reset Password</a></p>
<p>This link expires in 24 hours. If you didn't request this, you can safely ignore this email.</p>`, link)
	return s.send(ctx, to, subject, body)
}

func (s *EmailService) SendInvitationEmail(ctx context.Context, to, inviterName, orgName, token string) error {
	link := fmt.Sprintf("%s/invitations/accept?token=%s", s.cfg.FrontendURL, url.QueryEscape(token))
	subject := fmt.Sprintf("You've been invited to %s on ChaosPlane", orgName)
	body := fmt.Sprintf(`<h2>You're Invited</h2>
<p>%s has invited you to join <strong>%s</strong> on ChaosPlane.</p>
<p><a href="%s">Accept Invitation</a></p>
<p>This invitation expires in 7 days.</p>`, inviterName, orgName, link)
	return s.send(ctx, to, subject, body)
}

func (s *EmailService) SendEmailChangeNotification(ctx context.Context, oldEmail, newEmail, token string) error {
	link := fmt.Sprintf("%s/confirm-email-change?token=%s", s.cfg.FrontendURL, url.QueryEscape(token))
	subject := "Confirm your new email address"
	body := fmt.Sprintf(`<h2>Email Change Request</h2>
<p>A request was made to change your ChaosPlane email from %s to %s.</p>
<p><a href="%s">Confirm New Email</a></p>
<p>If you didn't request this change, please secure your account immediately.</p>`, oldEmail, newEmail, link)

	if err := s.send(ctx, newEmail, subject, body); err != nil {
		return err
	}
	notifySubject := "Your ChaosPlane email is being changed"
	notifyBody := fmt.Sprintf(`<h2>Email Change Notice</h2>
<p>A request was made to change your ChaosPlane email to %s.</p>
<p>If you didn't request this, please contact support immediately.</p>`, newEmail)
	return s.send(ctx, oldEmail, notifySubject, notifyBody)
}

func (s *EmailService) send(ctx context.Context, to, subject, htmlBody string) error {
	region := s.cfg.SESRegion
	if region == "" {
		region = "us-east-1"
	}
	accessKey := s.cfg.SESAccessKey
	secretKey := s.cfg.SESSecretKey
	sender := s.cfg.SESFromEmail

	if accessKey == "" || secretKey == "" || sender == "" {
		return fmt.Errorf("SES not configured (missing SES_ACCESS_KEY, SES_SECRET_KEY, or SES_FROM_EMAIL)")
	}

	endpoint := fmt.Sprintf("https://email.%s.amazonaws.com/", region)
	params := url.Values{
		"Action":                           {"SendEmail"},
		"Source":                           {sender},
		"Destination.ToAddresses.member.1": {to},
		"Message.Subject.Data":             {subject},
		"Message.Subject.Charset":          {"UTF-8"},
		"Message.Body.Html.Data":           {htmlBody},
		"Message.Body.Html.Charset":        {"UTF-8"},
	}

	body := []byte(params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create SES request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	signSESRequest(req, body, accessKey, secretKey, region)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("SES request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			XMLName xml.Name `xml:"ErrorResponse"`
			Error   struct {
				Message string `xml:"Message"`
			} `xml:"Error"`
		}
		if xmlErr := xml.NewDecoder(resp.Body).Decode(&errResp); xmlErr == nil {
			return fmt.Errorf("SES error: %s", errResp.Error.Message)
		}
		return fmt.Errorf("SES error: status %d", resp.StatusCode)
	}
	return nil
}

func signSESRequest(req *http.Request, payload []byte, accessKey, secretKey, region string) {
	now := time.Now().UTC()
	datestamp := now.Format("20060102")
	amzdate := now.Format("20060102T150405Z")

	req.Header.Set("X-Amz-Date", amzdate)

	payloadHash := sesSHA256Hex(payload)
	host := req.URL.Host

	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-amz-date:%s\n",
		req.Header.Get("Content-Type"), host, amzdate)
	signedHeaders := "content-type;host;x-amz-date"

	canonicalRequest := strings.Join([]string{
		"POST", "/", "", canonicalHeaders, signedHeaders, payloadHash,
	}, "\n")

	credentialScope := fmt.Sprintf("%s/%s/ses/aws4_request", datestamp, region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzdate, credentialScope, sesSHA256Hex([]byte(canonicalRequest)))

	kDate := sesHMAC([]byte("AWS4"+secretKey), []byte(datestamp))
	kRegion := sesHMAC(kDate, []byte(region))
	kService := sesHMAC(kRegion, []byte("ses"))
	kSigning := sesHMAC(kService, []byte("aws4_request"))

	signature := fmt.Sprintf("%x", sesHMAC(kSigning, []byte(stringToSign)))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature))

	_ = base64.StdEncoding
	_ = sort.Strings
}

func sesSHA256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:])
}

func sesHMAC(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
