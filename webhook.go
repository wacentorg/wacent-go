package wacent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

// VerifyWebhookSignature checks the X-WACENT-Signature header of a delivery.
// Wacent signs the raw body with HMAC-SHA256 using hex(sha256(secret)) as the key:
//
//	header = "sha256=" + hex(HMAC_SHA256(key = hex(SHA256(secret)), rawBody))
//
// Pass the body exactly as received.
func VerifyWebhookSignature(rawBody []byte, signatureHeader, secret string) bool {
	if signatureHeader == "" {
		return false
	}
	sum := sha256.Sum256([]byte(secret))
	mac := hmac.New(sha256.New, []byte(hex.EncodeToString(sum[:])))
	mac.Write(rawBody)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.TrimSpace(signatureHeader)), []byte(expected))
}

// ParseWebhook verifies the signature and decodes the event.
func ParseWebhook(rawBody []byte, signatureHeader, secret string) (*WebhookEvent, error) {
	if !VerifyWebhookSignature(rawBody, signatureHeader, secret) {
		return nil, errors.New("wacent: invalid webhook signature")
	}
	var ev WebhookEvent
	if err := json.Unmarshal(rawBody, &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
