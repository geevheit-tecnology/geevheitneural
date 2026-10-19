package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken   = errors.New("token JWT inválido ou malformado")
	ErrTokenExpired   = errors.New("token JWT expirado")
	ErrNotOwner       = errors.New("acesso negado: usuário não é o dono do sistema")
	ErrMFAEnforced    = errors.New("acesso negado: segundo fator 2FA (AAL2) é obrigatório")
	ErrMissingAuth    = errors.New("cabeçalho de autorização ausente ou inválido")
)

type TokenValidator struct {
	ownerUserID string
	jwtSecret   []byte
}

func NewTokenValidator(ownerUserID string, jwtSecret string) *TokenValidator {
	return &TokenValidator{
		ownerUserID: ownerUserID,
		jwtSecret:   []byte(jwtSecret),
	}
}

type AuthResult struct {
	Owner      system.OwnerPrincipal
	ClaimsJSON string
}

// ValidateToken decodifica e valida o JWT, garantindo sub == OWNER_USER_ID e aal == 'aal2'.
func (v *TokenValidator) ValidateToken(tokenStr string) (AuthResult, error) {
	if tokenStr == "" {
		return AuthResult{}, ErrMissingAuth
	}

	// Remove prefixo Bearer se presente
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		// Se segredo HS256 foi configurado, valida por ele
		if len(v.jwtSecret) > 0 {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", t.Header["alg"])
			}
			return v.jwtSecret, nil
		}
		// Para modo sem segredo configurado (ex: validação com JWKS ou skip de assinatura em teste)
		return nil, errors.New("chave de validação de assinatura não configurada")
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return AuthResult{}, ErrTokenExpired
		}
		return AuthResult{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return AuthResult{}, ErrInvalidToken
	}

	// 1. Valida expiração
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return AuthResult{}, ErrTokenExpired
		}
	}

	// 2. Valida se o sujeito (sub) é o único dono do sistema
	sub, _ := claims["sub"].(string)
	if sub == "" || sub != v.ownerUserID {
		return AuthResult{}, ErrNotOwner
	}

	// 3. Valida se o nível de garantia de autenticação é AAL2 (2FA obrigatório)
	aal, _ := claims["aal"].(string)
	if aal != "aal2" {
		return AuthResult{}, ErrMFAEnforced
	}

	ownerPrincipal, err := system.NewOwnerPrincipal(sub, aal)
	if err != nil {
		return AuthResult{}, err
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		claimsBytes = []byte("{}")
	}

	return AuthResult{
		Owner:      ownerPrincipal,
		ClaimsJSON: string(claimsBytes),
	}, nil
}
