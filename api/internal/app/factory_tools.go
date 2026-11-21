package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

var ErrFactoryToolUnavailable = errors.New("ferramenta indisponível: configure a chave no servidor")
var ErrFactoryToolSelection = errors.New("selecione um provedor suportado e confirme a ciência do plano/custo antes de ativar")

func isSupportedFactoryProvider(provider string) bool {
	switch provider {
	case "gemini", "groq", "openai", "anthropic", "openrouter":
		return true
	default:
		return false
	}
}

// FactoryToolSelection usa a tabela de conectores existente, sem armazenar chaves.
type FactoryToolSelection struct {
	Provider string
	Active   bool
}

type FactoryToolSettingsRepository interface {
	Get(context.Context) (*FactoryToolSelection, error)
	Save(context.Context, string, FactoryToolSelection) error
}

type FactoryToolDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Configured bool   `json:"configured"`
	Active     bool   `json:"active"`
}

// FactoryTools também implementa AgentRunner: a seleção é conferida antes de cada etapa.
type FactoryTools struct {
	uow            UnitOfWork
	guard          *system.ExecutionGuard
	runners        map[string]pipeline.AgentRunner
	legacyProvider string
}

func NewFactoryTools(uow UnitOfWork, guard *system.ExecutionGuard, runners map[string]pipeline.AgentRunner, legacyProvider string) *FactoryTools {
	return &FactoryTools{uow: uow, guard: guard, runners: runners, legacyProvider: legacyProvider}
}

func (uc *FactoryTools) selection(ctx context.Context, tx Repositories) (FactoryToolSelection, error) {
	if tx.FactoryTools == nil {
		return FactoryToolSelection{}, fmt.Errorf("repositório de ferramentas indisponível")
	}
	stored, err := tx.FactoryTools.Get(ctx)
	if err != nil {
		return FactoryToolSelection{}, err
	}
	if stored != nil {
		return *stored, nil
	}
	// Preserva a seleção já configurada antes da primeira escolha pelo painel.
	return FactoryToolSelection{Provider: uc.legacyProvider, Active: uc.runners[uc.legacyProvider] != nil}, nil
}

func (uc *FactoryTools) List(ctx context.Context) ([]FactoryToolDTO, error) {
	var selected FactoryToolSelection
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		selected, err = uc.selection(ctx, tx)
		return err
	})
	if err != nil {
		return nil, err
	}
	items := []FactoryToolDTO{
		{ID: "gemini", Name: "Google Gemini"},
		{ID: "groq", Name: "Groq"},
		{ID: "openai", Name: "OpenAI (GPT)"},
		{ID: "anthropic", Name: "Anthropic (Claude)"},
		{ID: "openrouter", Name: "OpenRouter (Multi-modelo)"},
	}
	for i := range items {
		items[i].Configured = uc.runners[items[i].ID] != nil
		items[i].Active = selected.Active && selected.Provider == items[i].ID
	}
	return items, nil
}

func (uc *FactoryTools) Select(ctx context.Context, owner system.OwnerPrincipal, provider string, active, freeTierConfirmed bool) error {
	if owner.ID == "" || owner.AAL != "aal2" {
		return system.ErrOwnerRequired
	}
	if !isSupportedFactoryProvider(provider) {
		return ErrFactoryToolSelection
	}
	if active && !freeTierConfirmed {
		return ErrFactoryToolSelection
	}
	if active && uc.runners[provider] == nil {
		return ErrFactoryToolUnavailable
	}
	return uc.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.FactoryTools == nil {
			return fmt.Errorf("repositório de ferramentas indisponível")
		}
		if tx.SystemState == nil {
			return fmt.Errorf("estado do sistema indisponível")
		}
		st, err := tx.SystemState.Get(ctx)
		if err != nil {
			return err
		}
		if st == nil {
			return fmt.Errorf("estado do sistema indisponível")
		}
		if active {
			if err := uc.guard.Allow(st.Paused, true); err != nil {
				return err
			}
		}
		if err := tx.FactoryTools.Save(ctx, owner.ID, FactoryToolSelection{Provider: provider, Active: active}); err != nil {
			return err
		}
		return tx.Audit.Record(ctx, AuditLogDTO{
			OwnerID: owner.ID, ActorID: &owner.ID, ActorType: "owner", Action: "factory.tool_selected", TargetType: "connector",
			Meta: map[string]any{"provider": provider, "active": active, "free_tier_confirmed": freeTierConfirmed},
		})
	})
}

func (uc *FactoryTools) selectedRunner(ctx context.Context) (pipeline.AgentRunner, error) {
	var selected FactoryToolSelection
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		selected, err = uc.selection(ctx, tx)
		if err != nil {
			return err
		}
		if tx.SystemState == nil {
			return fmt.Errorf("estado do sistema indisponível")
		}
		st, err := tx.SystemState.Get(ctx)
		if err != nil {
			return err
		}
		if st == nil {
			return fmt.Errorf("estado do sistema indisponível")
		}
		return uc.guard.Allow(st.Paused, selected.Active)
	})
	if err != nil {
		return nil, err
	}
	runner := uc.runners[selected.Provider]
	if runner == nil {
		return nil, ErrFactoryToolUnavailable
	}
	return runner, nil
}

// Ready rejeita um build sem ferramenta habilitada antes de mudar o cartão.
func (uc *FactoryTools) Ready(ctx context.Context) error {
	if _, err := uc.selectedRunner(ctx); err != nil {
		return err
	}
	for _, role := range []pipeline.Role{pipeline.RoleRequisitos, pipeline.RoleArquiteto, pipeline.RoleDev, pipeline.RoleQA, pipeline.RoleSeguranca, pipeline.RolePublicador, pipeline.RoleSuporte} {
		if err := uc.checkAgent(ctx, role); err != nil {
			return err
		}
	}
	return nil
}

func (uc *FactoryTools) checkAgent(ctx context.Context, role pipeline.Role) error {
	return uc.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.Agents == nil {
			return fmt.Errorf("estado dos agentes indisponível")
		}
		agents, err := tx.Agents.List(ctx)
		if err != nil {
			return fmt.Errorf("falha ao consultar agentes: %w", err)
		}
		for _, agent := range agents {
			if agent.Role == string(role) {
				if !agent.Active {
					return fmt.Errorf("agente %s desligado: %w", role, system.ErrActorInactive)
				}
				return nil
			}
		}
		return fmt.Errorf("agente %s não encontrado", role)
	})
}

func (uc *FactoryTools) RunStep(ctx context.Context, role pipeline.Role, prompt string, meta map[string]any) (*pipeline.RunnerResult, error) {
	runner, err := uc.selectedRunner(ctx)
	if err != nil {
		return nil, err
	}
	if err := uc.checkAgent(ctx, role); err != nil {
		return nil, err
	}
	// Não troca automaticamente de provedor ao esgotar cota ou receber erro.
	return runner.RunStep(ctx, role, prompt, meta)
}
