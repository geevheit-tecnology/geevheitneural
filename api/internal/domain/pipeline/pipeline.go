package pipeline

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPipelineAlreadyRunning = errors.New("pipeline já está em execução para este cartão")
	ErrSpendLimitExceeded     = errors.New("limite de gasto do projeto atingido: execução interrompida")
	ErrInvalidPipelineStage   = errors.New("estágio do pipeline inválido")
)

type Role string

const (
	RoleRequisitos Role = "requisitos"
	RoleArquiteto  Role = "arquiteto"
	RoleDev        Role = "dev"
	RoleQA         Role = "qa"
	RoleSeguranca  Role = "seguranca"
	RolePublicador Role = "publicador"
	RoleSuporte    Role = "suporte"
)

type Status string

const (
	StatusIdle             Status = "idle"
	StatusRunning          Status = "running"
	StatusWaitingApproval  Status = "waiting_approval"
	StatusPausedSpendLimit Status = "paused_spend_limit"
	StatusCompleted        Status = "completed"
	StatusFailed           Status = "failed"
)

type RunnerResult struct {
	Role         Role           `json:"role"`
	Output       string         `json:"output"`
	CostCents    int64          `json:"cost_cents"`
	Artifacts    map[string]any `json:"artifacts,omitempty"`
	NextRole     *Role          `json:"next_role,omitempty"`
	RequiresGate bool           `json:"requires_gate"`
}

// AgentRunner abstrai o provedor de IA (Claude SDK, OpenAI, Local, Mock)
type AgentRunner interface {
	RunStep(ctx context.Context, role Role, prompt string, contextMeta map[string]any) (*RunnerResult, error)
}

type CardPipeline struct {
	ID              string
	OwnerID         string
	CardID          string
	Stage           Role
	ActiveRole      Role
	SpendTotalCents int64
	SpendLimitCents int64
	Status          Status
	CurrentTask     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewCardPipeline(id, ownerID, cardID string, spendLimitCents int64, now time.Time) *CardPipeline {
	if spendLimitCents <= 0 {
		spendLimitCents = 10000 // Padrão R$ 100,00 de limite de IA por projeto
	}
	return &CardPipeline{
		ID:              id,
		OwnerID:         ownerID,
		CardID:          cardID,
		Stage:           RoleRequisitos,
		ActiveRole:      RoleRequisitos,
		SpendTotalCents: 0,
		SpendLimitCents: spendLimitCents,
		Status:          StatusIdle,
		CurrentTask:     "Aguardando início",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (p *CardPipeline) AddSpend(cents int64) error {
	if p.SpendLimitCents > 0 && p.SpendTotalCents+cents > p.SpendLimitCents {
		return ErrSpendLimitExceeded
	}
	p.SpendTotalCents += cents
	return nil
}
