package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type StartBuildCommand struct {
	CardID          string `json:"card_id"`
	SpendLimitCents int64  `json:"spend_limit_cents"`
}

type StartBuildUseCase struct {
	uow    UnitOfWork
	events EventPublisher
	guard  *system.ExecutionGuard
	runner pipeline.AgentRunner
}

func NewStartBuildUseCase(
	uow UnitOfWork,
	events EventPublisher,
	guard *system.ExecutionGuard,
	runner pipeline.AgentRunner,
) *StartBuildUseCase {
	return &StartBuildUseCase{
		uow:    uow,
		events: events,
		guard:  guard,
		runner: runner,
	}
}

func (uc *StartBuildUseCase) Execute(ctx context.Context, principal system.OwnerPrincipal, cmd StartBuildCommand) (*PipelineDTO, error) {
	var pipelineDTO PipelineDTO
	var cardTitle string
	var cardClientID string

	// 1. Iniciar ou recuperar Pipeline e mudar cartão para 'em_execucao'
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		sysState, err := tx.SystemState.Get(ctx)
		if err != nil {
			return fmt.Errorf("falha ao verificar estado do sistema: %w", err)
		}
		if sysState.Paused {
			return card.ErrSystemPaused
		}

		c, err := tx.Cards.GetByID(ctx, cmd.CardID)
		if err != nil {
			return fmt.Errorf("cartão não encontrado: %w", err)
		}
		cardTitle = c.Title
		cardClientID = c.ClientID

		c.Stage = card.StageEmExecucao
		c.UpdatedAt = time.Now()
		if err := tx.Cards.Update(ctx, c); err != nil {
			return fmt.Errorf("falha ao atualizar estágio do cartão: %w", err)
		}

		limit := cmd.SpendLimitCents
		if limit <= 0 {
			limit = 5000 // R$ 50,00 padrão
		}

		pipelineDTO = PipelineDTO{
			CardID:          cmd.CardID,
			Stage:           string(card.StageEmExecucao),
			ActiveRole:      string(pipeline.RoleRequisitos),
			SpendTotalCents: 0,
			SpendLimitCents: limit,
			Status:          string(pipeline.StatusRunning),
			CurrentTask:     "Iniciando pipeline de fábrica autônoma",
		}

		if err := tx.Pipelines.Upsert(ctx, &pipelineDTO); err != nil {
			return err
		}

		_ = tx.CardEvents.Create(ctx, CardEventDTO{
			CardID:    cmd.CardID,
			ActorType: "owner",
			Kind:      "build_started",
			Payload: map[string]any{
				"spend_limit_cents": limit,
			},
			At: time.Now(),
		})

		_ = tx.Audit.Record(ctx, AuditLogDTO{
			Action:     "build_pipeline_started",
			ActorType:  "owner",
			TargetType: "card",
			TargetID:   &cmd.CardID,
			Meta: map[string]any{
				"spend_limit_cents": limit,
			},
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = uc.events.Publish(ctx, TopicPipelineChanged, pipelineDTO)
	_ = uc.events.Publish(ctx, TopicCardsChanged, map[string]string{"id": cmd.CardID, "stage": string(card.StageEmExecucao)})

	// 2. Execução assíncrona do Pipeline de Agentes
	go uc.runFactoryPipeline(context.Background(), cmd.CardID, cardTitle, cardClientID, pipelineDTO.SpendLimitCents)

	return &pipelineDTO, nil
}

func (uc *StartBuildUseCase) runFactoryPipeline(ctx context.Context, cardID, title, clientID string, spendLimitCents int64) {
	roles := []pipeline.Role{
		pipeline.RoleRequisitos,
		pipeline.RoleArquiteto,
		pipeline.RoleDev,
		pipeline.RoleQA,
		pipeline.RoleSeguranca,
		pipeline.RolePublicador,
	}

	totalSpend := int64(0)
	var generatedStories []map[string]any
	var acceptanceCode string

	for _, role := range roles {
		// Verifica teto de gastos
		if totalSpend >= spendLimitCents {
			_ = uc.uow.RunInTx(ctx, func(tx Repositories) error {
				p, _ := tx.Pipelines.GetByCardID(ctx, cardID)
				if p != nil {
					p.Status = string(pipeline.StatusPausedSpendLimit)
					p.CurrentTask = fmt.Sprintf("Pipeline pausado: atingiu teto de gastos de R$ %.2f", float64(spendLimitCents)/100.0)
					_ = tx.Pipelines.Upsert(ctx, p)
				}
				_ = tx.CardEvents.Create(ctx, CardEventDTO{
					CardID:    cardID,
					ActorType: "system",
					Kind:      "pipeline_spend_limit_reached",
					Payload:   map[string]any{"spend_cents": totalSpend, "limit_cents": spendLimitCents},
					At:        time.Now(),
				})
				return nil
			})
			return
		}

		// Atualiza papel ativo
		_ = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			p, _ := tx.Pipelines.GetByCardID(ctx, cardID)
			if p != nil {
				p.ActiveRole = string(role)
				p.CurrentTask = fmt.Sprintf("Agente %s em execução", role)
				_ = tx.Pipelines.Upsert(ctx, p)
			}
			return nil
		})

		// Executa agente via AgentRunner
		contextMeta := map[string]any{
			"card_id":       cardID,
			"title":         title,
			"client_id":     clientID,
			"stories":       generatedStories,
			"previous_code": acceptanceCode,
		}

		res, err := uc.runner.RunStep(ctx, role, fmt.Sprintf("Executar etapa para o projeto %s", title), contextMeta)
		if err != nil {
			_ = uc.uow.RunInTx(ctx, func(tx Repositories) error {
				p, _ := tx.Pipelines.GetByCardID(ctx, cardID)
				if p != nil {
					p.Status = string(pipeline.StatusFailed)
					p.CurrentTask = fmt.Sprintf("Falha no agente %s: %v", role, err)
					_ = tx.Pipelines.Upsert(ctx, p)
				}
				return nil
			})
			return
		}

		totalSpend += res.CostCents

		// Persiste resultados específicos de cada etapa
		_ = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			_ = tx.Pipelines.AddSpend(ctx, cardID, res.CostCents)

			_ = tx.CardEvents.Create(ctx, CardEventDTO{
				CardID:    cardID,
				ActorType: "agent",
				Kind:      fmt.Sprintf("agent_%s_finished", role),
				Payload: map[string]any{
					"output": res.Output,
					"cost":   res.CostCents,
				},
				At: time.Now(),
			})

			switch role {
			case pipeline.RoleRequisitos:
				if stories, ok := res.Artifacts["stories"].([]map[string]any); ok {
					generatedStories = stories
					rawToken := generateSecureToken()
					tokenHash := hashToken(rawToken)
					_ = tx.Specifications.Upsert(ctx, &SpecificationDTO{
						CardID:             cardID,
						UserStories:        stories,
						TestsCode:          "",
						TestsPassed:        false,
						ClientApproved:     false,
						PaymentSignalCents: 15000, // 30% sinal exemplo
						PaymentSignalPaid:  false,
					}, tokenHash)
				}

			case pipeline.RoleDev:
				if code, ok := res.Artifacts["pr_url"].(string); ok {
					acceptanceCode = code
				}

			case pipeline.RoleQA:
				spec, _ := tx.Specifications.GetByCardID(ctx, cardID)
				if spec != nil {
					spec.TestsPassed = true
					spec.TestsCode = acceptanceCode
					_ = tx.Specifications.Upsert(ctx, spec, "")
				}

			case pipeline.RolePublicador:
				previewURL := fmt.Sprintf("https://preview-%s.geevheit.internal", cardID[:8])
				_ = tx.Previews.Create(ctx, &PreviewDTO{
					CardID:     cardID,
					PreviewURL: previewURL,
					Status:     "ready",
					ExpiresAt:  time.Now().Add(48 * time.Hour),
					RollbackPlan: map[string]any{
						"strategy":     "blue_green",
						"image_backup": "geevheit/app:" + cardID[:8] + "-prev",
						"database_tag": "migration_v4",
					},
				})

				// INVARIANTE: O agente Publicador transita para entrega_aprovar, NUNCA aprova nem entrega direto!
				c, _ := tx.Cards.GetByID(ctx, cardID)
				if c != nil {
					c.Stage = card.StageEntregaAprovar
					c.UpdatedAt = time.Now()
					_ = tx.Cards.Update(ctx, c)
				}

				p, _ := tx.Pipelines.GetByCardID(ctx, cardID)
				if p != nil {
					p.Stage = string(card.StageEntregaAprovar)
					p.Status = string(pipeline.StatusCompleted)
					p.CurrentTask = "Build concluído! Aguardando aprovação de entrega pelo Dono."
					_ = tx.Pipelines.Upsert(ctx, p)
				}
			}

			return nil
		})
	}
}

type GetCardPipelineUseCase struct {
	uow UnitOfWork
}

func NewGetCardPipelineUseCase(uow UnitOfWork) *GetCardPipelineUseCase {
	return &GetCardPipelineUseCase{uow: uow}
}

func (uc *GetCardPipelineUseCase) Execute(ctx context.Context, cardID string) (*PipelineDTO, error) {
	var dto *PipelineDTO
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		p, err := tx.Pipelines.GetByCardID(ctx, cardID)
		if err != nil {
			return err
		}
		dto = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dto, nil
}

func generateSecureToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
