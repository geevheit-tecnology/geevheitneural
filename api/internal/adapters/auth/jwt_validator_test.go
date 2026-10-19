package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/auth"
	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret      = "test-secret-at-least-32-bytes-long-12345"
	testOwnerUserID = "a0000000-0000-0000-0000-000000000001"
)

func generateTestToken(sub string, aal string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub": sub,
		"aal": aal,
		"exp": exp.Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))
	return tokenString
}

func TestTokenValidator(t *testing.T) {
	validator := auth.NewTokenValidator(testOwnerUserID, testSecret)

	t.Run("Valid token with owner sub and aal2 succeeds", func(t *testing.T) {
		tok := generateTestToken(testOwnerUserID, "aal2", time.Now().Add(time.Hour))
		res, err := validator.ValidateToken("Bearer " + tok)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Owner.ID != testOwnerUserID {
			t.Errorf("expected owner ID %s, got %s", testOwnerUserID, res.Owner.ID)
		}
		if res.Owner.AAL != "aal2" {
			t.Errorf("expected AAL aal2, got %s", res.Owner.AAL)
		}
	})

	t.Run("Expired token fails", func(t *testing.T) {
		tok := generateTestToken(testOwnerUserID, "aal2", time.Now().Add(-time.Hour))
		_, err := validator.ValidateToken(tok)
		if !errors.Is(err, auth.ErrTokenExpired) {
			t.Errorf("expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("Different sub fails with ErrNotOwner", func(t *testing.T) {
		otherUser := "b0000000-0000-0000-0000-000000000099"
		tok := generateTestToken(otherUser, "aal2", time.Now().Add(time.Hour))
		_, err := validator.ValidateToken(tok)
		if !errors.Is(err, auth.ErrNotOwner) {
			t.Errorf("expected ErrNotOwner, got %v", err)
		}
	})

	t.Run("aal1 fails with ErrMFAEnforced", func(t *testing.T) {
		tok := generateTestToken(testOwnerUserID, "aal1", time.Now().Add(time.Hour))
		_, err := validator.ValidateToken(tok)
		if !errors.Is(err, auth.ErrMFAEnforced) {
			t.Errorf("expected ErrMFAEnforced, got %v", err)
		}
	})

	t.Run("Empty token fails with ErrMissingAuth", func(t *testing.T) {
		_, err := validator.ValidateToken("")
		if !errors.Is(err, auth.ErrMissingAuth) {
			t.Errorf("expected ErrMissingAuth, got %v", err)
		}
	})
}
