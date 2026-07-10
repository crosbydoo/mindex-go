package auth

import "testing"

func TestGenerateToken_Deterministic(t *testing.T) {
	password := "test-admin-password"
	token1 := GenerateToken(password)
	token2 := GenerateToken(password)

	if token1 == "" {
		t.Fatal("expected non-empty token")
	}
	if token1 != token2 {
		t.Fatalf("expected deterministic token, got %q and %q", token1, token2)
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "secret-password"

	if !VerifyPassword(password, password) {
		t.Fatal("expected password verification to succeed")
	}
	if VerifyPassword("wrong", password) {
		t.Fatal("expected password verification to fail for wrong password")
	}
	if VerifyPassword(password, "") {
		t.Fatal("expected password verification to fail when admin password is empty")
	}
}

func TestVerifyToken(t *testing.T) {
	password := "admin-secret"
	token := GenerateToken(password)

	if !VerifyToken(token, password) {
		t.Fatal("expected token verification to succeed")
	}
	if VerifyToken("invalid-token", password) {
		t.Fatal("expected token verification to fail for invalid token")
	}
	if VerifyToken(token, "other-password") {
		t.Fatal("expected token verification to fail for different admin password")
	}
	if VerifyToken("", password) {
		t.Fatal("expected token verification to fail for empty token")
	}
}

func TestVerifyToken_LengthMismatch(t *testing.T) {
	password := "admin-secret"
	token := GenerateToken(password)

	if VerifyToken(token+"extra", password) {
		t.Fatal("expected token verification to fail for length mismatch")
	}
}
