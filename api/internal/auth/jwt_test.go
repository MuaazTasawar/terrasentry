package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-do-not-use-in-prod"

func TestGenerateAndValidateToken_RoundTrip(t *testing.T) {
	token, err := GenerateToken(testSecret, 24, "user-123", "on-call@example.com")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}

	claims, err := ValidateToken(testSecret, token)
	if err != nil {
		t.Fatalf("ValidateToken returned error on a freshly issued token: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got %q", claims.UserID)
	}
	if claims.Email != "on-call@example.com" {
		t.Errorf("expected Email 'on-call@example.com', got %q", claims.Email)
	}
}

func TestValidateToken_WrongSecretRejected(t *testing.T) {
	token, err := GenerateToken(testSecret, 24, "user-123", "on-call@example.com")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	_, err = ValidateToken("a-completely-different-secret", token)
	if err == nil {
		t.Fatal("expected ValidateToken to reject a token signed with a different secret, got nil error")
	}
}

func TestValidateToken_ExpiredTokenRejected(t *testing.T) {
	// Build an already-expired token by hand rather than waiting on a timer.
	now := time.Now()
	claims := Claims{
		UserID: "user-456",
		Email:  "expired@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // expired 1h ago
			Subject:   "user-456",
			Issuer:    "terrasentry-api",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to build expired test token: %v", err)
	}

	_, err = ValidateToken(testSecret, signed)
	if err == nil {
		t.Fatal("expected ValidateToken to reject an expired token, got nil error")
	}
}

func TestValidateToken_MalformedTokenRejected(t *testing.T) {
	_, err := ValidateToken(testSecret, "not-a-real-jwt")
	if err == nil {
		t.Fatal("expected ValidateToken to reject a malformed token string, got nil error")
	}
}

func TestValidateToken_WrongSigningMethodRejected(t *testing.T) {
	// Regression guard: a token signed with 'none' (alg confusion attack)
	// must never validate, even if ValidateToken is later refactored.
	now := time.Now()
	claims := Claims{
		UserID: "attacker",
		Email:  "attacker@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			Subject:   "attacker",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to build 'none'-alg test token: %v", err)
	}

	_, err = ValidateToken(testSecret, signed)
	if err == nil {
		t.Fatal("expected ValidateToken to reject a 'none'-algorithm token, got nil error")
	}
}
