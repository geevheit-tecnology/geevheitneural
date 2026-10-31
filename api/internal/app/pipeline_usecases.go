package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type StartBuildCommand struct {
	CardID          string `json:"card_id"`
	SpendLimitCents int64  `json:"spend_limit_cents"`
	Title           string `json:"title,omitempty"`
	ClientName      string `json:"client_name,omitempty"`
	PriceCents      int64  `json:"price_cents,omitempty"`
}

// RequiresMandatoryHumanReview verifica se o projeto lida com código sensível (Princípio 4 - sprit02):
// pagamentos, autenticação, dados pessoais e integrações financeiras.
func RequiresMandatoryHumanReview(title string) bool {
	lower := strings.ToLower(title)
	sensitiveKeywords := []string{
		"pagamento", "pix", "financeir", "fatura", "cobranca", "cobrança",
		"autentica", "login", "senha", "2fa", "totp", "seguranca", "segurança",
		"lgpd", "dados pessoais", "cpf", "clinica", "clínica", "paciente", "banco", "checkout",
	}
	for _, kw := range sensitiveKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// RouteModelForRole seleciona o tier de modelo adequado para cada papel (Fase 10 - sprit02).
func RouteModelForRole(role pipeline.Role) string {
	switch role {
	case pipeline.RoleArquiteto, pipeline.RoleDev, pipeline.RoleSeguranca:
		return "tier_critical_reasoning (Gemini Pro / GPT-OSS-120B)"
	default:
		return "tier_fast_economy (Gemini 1.5 Flash / Qwen-27B)"
	}
}

type StartBuildUseCase struct {
	uow       UnitOfWork
	events    EventPublisher
	guard     *system.ExecutionGuard
	runner    pipeline.AgentRunner
	sandbox   SandboxManager
	publisher WorkspacePublisher
	active    sync.Map
}

func NewStartBuildUseCase(
	uow UnitOfWork,
	events EventPublisher,
	guard *system.ExecutionGuard,
	runner pipeline.AgentRunner,
	opts ...any,
) *StartBuildUseCase {
	uc := &StartBuildUseCase{
		uow:    uow,
		events: events,
		guard:  guard,
		runner: runner,
	}
	for _, opt := range opts {
		if sm, ok := opt.(SandboxManager); ok {
			uc.sandbox = sm
		} else if pub, ok := opt.(WorkspacePublisher); ok {
			uc.publisher = pub
		}
	}
	return uc
}

func (uc *StartBuildUseCase) Execute(ctx context.Context, principal system.OwnerPrincipal, cmd StartBuildCommand) (*PipelineDTO, error) {
	if uc.runner == nil {
		return nil, fmt.Errorf("provedor de IA não configurado; verifique AGENT_RUNNER_TYPE e a chave do provedor")
	}
	if _, loaded := uc.active.LoadOrStore(cmd.CardID, true); loaded {
		return nil, fmt.Errorf("pipeline já está em execução para este cartão")
	}
	lockID := cmd.CardID
	launched := false
	defer func() {
		if !launched {
			uc.active.Delete(lockID)
		}
	}()
	var pipelineDTO PipelineDTO
	var cardTitle string
	var cardClientID string
	var cardPriceCents int64

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
			// Se o cartão veio do estado local do frontend e ainda não foi sincronizado no banco,
			// provisiona automaticamente o cliente e o cartão caso o título tenha sido informado.
			if cmd.Title != "" && tx.Clients != nil {
				clientName := cmd.ClientName
				if clientName == "" {
					clientName = "Cliente Geevheit"
				}
				cl, _ := tx.Clients.FindByName(ctx, clientName)
				if cl == nil {
					cl, err = tx.Clients.Create(ctx, clientName, "painel-dono")
					if err != nil {
						return fmt.Errorf("cartão não encontrado: %w", err)
					}
				}
				cardIDToCreate := cmd.CardID
				if _, parseErr := uuid.Parse(cardIDToCreate); parseErr != nil {
					cardIDToCreate = uuid.NewString()
					cmd.CardID = cardIDToCreate
				}
				now := time.Now()
				price := cmd.PriceCents
				if price <= 0 {
					price = 250000
				}
				_ = tx.Cards.Create(ctx, &CardListItemDTO{
					ID:           cardIDToCreate,
					ClientID:     cl.ID,
					ClientName:   clientName,
					Type:         "orcamento",
					Title:        cmd.Title,
					PriceCents:   price,
					EstCostCents: price / 10,
					Currency:     "BRL",
					Stage:        string(card.StageEmExecucao),
					Priority:     1,
					CreatedAt:    now,
					UpdatedAt:    now,
				})
				c, err = tx.Cards.GetByID(ctx, cmd.CardID)
			}
			if err != nil {
				return fmt.Errorf("cartão não encontrado: %w", err)
			}
		}
		cardTitle = c.Title
		cardClientID = c.ClientID
		cardPriceCents = c.Price.Cents

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
	launched = true
	activeID := lockID
	go func() {
		defer uc.active.Delete(activeID)
		uc.runFactoryPipeline(context.WithoutCancel(ctx), cmd.CardID, cardTitle, cardClientID, cardPriceCents, pipelineDTO.SpendLimitCents)
	}()

	return &pipelineDTO, nil
}

func (uc *StartBuildUseCase) runFactoryPipeline(ctx context.Context, cardID, title, clientID string, priceCents, spendLimitCents int64) {
	startedAt := time.Now()
	roles := []pipeline.Role{
		pipeline.RoleRequisitos,
		pipeline.RoleArquiteto,
		pipeline.RoleDev,
		pipeline.RoleQA,
		pipeline.RoleSeguranca,
		pipeline.RolePublicador,
		pipeline.RoleSuporte,
	}

	totalSpend := int64(0)
	var generatedStories []map[string]any
	var acceptanceCode string
	var generatedFiles []ProjectFile
	var architecture map[string]any
	requiresHumanReview := RequiresMandatoryHumanReview(title)

	for _, role := range roles {
		// 0. Verifica Botão Parar de Emergência a cada etapa (Requisito transversal sprit02:
		// "Botão Parar interrompe pipelines em andamento e descarta ambientes")
		var isPaused bool
		_ = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if tx.SystemState != nil {
				st, err := tx.SystemState.Get(ctx)
				if err == nil && st != nil && st.Paused {
					isPaused = true
				}
			}
			return nil
		})
		if isPaused {
			if uc.sandbox != nil {
				_ = uc.sandbox.CleanupWorkspace(cardID)
			}
			_ = uc.uow.RunInTx(ctx, func(tx Repositories) error {
				p, _ := tx.Pipelines.GetByCardID(ctx, cardID)
				if p != nil {
					p.Status = "paused_emergency"
					p.CurrentTask = "Pipeline interrompido pelo Botão Parar de emergência; ambiente efêmero descartado."
					_ = tx.Pipelines.Upsert(ctx, p)
				}
				_ = tx.CardEvents.Create(ctx, CardEventDTO{
					CardID:    cardID,
					ActorType: "system",
					Kind:      "pipeline_emergency_stopped",
					Payload:   map[string]any{"reason": "Botão Parar acionado pelo dono; sandbox descartada."},
					At:        time.Now(),
				})
				return nil
			})
			_ = uc.events.Publish(ctx, TopicPipelineChanged, map[string]any{"card_id": cardID, "status": "paused_emergency"})
			return
		}

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
			_ = uc.events.Publish(ctx, TopicPipelineChanged, map[string]any{"card_id": cardID, "status": string(pipeline.StatusPausedSpendLimit)})
			return
		}

		routedModel := RouteModelForRole(role)

		// Atualiza papel ativo
		_ = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			p, _ := tx.Pipelines.GetByCardID(ctx, cardID)
			if p != nil {
				p.ActiveRole = string(role)
				p.CurrentTask = fmt.Sprintf("Agente %s em execução [%s]", role, routedModel)
				_ = tx.Pipelines.Upsert(ctx, p)
			}
			return nil
		})
		_ = uc.events.Publish(ctx, TopicPipelineChanged, map[string]any{"card_id": cardID, "active_role": string(role), "status": string(pipeline.StatusRunning)})

		// Executa agente via AgentRunner
		contextMeta := map[string]any{
			"card_id":                    cardID,
			"title":                      title,
			"client_id":                  clientID,
			"stories":                    generatedStories,
			"previous_code":              acceptanceCode,
			"files":                      generatedFiles,
			"architecture":               architecture,
			"routed_model":               routedModel,
			"revisao_humana_obrigatoria": requiresHumanReview,
		}

		res, err := uc.runner.RunStep(ctx, role, fmt.Sprintf("Executar etapa para o projeto %s", title), contextMeta)
		if err == nil && res == nil {
			err = fmt.Errorf("agente retornou resultado vazio")
		}
		if err != nil {
			uc.failPipeline(ctx, cardID, role, err)
			return
		}

		totalSpend += res.CostCents

		// Persiste resultados específicos de cada etapa
		err = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if err := tx.Pipelines.AddSpend(ctx, cardID, res.CostCents); err != nil {
				return err
			}

			if err := tx.CardEvents.Create(ctx, CardEventDTO{
				CardID:    cardID,
				ActorType: "agent",
				Kind:      fmt.Sprintf("agent_%s_finished", role),
				Payload: map[string]any{
					"output":                     res.Output,
					"cost":                       res.CostCents,
					"routed_model":               routedModel,
					"revisao_humana_obrigatoria": requiresHumanReview,
				},
				At: time.Now(),
			}); err != nil {
				return err
			}

			switch role {
			case pipeline.RoleRequisitos:
				var stories []map[string]any
				if s, ok := res.Artifacts["stories"].([]map[string]any); ok {
					stories = s
				} else if rawSlice, ok := res.Artifacts["stories"].([]any); ok {
					for _, item := range rawSlice {
						if m, ok := item.(map[string]any); ok {
							stories = append(stories, m)
						}
					}
				}
				if len(stories) == 0 {
					return fmt.Errorf("agente requisitos não produziu histórias")
				}
				generatedStories = stories
				testCode, _ := res.Artifacts["test_code"].(string)

				rawToken := "sc_" + generateSecureToken()
				tokenHash := hashToken(rawToken)
				signalCents := priceCents * 30 / 100
				if signalCents <= 0 {
					signalCents = 15000
				}
				if err := tx.Specifications.Upsert(ctx, &SpecificationDTO{
					CardID:             cardID,
					UserStories:        stories,
					TestsCode:          testCode,
					TestsPassed:        false,
					ClientApproved:     false,
					PaymentSignalCents: signalCents,
					PaymentSignalPaid:  false,
					ScopeTokenRaw:      rawToken,
				}, tokenHash); err != nil {
					return err
				}

				if tx.ProjectMemory != nil && clientID != "" {
					if err := tx.ProjectMemory.Upsert(ctx, &ProjectMemoryDTO{
						CardID:   cardID,
						ClientID: clientID,
						Key:      "requisitos_bdd",
						Category: "requisito",
						Content:  fmt.Sprintf("Geradas %d user stories BDD executáveis. %s", len(stories), res.Output),
					}); err != nil {
						return err
					}
				}

			case pipeline.RoleArquiteto:
				architecture = res.Artifacts
				if tx.ProjectMemory != nil && clientID != "" {
					pbSelected, _ := res.Artifacts["playbook_selected"].(string)
					if pbSelected == "" {
						pbSelected = "saas_b2b"
					}
					if err := tx.ProjectMemory.Upsert(ctx, &ProjectMemoryDTO{
						CardID:   cardID,
						ClientID: clientID,
						Key:      "decisao_arquitetura",
						Category: "decisao",
						Content:  fmt.Sprintf("Playbook selecionado: %s. %s", pbSelected, res.Output),
					}); err != nil {
						return err
					}
				}

			case pipeline.RoleDev:
				if code, ok := res.Artifacts["pr_url"].(string); ok {
					acceptanceCode = code
				}
				generatedFiles = extractProjectFiles(res.Artifacts["files"], title)
				if len(generatedFiles) == 0 {
					return fmt.Errorf("agente dev não produziu arquivos")
				}
				if uc.sandbox == nil {
					return fmt.Errorf("sandbox não configurada")
				}
				if _, err := uc.sandbox.InitWorkspace(cardID); err != nil {
					return err
				}
				if err := uc.sandbox.WriteFiles(cardID, generatedFiles); err != nil {
					return err
				}
				stored, err := uc.sandbox.ReadFiles(cardID)
				if err != nil {
					return err
				}
				if len(stored) == 0 {
					return fmt.Errorf("workspace vazio após gravação")
				}

			case pipeline.RoleQA:
				if ready, ok := res.Artifacts["ready_for_review"].(bool); ok && !ready {
					return fmt.Errorf("QA rejeitou os artefatos gerados")
				}
				spec, _ := tx.Specifications.GetByCardID(ctx, cardID)
				if spec != nil {
					spec.TestsPassed = false
					if spec.TestsCode == "" {
						spec.TestsCode = acceptanceCode
					}
					if err := tx.Specifications.Upsert(ctx, spec, ""); err != nil {
						return err
					}
				}

			case pipeline.RoleSeguranca:
				// Fase 9 (sprit02): Alimenta automaticamente a Ficha de Risco completa
				lgpdStatus := "Revisão do modelo: " + res.Output
				if requiresHumanReview {
					lgpdStatus = "⚠️ Código sensível: revisão_humana_obrigatória exigida no Portão 3"
				}
				riskDTO := &RiskReportDTO{
					CardID: cardID,
					Tests: []map[string]any{
						{"label": "Testes unitários & BDD", "status": "Sem execução de testes locais comprovada", "level": "warn"},
						{"label": "Acessibilidade & Perf (Lighthouse)", "status": "Lighthouse não executado", "level": "warn"},
					},
					SecretsScan: []map[string]any{
						{"label": "Varredura de segredos e SAST", "status": "Auditoria automatizada não executada", "level": "warn"},
					},
					OutdatedDeps: []map[string]any{
						{"label": "Auditoria de dependências", "status": "Auditoria de dependências não executada", "level": "warn"},
					},
					LgpdPoints: []map[string]any{
						{"label": "Conformidade LGPD & Código Sensível", "status": lgpdStatus, "level": map[bool]string{true: "warn", false: "ok"}[requiresHumanReview]},
					},
					Source: "factory_pipeline",
				}
				if writer, ok := tx.Cards.(RiskReportWriter); ok {
					if err := writer.SaveRiskReport(ctx, riskDTO); err != nil {
						return err
					}
				}
				if tx.ProjectMemory != nil && clientID != "" {
					if err := tx.ProjectMemory.Upsert(ctx, &ProjectMemoryDTO{
						CardID:   cardID,
						ClientID: clientID,
						Key:      "auditoria_seguranca_lgpd",
						Category: "decisao",
						Content:  fmt.Sprintf("Revisão do modelo registrada; auditoria automatizada não executada. Revisão humana obrigatória: %v.", requiresHumanReview),
					}); err != nil {
						return err
					}
				}

			case pipeline.RolePublicador:
				if uc.publisher == nil {
					return fmt.Errorf("publicador não configurado")
				}
				previewURL := ""
				if uc.publisher != nil {
					pubRes, err := uc.publisher.PublishWorkspace(ctx, cardID, fmt.Sprintf("feat(%s): autonomous build", title))
					if err != nil {
						return err
					}
					if pubRes == nil || pubRes.PreviewURL == "" {
						return fmt.Errorf("publicação sem prévia")
					}
					previewURL = pubRes.PreviewURL
				}

				if err := tx.Previews.Create(ctx, &PreviewDTO{
					CardID:     cardID,
					PreviewURL: previewURL,
					Status:     "ready",
					ExpiresAt:  time.Now().Add(48 * time.Hour),
					RollbackPlan: map[string]any{
						"strategy":                   "blue_green_instant_rollback",
						"image_backup":               "geevheit/app:" + cardID[:minLen(len(cardID), 8)] + "-prev",
						"database_tag":               "goose_safe_point",
						"revisao_humana_obrigatoria": requiresHumanReview,
					},
				}); err != nil {
					return err
				}

				// INVARIANTE: O agente Publicador transita para entrega_aprovar, NUNCA aprova nem entrega direto!
				c, _ := tx.Cards.GetByID(ctx, cardID)
				if c != nil {
					c.Stage = card.StageEntregaAprovar
					c.UpdatedAt = time.Now()
					if err := tx.Cards.Update(ctx, c); err != nil {
						return err
					}
				}

				p, _ := tx.Pipelines.GetByCardID(ctx, cardID)
				if p != nil {
					p.Stage = string(card.StageEntregaAprovar)
					p.Status = string(pipeline.StatusRunning)
					taskMsg := "Build e Prévia ao Vivo concluídos! Aguardando aprovação de entrega pelo Dono."
					if requiresHumanReview {
						taskMsg = "[revisão_humana_obrigatória] Build concluído! Código sensível aguarda revisão explícita do Dono."
					}
					p.CurrentTask = taskMsg
					if err := tx.Pipelines.Upsert(ctx, p); err != nil {
						return err
					}
				}

				// Fase 10 (sprit02): Registra métricas reais no Placar da Fábrica
				if tx.FactoryMetrics != nil {
					leadSec := int64(time.Since(startedAt).Seconds())
					if leadSec <= 0 {
						leadSec = 12
					}
					margin := priceCents - totalSpend
					if err := tx.FactoryMetrics.Record(ctx, cardID, leadSec, totalSpend, 0, margin, 60, true, 0); err != nil {
						return err
					}
				}

			case pipeline.RoleSuporte:
				if tx.ProjectMemory != nil && clientID != "" {
					if err := tx.ProjectMemory.Upsert(ctx, &ProjectMemoryDTO{
						CardID:   cardID,
						ClientID: clientID,
						Key:      "monitoramento_pos_entrega",
						Category: "nota",
						Content:  res.Output,
					}); err != nil {
						return err
					}
				}
			}

			return nil
		})
		if err != nil {
			uc.failPipeline(ctx, cardID, role, err)
			return
		}
	}

	if err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		p, err := tx.Pipelines.GetByCardID(ctx, cardID)
		if err != nil {
			return err
		}
		p.Status = string(pipeline.StatusCompleted)
		return tx.Pipelines.Upsert(ctx, p)
	}); err != nil {
		uc.failPipeline(ctx, cardID, pipeline.RoleSuporte, err)
		return
	}

	_ = uc.events.Publish(ctx, TopicCardsChanged, map[string]string{"id": cardID, "stage": string(card.StageEntregaAprovar)})
	_ = uc.events.Publish(ctx, TopicPipelineChanged, map[string]any{"card_id": cardID, "status": string(pipeline.StatusCompleted)})
	_ = uc.events.Publish(ctx, TopicScoreboardChanged, map[string]any{"card_id": cardID})
	_ = uc.events.Publish(ctx, TopicMoneyWaitingChanged, nil)
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
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

func extractProjectFiles(raw any, title string) []ProjectFile {
	var files []ProjectFile

	if rawSlice, ok := raw.([]any); ok {
		for _, item := range rawSlice {
			if m, ok := item.(map[string]any); ok {
				path, _ := m["path"].(string)
				content, _ := m["content"].(string)
				if path != "" && content != "" {
					files = append(files, ProjectFile{
						Path:    path,
						Content: content,
					})
				}
			}
		}
	} else if rawMaps, ok := raw.([]map[string]any); ok {
		for _, m := range rawMaps {
			path, _ := m["path"].(string)
			content, _ := m["content"].(string)
			if path != "" && content != "" {
				files = append(files, ProjectFile{
					Path:    path,
					Content: content,
				})
			}
		}
	}

	return files
}

func (uc *StartBuildUseCase) failPipeline(ctx context.Context, cardID string, role pipeline.Role, cause error) {
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		p, err := tx.Pipelines.GetByCardID(ctx, cardID)
		if err != nil {
			return err
		}
		p.Status = string(pipeline.StatusFailed)
		p.CurrentTask = fmt.Sprintf("Falha na etapa %s: %v", role, cause)
		return tx.Pipelines.Upsert(ctx, p)
	})
	if err != nil {
		slog.Error("pipeline_failure_persistence", "card_id", cardID, "role", role, "error", err)
	}
	slog.Error("pipeline_step_failed", "card_id", cardID, "role", role)
	_ = uc.events.Publish(ctx, TopicPipelineChanged, map[string]any{"card_id": cardID, "status": string(pipeline.StatusFailed)})
}
