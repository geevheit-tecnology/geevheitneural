package portal

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTokenExpired       = errors.New("token de acesso expirado")
	ErrTokenRevoked       = errors.New("token de acesso revogado")
	ErrAlreadyAccepted    = errors.New("a entrega deste projeto já foi formalmente aceita")
	ErrDeliveryNotReady   = errors.New("o projeto ainda não está na etapa de entrega para ser aceito")
	ErrMissingAcceptedBy  = errors.New("identificação do responsável pelo aceite é obrigatória")
)

type PortalToken struct {
	ID                string     `json:"id"`
	OwnerID           string     `json:"owner_id"`
	CardID            string     `json:"card_id"`
	ClientID          string     `json:"client_id"`
	TokenHash         string     `json:"token_hash"`
	ExpiresAt         time.Time  `json:"expires_at"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	LastAccessedAt    *time.Time `json:"last_accessed_at,omitempty"`
	AcceptedAt        *time.Time `json:"accepted_at,omitempty"`
	AcceptedBy        string     `json:"accepted_by,omitempty"`
	AcceptedIP        string     `json:"accepted_ip,omitempty"`
	AcceptedUserAgent string     `json:"accepted_user_agent,omitempty"`
	AcceptedStateHash string     `json:"accepted_state_hash,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func NewPortalToken(cardID, clientID, ownerID, tokenHash string, expiresAt time.Time) *PortalToken {
	now := time.Now()
	return &PortalToken{
		ID:        uuid.NewString(),
		OwnerID:   ownerID,
		CardID:    cardID,
		ClientID:  clientID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *PortalToken) IsActive(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return now.Before(t.ExpiresAt)
}

func (t *PortalToken) Revoke(now time.Time) {
	t.RevokedAt = &now
	t.UpdatedAt = now
}

func (t *PortalToken) Touch(now time.Time) {
	t.LastAccessedAt = &now
	t.UpdatedAt = now
}

func (t *PortalToken) Accept(by, ip, ua string, cardState string, now time.Time) error {
	if t.AcceptedAt != nil {
		return ErrAlreadyAccepted
	}
	if by == "" {
		return ErrMissingAcceptedBy
	}

	stateHashBytes := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", t.CardID, cardState, now.Format(time.RFC3339))))
	stateHash := hex.EncodeToString(stateHashBytes[:])

	t.AcceptedAt = &now
	t.AcceptedBy = by
	t.AcceptedIP = ip
	t.AcceptedUserAgent = ua
	t.AcceptedStateHash = stateHash
	t.UpdatedAt = now
	return nil
}

// PortalStageView representa uma das 4 etapas visíveis pelo cliente
type PortalStageView struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	Status      string     `json:"status"` // "concluido", "em_andamento", "pendente"
	Order       int        `json:"order"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// ClientPortalDTO representa o DTO estritamente isolado para o cliente
// NUNCA expõe: custos, margem, lucros, agentes, notas internas ou dados de outros clientes
type ClientPortalDTO struct {
	ProjectTitle       string            `json:"project_title"`
	ClientName         string            `json:"client_name"`
	AppName            string            `json:"app_name"`
	CurrentStageLabel  string            `json:"current_stage_label"`
	ProgressPercentage int               `json:"progress_percentage"`
	Stages             []PortalStageView `json:"stages"`
	CanAccept          bool              `json:"can_accept"`
	IsAccepted         bool              `json:"is_accepted"`
	AcceptedAt         *time.Time        `json:"accepted_at,omitempty"`
	AcceptedBy         string            `json:"accepted_by,omitempty"`
}
