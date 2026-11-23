package elevenlabs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

var residencySuffix = regexp.MustCompile(`_residency_[a-z0-9]+$`)

// Official Speech Engine signing: HS256 using SHA256(base API key), with
// provider issuer/subject and a 60-second clock allowance. No token is logged.
func verifyJWT(value, key string, now time.Time) bool {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		value = strings.TrimSpace(value[7:])
	}
	if len(value) > 12000 {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return false
	}
	headerRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if json.Unmarshal(headerRaw, &header) != nil || header.Algorithm != "HS256" {
		return false
	}
	baseKey := residencySuffix.ReplaceAllString(strings.TrimSpace(key), "")
	secret := sha256.Sum256([]byte(baseKey))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(mac.Sum(nil), signature) {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims struct {
		Issuer  string `json:"iss"`
		Subject string `json:"sub"`
		Expires int64  `json:"exp"`
		Issued  int64  `json:"iat"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return false
	}
	return claims.Issuer == "https://api.elevenlabs.io/convai/speech-engine" && claims.Subject == "convai_speech_engine_upstream" && claims.Expires != 0 && claims.Issued != 0 && claims.Expires+60 >= now.Unix() && claims.Issued-60 <= now.Unix()
}
