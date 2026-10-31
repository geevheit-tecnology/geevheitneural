package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"
)

const (
	jwksCacheTTL        = 10 * time.Minute
	jwksMinRefreshDelay = 30 * time.Second
	jwksMaxBodyBytes    = 1 << 20
)

// jwksProvider busca e guarda em cache as chaves públicas publicadas pelo Supabase Auth.
type jwksProvider struct {
	url    string
	client *http.Client

	mu          sync.Mutex
	keys        map[string]any
	fetchedAt   time.Time
	lastAttempt time.Time
}

func newJWKSProvider(url string) *jwksProvider {
	return &jwksProvider{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Key devolve a chave pública do kid informado. Atualiza o cache quando expirado
// ou quando o kid é desconhecido (rotação de chaves), respeitando um intervalo mínimo.
func (p *jwksProvider) Key(ctx context.Context, kid string) (any, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	expired := time.Since(p.fetchedAt) > jwksCacheTTL
	if key, ok := p.keys[kid]; ok && !expired {
		return key, nil
	}

	if time.Since(p.lastAttempt) >= jwksMinRefreshDelay || p.keys == nil {
		p.lastAttempt = time.Now()
		keys, err := p.fetch(ctx)
		if err != nil {
			if key, ok := p.keys[kid]; ok {
				return key, nil // mantém a chave em cache se a atualização falhar
			}
			return nil, err
		}
		p.keys = keys
		p.fetchedAt = time.Now()
	}

	key, ok := p.keys[kid]
	if !ok {
		return nil, fmt.Errorf("chave JWKS não encontrada para kid %q", kid)
	}
	return key, nil
}

type jwksDocument struct {
	Keys []struct {
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		Crv string `json:"crv"`
		X   string `json:"x"`
		Y   string `json:"y"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func (p *jwksProvider) fetch(ctx context.Context) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar JWKS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS respondeu status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, jwksMaxBodyBytes))
	if err != nil {
		return nil, err
	}
	var doc jwksDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("JWKS inválido: %w", err)
	}

	keys := make(map[string]any, len(doc.Keys))
	for _, k := range doc.Keys {
		switch k.Kty {
		case "EC":
			if k.Crv != "P-256" {
				continue
			}
			x, errX := decodeBigInt(k.X)
			y, errY := decodeBigInt(k.Y)
			if errX != nil || errY != nil {
				continue
			}
			pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
			if !pub.Curve.IsOnCurve(x, y) {
				continue
			}
			keys[k.Kid] = pub
		case "RSA":
			n, errN := decodeBigInt(k.N)
			e, errE := decodeBigInt(k.E)
			if errN != nil || errE != nil || !e.IsInt64() {
				continue
			}
			keys[k.Kid] = &rsa.PublicKey{N: n, E: int(e.Int64())}
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("JWKS sem chaves utilizáveis")
	}
	return keys, nil
}

func decodeBigInt(s string) (*big.Int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(raw), nil
}
