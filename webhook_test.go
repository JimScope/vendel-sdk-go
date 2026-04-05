package vendel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func computeHMAC(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	payload := `{"event":"sms.sent","message_id":"m1"}`
	sig := computeHMAC(payload, "webhook_secret_123")
	if !VerifyWebhookSignature(payload, sig, "webhook_secret_123") {
		t.Error("expected valid signature")
	}
}

func TestVerifyWebhookSignature_Invalid(t *testing.T) {
	if VerifyWebhookSignature("payload", "not_a_valid_hex_signature_at_all_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "secret") {
		t.Error("expected invalid signature")
	}
}

func TestVerifyWebhookSignature_WrongSecret(t *testing.T) {
	payload := `{"event":"test"}`
	sig := computeHMAC(payload, "correct_secret")
	if VerifyWebhookSignature(payload, sig, "wrong_secret") {
		t.Error("expected rejection with wrong secret")
	}
}

func TestVerifyWebhookSignature_EmptyPayload(t *testing.T) {
	sig := computeHMAC("", "secret")
	if !VerifyWebhookSignature("", sig, "secret") {
		t.Error("empty payload should still verify")
	}
}
