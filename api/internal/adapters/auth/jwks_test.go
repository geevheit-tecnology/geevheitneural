package auth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/auth"
	"github.com/golang-jwt/jwt/v5"
)

func pad32(b []byte) []byte {
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

func jwksServer(t *testing.T, kid string, pub *ecdsa.PublicKey) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{{
				"kty": "EC", "crv": "P-256", "alg": "ES256", "kid": kid,
				"x": base64.RawURLEncoding.EncodeToString(pad32(pub.X.Bytes())),
				"y": base64.RawURLEncoding.EncodeToString(pad32(pub.Y.Bytes())),
			}},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func signES256(t *testing.T, key *ecdsa.PrivateKey, kid, sub, aal string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": sub, "aal": aal, "exp": time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestTokenValidatorES256JWKS(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	other, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	srv := jwksServer(t, "kid-1", &key.PublicKey)
	validator := auth.NewTokenValidator(testOwnerUserID, testSecret).WithJWKS(srv.URL)

	t.Run("ES256 valid owner aal2 succeeds", func(t *testing.T) {
		res, err := validator.ValidateToken("Bearer " + signES256(t, key, "kid-1", testOwnerUserID, "aal2"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Owner.ID != testOwnerUserID {
			t.Errorf("unexpected owner %s", res.Owner.ID)
		}
	})
	t.Run("ES256 signed by another key fails", func(t *testing.T) {
		_, err := validator.ValidateToken(signES256(t, other, "kid-1", testOwnerUserID, "aal2"))
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
	t.Run("ES256 unknown kid fails", func(t *testing.T) {
		_, err := validator.ValidateToken(signES256(t, key, "kid-unknown", testOwnerUserID, "aal2"))
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
	t.Run("ES256 other user fails with ErrNotOwner", func(t *testing.T) {
		_, err := validator.ValidateToken(signES256(t, key, "kid-1", "b0000000-0000-0000-0000-000000000099", "aal2"))
		if !errors.Is(err, auth.ErrNotOwner) {
			t.Errorf("expected ErrNotOwner, got %v", err)
		}
	})
	t.Run("ES256 aal1 fails with ErrMFAEnforced", func(t *testing.T) {
		_, err := validator.ValidateToken(signES256(t, key, "kid-1", testOwnerUserID, "aal1"))
		if !errors.Is(err, auth.ErrMFAEnforced) {
			t.Errorf("expected ErrMFAEnforced, got %v", err)
		}
	})
	t.Run("alg none is rejected", func(t *testing.T) {
		tok := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": testOwnerUserID, "aal": "aal2", "exp": time.Now().Add(time.Hour).Unix()})
		s, _ := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if _, err := validator.ValidateToken(s); !errors.Is(err, auth.ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
	t.Run("ES256 without JWKS configured fails", func(t *testing.T) {
		plain := auth.NewTokenValidator(testOwnerUserID, testSecret)
		if _, err := plain.ValidateToken(signES256(t, key, "kid-1", testOwnerUserID, "aal2")); !errors.Is(err, auth.ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}
