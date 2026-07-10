package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

const sessionLabel = "mindex-admin-session"

func GenerateToken(adminPassword string) string {
	mac := hmac.New(sha256.New, []byte(adminPassword))
	_, _ = mac.Write([]byte(sessionLabel))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyPassword(password, adminPassword string) bool {
	if adminPassword == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(adminPassword)) == 1
}

func VerifyToken(token, adminPassword string) bool {
	if adminPassword == "" || token == "" {
		return false
	}

	expected := GenerateToken(adminPassword)
	if len(token) != len(expected) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
