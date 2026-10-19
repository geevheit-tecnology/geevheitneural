package system

import (
	"errors"
	"time"
)

var (
	ErrSystemPaused  = errors.New("sistema pausado: nenhuma execução de agente ou conector é liberada")
	ErrActorInactive = errors.New("ator inativo: conectores e agentes desligados não podem agir")
)

// SystemState representa o estado global de pausa do sistema (Botão Parar).
type SystemState struct {
	Paused    bool
	PausedBy  *string
	PausedAt  *time.Time
	Reason    string
	UpdatedAt time.Time
}

// Pause altera o estado do sistema para pausado com registro de auditoria.
func (s *SystemState) Pause(owner OwnerPrincipal, reason string, now time.Time) {
	s.Paused = true
	s.PausedBy = &owner.ID
	s.PausedAt = &now
	s.Reason = reason
	s.UpdatedAt = now
}

// Resume retoma o estado normal do sistema.
func (s *SystemState) Resume(owner OwnerPrincipal, now time.Time) {
	s.Paused = false
	s.PausedBy = nil
	s.PausedAt = nil
	s.Reason = ""
	s.UpdatedAt = now
}

// ExecutionGuard centraliza a regra de permissão de execução para agentes e conectores.
type ExecutionGuard struct{}

func NewExecutionGuard() *ExecutionGuard {
	return &ExecutionGuard{}
}

// Allow verifica se um ator (agente ou conector) tem autorização para executar ações.
// Nenhuma execução é permitida se o sistema estiver pausado ou se o ator estiver inativo.
func (g *ExecutionGuard) Allow(systemPaused bool, actorActive bool) error {
	if systemPaused {
		return ErrSystemPaused
	}
	if !actorActive {
		return ErrActorInactive
	}
	return nil
}

// CanExecute valida se a entidade Principal pode agir no estado atual do sistema.
func (g *ExecutionGuard) CanExecute(actor Principal, state SystemState) bool {
	if actor.IsOwner() {
		return true
	}
	if agent, ok := actor.(AgentPrincipal); ok {
		return g.Allow(state.Paused, agent.Active) == nil
	}
	return false
}

