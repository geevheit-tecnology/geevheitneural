package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/google/uuid"
)

type IntakeUseCase struct {
	uow       UnitOfWork
	publisher EventPublisher
}

func NewIntakeUseCase(uow UnitOfWork, publisher EventPublisher) *IntakeUseCase {
	return &IntakeUseCase{
		uow:       uow,
		publisher: publisher,
	}
}

type IntakeResult struct {
	CardID  string `json:"card_id,omitempty"`
	Spam    bool   `json:"spam,omitempty"`
	Message string `json:"message"`
}

func (uc *IntakeUseCase) Execute(ctx context.Context, cmd IntakeCommand) (*IntakeResult, error) {
	// 1. Honeypot check: se o campo "website" foi preenchido, é um bot spambot
	if strings.TrimSpace(cmd.Website) != "" {
		return &IntakeResult{
			Spam:    true,
			Message: "Solicitação recebida com sucesso.",
		}, nil
	}

	cmd.ClientName = strings.TrimSpace(cmd.ClientName)
	cmd.Contact = strings.TrimSpace(cmd.Contact)
	cmd.Description = strings.TrimSpace(cmd.Description)

	if cmd.ClientName == "" || cmd.Description == "" {
		return nil, fmt.Errorf("nome do cliente e descrição são obrigatórios")
	}

	var cardID string

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		// 2. Busca ou cria cliente
		client, err := tx.Clients.FindByName(ctx, cmd.ClientName)
		if err != nil {
			return fmt.Errorf("erro ao verificar cliente: %w", err)
		}

		if client == nil {
			client, err = tx.Clients.Create(ctx, cmd.ClientName, cmd.Contact)
			if err != nil {
				return fmt.Errorf("erro ao criar cliente: %w", err)
			}
		}

		// 3. Cria cartão de demanda
		cID := uuid.New().String()
		cardID = cID
		now := time.Now()

		title := cmd.Description
		if len(title) > 60 {
			title = title[:57] + "..."
		}

		cardType := "orcamento"
		if cmd.Type == "bug" || cmd.Type == "suporte" || cmd.Type == "financeiro" {
			cardType = cmd.Type
		}

		cardItem := &CardListItemDTO{
			ID:           cID,
			ClientID:     client.ID,
			ClientName:   client.Name,
			AppName:      "Novo",
			Type:         cardType,
			Title:        title,
			PriceCents:   0, // Aguarda proposta de preço
			EstCostCents: 0,
			Currency:     "BRL",
			Stage:        string(card.StagePropostaAprovar),
			Priority:     0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := tx.Cards.Create(ctx, cardItem); err != nil {
			return fmt.Errorf("erro ao criar cartão: %w", err)
		}

		// 4. Registra evento de histórico
		if err := tx.CardEvents.Create(ctx, CardEventDTO{
			ID:        uuid.New().String(),
			CardID:    cID,
			Kind:      "intake_submitted",
			ActorType: "public",
			Payload: map[string]any{
				"client_name": client.Name,
				"contact":     cmd.Contact,
				"description": cmd.Description,
				"type":        cmd.Type,
			},
			At: now,
		}); err != nil {
			return fmt.Errorf("erro ao registrar evento: %w", err)
		}

		// 5. Mascara IP para LGPD e registra audit log
		ipHash := ""
		if cmd.ClientIP != "" {
			h := sha256.Sum256([]byte(cmd.ClientIP))
			ipHash = hex.EncodeToString(h[:8])
		}

		return tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.New().String(),
			OwnerID:    "00000000-0000-0000-0000-000000000001",
			Action:     "intake.submitted",
			ActorType:  "public",
			TargetType: "card",
			TargetID:   &cID,
			Meta: map[string]any{
				"client_id": client.ID,
				"ip_hash":   ipHash,
			},
			At: now,
		})
	})

	if err != nil {
		return nil, err
	}

	// 6. Notifica via SSE
	_ = uc.publisher.Publish(ctx, TopicCardsChanged, map[string]string{"card_id": cardID})
	_ = uc.publisher.Publish(ctx, TopicMoneyWaitingChanged, nil)

	return &IntakeResult{
		CardID:  cardID,
		Message: "Solicitação recebida com sucesso. Nossa equipe entrará em contato com a proposta.",
	}, nil
}
