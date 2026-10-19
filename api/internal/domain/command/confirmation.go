package command

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	ErrTokenNotFound    = errors.New("token_not_found")
	ErrTokenExpired     = errors.New("token_expired")
	ErrTokenAlreadyUsed = errors.New("token_already_used")
)

type ActionType string

const (
	ActionTypeRejectCard  ActionType = "reject_card"
	ActionTypePauseSystem ActionType = "pause_system"
)

// PendingAction representa a ação destrutiva guardada aguardando confirmação explícita.
type PendingAction struct {
	Token     string
	Type      ActionType
	Prompt    string // texto exibido na confirmação
	TargetID  string
	Payload   map[string]any
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
}

// ConfirmationService gerencia tokens criptográficos de uso único para ações destrutivas.
type ConfirmationService struct {
	mu      sync.RWMutex
	pending map[string]*PendingAction
	ttl     time.Duration
}

func NewConfirmationService(ttl time.Duration) *ConfirmationService {
	return &ConfirmationService{
		pending: make(map[string]*PendingAction),
		ttl:     ttl,
	}
}

// IssueToken gera um token aleatório seguro de 32 bytes em hexadecimal (64 chars) com expiração.
func (s *ConfirmationService) IssueToken(actType ActionType, targetID string, prompt string, payload map[string]any) (*PendingAction, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(bytes)

	now := time.Now()
	act := &PendingAction{
		Token:     token,
		Type:      actType,
		Prompt:    prompt,
		TargetID:  targetID,
		Payload:   payload,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
		Used:      false,
	}

	s.mu.Lock()
	s.pending[token] = act
	s.mu.Unlock()

	return act, nil
}

// Consume valida o token e o consome de forma atômica e irreversível (uso estritamente único).
func (s *ConfirmationService) Consume(token string) (*PendingAction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	act, ok := s.pending[token]
	if !ok {
		return nil, ErrTokenNotFound
	}

	if act.Used {
		return nil, ErrTokenAlreadyUsed
	}

	if time.Now().After(act.ExpiresAt) {
		delete(s.pending, token)
		return nil, ErrTokenExpired
	}

	// Marca como consumido e remove
	act.Used = true
	delete(s.pending, token)

	return act, nil
}
