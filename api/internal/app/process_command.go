package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/command"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type ProcessCommandInput struct {
	Owner system.OwnerPrincipal
	Text  string
}

type CommandResult struct {
	Intent               string         `json:"intent"`
	Message              string         `json:"message"`
	RequiresConfirmation bool           `json:"requires_confirmation"`
	ConfirmationToken    string         `json:"confirmation_token,omitempty"`
	ConfirmationPrompt   string         `json:"confirmation_prompt,omitempty"`
	Data                 map[string]any `json:"data,omitempty"`
}

type ProcessCommandUseCase struct {
	cardRepo     CardRepository
	moneyReader  MoneyWaitingReader
	financeRepo  FinanceRepository
	decideGateUC *DecideGateUseCase
	pauseUC      *PauseSystemUseCase
	resumeUC     *ResumeSystemUseCase
	confirmSvc   *command.ConfirmationService
	startBuildUC *StartBuildUseCase
	createCardUC *CreateCardUseCase
}

func (uc *ProcessCommandUseCase) WithStartBuild(sb *StartBuildUseCase) *ProcessCommandUseCase {
	uc.startBuildUC = sb
	return uc
}

func (uc *ProcessCommandUseCase) WithCreateCard(cc *CreateCardUseCase) *ProcessCommandUseCase {
	uc.createCardUC = cc
	return uc
}

func NewProcessCommandUseCase(
	cardRepo CardRepository,
	moneyReader MoneyWaitingReader,
	financeRepo FinanceRepository,
	decideGateUC *DecideGateUseCase,
	pauseUC *PauseSystemUseCase,
	resumeUC *ResumeSystemUseCase,
	confirmSvc *command.ConfirmationService,
) *ProcessCommandUseCase {
	return &ProcessCommandUseCase{
		cardRepo:     cardRepo,
		moneyReader:  moneyReader,
		financeRepo:  financeRepo,
		decideGateUC: decideGateUC,
		pauseUC:      pauseUC,
		resumeUC:     resumeUC,
		confirmSvc:   confirmSvc,
	}
}

func (uc *ProcessCommandUseCase) Execute(ctx context.Context, in ProcessCommandInput) (*CommandResult, error) {
	parsed, err := command.ParseCommand(in.Text)
	if err != nil {
		return nil, err
	}

	switch parsed.Intent {
	case command.IntentStatus:
		cents, count, err := uc.moneyReader.GetMoneyWaiting(ctx)
		if err != nil {
			return nil, fmt.Errorf("ler status: %w", err)
		}
		formatted := fmt.Sprintf("R$ %.2f", float64(cents)/100.0)
		return &CommandResult{
			Intent:  string(parsed.Intent),
			Message: fmt.Sprintf("%s parados esperando você · %d decisões pendentes", formatted, count),
			Data: map[string]any{
				"total_cents": cents,
				"count":       count,
			},
		}, nil

	case command.IntentFinances:
		overview, err := uc.financeRepo.GetOverview(ctx)
		if err != nil {
			return nil, fmt.Errorf("ler finanças: %w", err)
		}
		rev := fmt.Sprintf("R$ %.2f", float64(overview.TotalRevenueCents)/100.0)
		cost := fmt.Sprintf("R$ %.2f", float64(overview.TotalCostCents)/100.0)
		profit := fmt.Sprintf("R$ %.2f", float64(overview.NetProfitCents)/100.0)
		return &CommandResult{
			Intent:  string(parsed.Intent),
			Message: fmt.Sprintf("Receita %s, custo %s, lucro %s. %s", rev, cost, profit, overview.Suggestion),
			Data: map[string]any{
				"revenue_cents": overview.TotalRevenueCents,
				"cost_cents":    overview.TotalCostCents,
				"profit_cents":  overview.NetProfitCents,
			},
		}, nil

	case command.IntentApprove:
		match, err := uc.resolveTargetCard(ctx, parsed.Target)
		if err != nil {
			return nil, err
		}

		err = uc.decideGateUC.Execute(ctx, DecideGateInput{
			Owner:    in.Owner,
			CardID:   match.CardID,
			Decision: card.DecisionAprovado,
			Note:     "Aprovado via comando do dono: " + in.Text,
		})
		if err != nil {
			return nil, err
		}

		return &CommandResult{
			Intent:  string(parsed.Intent),
			Message: fmt.Sprintf("Cartão de %s (%s) aprovado e avançado com sucesso!", match.ClientName, match.Title),
			Data: map[string]any{
				"card_id": match.CardID,
			},
		}, nil

	case command.IntentBuild:
		var cardID string
		var cardTitle string
		var clientName string

		match, err := uc.resolveTargetCard(ctx, parsed.Target)
		if err != nil {
			// Se o cartão não existe previamente, cria automaticamente e inicia o pipeline (sprit02)
			if uc.createCardUC != nil {
				title := parsed.Target
				client := "Cliente Direto"
				if idx := strings.Index(strings.ToLower(title), " para "); idx != -1 {
					client = strings.TrimSpace(title[idx+6:])
					title = strings.TrimSpace(title[:idx])
				}
				if title == "" {
					title = "Novo Projeto da Fábrica"
				}
				created, createErr := uc.createCardUC.Execute(ctx, CreateCardCommand{
					Title:      title,
					ClientName: client,
					PriceCents: 350000,
					Stage:      string(card.StageEmExecucao),
				})
				if createErr != nil {
					return nil, fmt.Errorf("falha ao criar cartão para construção: %w", createErr)
				}
				cardID = created.ID
				cardTitle = created.Title
				clientName = created.ClientName
			} else {
				return nil, err
			}
		} else {
			cardID = match.CardID
			cardTitle = match.Title
			clientName = match.ClientName
		}

		if uc.startBuildUC != nil {
			_, err := uc.startBuildUC.Execute(ctx, in.Owner, StartBuildCommand{
				CardID:          cardID,
				SpendLimitCents: 5000,
				Title:           cardTitle,
				ClientName:      clientName,
				PriceCents:      350000,
			})
			if err != nil {
				return nil, err
			}
		}

		return &CommandResult{
			Intent:  string(parsed.Intent),
			Message: fmt.Sprintf("Pipeline da Fábrica Autônoma iniciado para '%s' (%s)! Os 7 agentes estão construindo a solução no sandbox isolado.", cardTitle, clientName),
			Data: map[string]any{
				"card_id": cardID,
			},
		}, nil

	case command.IntentCreate:
		title := parsed.Target
		clientName := "Cliente Direto"
		if idx := strings.Index(strings.ToLower(title), " para "); idx != -1 {
			clientName = strings.TrimSpace(title[idx+6:])
			title = strings.TrimSpace(title[:idx])
		}
		if title == "" {
			title = "Nova Demanda de Software"
		}
		if clientName == "" {
			clientName = "Cliente Direto"
		}

		if uc.createCardUC != nil {
			created, err := uc.createCardUC.Execute(ctx, CreateCardCommand{
				Title:      title,
				ClientName: clientName,
				PriceCents: 250000,
				Stage:      string(card.StagePropostaAprovar),
			})
			if err != nil {
				return nil, fmt.Errorf("falha ao criar demanda: %w", err)
			}
			return &CommandResult{
				Intent:  string(parsed.Intent),
				Message: fmt.Sprintf("Demanda '%s' criada para %s! Já está disponível na coluna Propostas.", created.Title, created.ClientName),
				Data: map[string]any{
					"card_id":     created.ID,
					"title":       created.Title,
					"client_name": created.ClientName,
					"stage":       created.Stage,
				},
			}, nil
		}
		return &CommandResult{
			Intent:  string(parsed.Intent),
			Message: fmt.Sprintf("Demanda '%s' registrada.", title),
		}, nil

	case command.IntentReject:
		// Ação destrutiva: Exige token de confirmação de uso único (F3-02)
		match, err := uc.resolveTargetCard(ctx, parsed.Target)
		if err != nil {
			return nil, err
		}

		prompt := fmt.Sprintf("Tem certeza que deseja recusar o cartão '%s' de %s? O cartão sairá do Kanban e irá para o histórico recusado.", match.Title, match.ClientName)
		act, err := uc.confirmSvc.IssueToken(command.ActionTypeRejectCard, match.CardID, prompt, map[string]any{
			"raw_command": in.Text,
			"client_name": match.ClientName,
		})
		if err != nil {
			return nil, fmt.Errorf("emitir token de confirmação: %w", err)
		}

		return &CommandResult{
			Intent:               string(parsed.Intent),
			Message:              prompt,
			RequiresConfirmation: true,
			ConfirmationToken:    act.Token,
			ConfirmationPrompt:   prompt,
			Data: map[string]any{
				"card_id": match.CardID,
			},
		}, nil

	case command.IntentDelete:
		match, err := uc.resolveTargetCard(ctx, parsed.Target)
		if err != nil {
			return nil, err
		}

		prompt := fmt.Sprintf("Tem certeza que deseja EXCLUIR DEFINITIVAMENTE a demanda '%s' de %s? Esta ação não pode ser desfeita.", match.Title, match.ClientName)
		act, err := uc.confirmSvc.IssueToken(command.ActionTypeDeleteCard, match.CardID, prompt, map[string]any{
			"raw_command": in.Text,
			"client_name": match.ClientName,
		})
		if err != nil {
			return nil, fmt.Errorf("emitir token de confirmação: %w", err)
		}

		return &CommandResult{
			Intent:               string(parsed.Intent),
			Message:              prompt,
			RequiresConfirmation: true,
			ConfirmationToken:    act.Token,
			ConfirmationPrompt:   prompt,
			Data: map[string]any{
				"card_id": match.CardID,
			},
		}, nil

	case command.IntentPauseSystem:
		// Ação destrutiva de emergência: exige confirmação se não tiver confirmação explícita
		prompt := "Tem certeza que deseja parar o sistema? Todos os agentes e conexões serão bloqueados de agir."
		act, err := uc.confirmSvc.IssueToken(command.ActionTypePauseSystem, "system", prompt, map[string]any{
			"raw_command": in.Text,
		})
		if err != nil {
			return nil, fmt.Errorf("emitir token de confirmação: %w", err)
		}

		return &CommandResult{
			Intent:               string(parsed.Intent),
			Message:              prompt,
			RequiresConfirmation: true,
			ConfirmationToken:    act.Token,
			ConfirmationPrompt:   prompt,
		}, nil

	case command.IntentResumeSystem:
		err := uc.resumeUC.Execute(ctx, in.Owner)
		if err != nil {
			return nil, err
		}
		return &CommandResult{
			Intent:  string(parsed.Intent),
			Message: "Sistema retomado. Os conectores seguem desligados até você ligar.",
		}, nil

	default:
		return &CommandResult{
			Intent:  string(command.IntentUnknown),
			Message: "Comando não compreendido. Tente: 'status', 'lucro', 'aprovar [cliente]' ou 'recusar [cliente]'.",
		}, nil
	}
}

func (uc *ProcessCommandUseCase) resolveTargetCard(ctx context.Context, target string) (*command.TargetMatch, error) {
	cards, err := uc.cardRepo.ListEnriched(ctx)
	if err != nil {
		return nil, fmt.Errorf("buscar cartões: %w", err)
	}

	available := make([]command.TargetMatch, 0, len(cards))
	for _, c := range cards {
		if c.Stage != "entregue" && c.Stage != "recusado" {
			available = append(available, command.TargetMatch{
				CardID:     c.ID,
				ClientName: c.ClientName,
				Title:      c.Title,
				Stage:      c.Stage,
				PriceCents: c.PriceCents,
			})
		}
	}

	return command.ResolveTarget(target, available)
}

// ConfirmActionUseCase executa a ação destrutiva após o dono submeter o token de uso único.
type ConfirmActionUseCase struct {
	confirmSvc   *command.ConfirmationService
	decideGateUC *DecideGateUseCase
	pauseUC      *PauseSystemUseCase
	deleteCardUC *DeleteCardUseCase
}

func NewConfirmActionUseCase(
	confirmSvc *command.ConfirmationService,
	decideGateUC *DecideGateUseCase,
	pauseUC *PauseSystemUseCase,
	deleteCardUC *DeleteCardUseCase,
) *ConfirmActionUseCase {
	return &ConfirmActionUseCase{
		confirmSvc:   confirmSvc,
		decideGateUC: decideGateUC,
		pauseUC:      pauseUC,
		deleteCardUC: deleteCardUC,
	}
}

func (uc *ConfirmActionUseCase) Execute(ctx context.Context, owner system.OwnerPrincipal, token string) (string, error) {
	act, err := uc.confirmSvc.Consume(token)
	if err != nil {
		return "", err
	}

	switch act.Type {
	case command.ActionTypeDeleteCard:
		if uc.deleteCardUC != nil {
			err := uc.deleteCardUC.Execute(ctx, act.TargetID)
			if err != nil {
				return "", err
			}
		}
		return "Demanda excluída permanentemente com sucesso.", nil

	case command.ActionTypeRejectCard:
		err := uc.decideGateUC.Execute(ctx, DecideGateInput{
			Owner:    owner,
			CardID:   act.TargetID,
			Decision: card.DecisionRecusado,
			Note:     "Ação destrutiva confirmada pelo dono com token de uso único",
		})
		if err != nil {
			return "", err
		}
		return "Cartão recusado e arquivado com sucesso.", nil

	case command.ActionTypePauseSystem:
		err := uc.pauseUC.Execute(ctx, owner, "Parada de emergência confirmada pelo dono via comando")
		if err != nil {
			return "", err
		}
		return "Sistema pausado com sucesso. Nenhum agente nem conexão pode agir.", nil

	default:
		return "", fmt.Errorf("tipo de ação desconhecido: %s", act.Type)
	}
}
