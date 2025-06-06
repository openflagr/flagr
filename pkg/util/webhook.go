package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// GenerateWebhookSignature generates an HMAC-SHA256 signature for webhook payloads
func GenerateWebhookSignature(payload interface{}, secret string) string {
	// Convert payload to JSON string
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return ""
	}

	// Create HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payloadBytes)

	// Return hex-encoded signature
	return hex.EncodeToString(h.Sum(nil))
}
