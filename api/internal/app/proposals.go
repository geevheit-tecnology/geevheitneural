package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type ProposalUseCases struct {
	uow       UnitOfWork
	publisher EventPublisher
	baseURL   string
}

func NewProposalUseCases(uow UnitOfWork, publisher EventPublisher, baseURL string) *ProposalUseCases {
	return &ProposalUseCases{
		uow:       uow,
		publisher: publisher,
		baseURL:   baseURL,
	}
}

type CreateProposalCommand struct {
	CardID      string `json:"card_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	Body        string `json:"body"`
}

func (uc *ProposalUseCases) CreateProposal(ctx context.Context, principal system.OwnerPrincipal, cmd CreateProposalCommand) (*ProposalDTO, error) {
	if cmd.CardID == "" {
		return nil, fmt.Errorf("card_id é obrigatório")
	}
	if cmd.AmountCents <= 0 {
		return nil, fmt.Errorf("amount_cents deve ser maior que zero")
	}
	if cmd.Currency == "" {
		cmd.Currency = "BRL"
	}

	// Gera token criptográfico seguro de 32 bytes
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("falha ao gerar token de aceite: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	// Hash SHA-256 do token para salvar no banco
	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	now := time.Now()
	proposalID := uuid.New().String()

	dto := &ProposalDTO{
		ID:              proposalID,
		CardID:          cmd.CardID,
		AmountCents:     cmd.AmountCents,
		Currency:        cmd.Currency,
		Body:            cmd.Body,
		AcceptTokenHash: tokenHash,
		AcceptTokenRaw:  rawToken,
		AcceptURL:       fmt.Sprintf("%s/aceite/%s", uc.baseURL, rawToken),
		Status:          "pendente",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		// 1. Cria a proposta
		if err := tx.Proposals.Create(ctx, dto); err != nil {
			return fmt.Errorf("erro ao salvar proposta: %w", err)
		}

		// 2. Atualiza o preço do cartão
		c, err := tx.Cards.GetByID(ctx, cmd.CardID)
		if err != nil {
			return fmt.Errorf("erro ao buscar cartão: %w", err)
		}
		c.Price = money.Money{
			Cents:    cmd.AmountCents,
			Currency: cmd.Currency,
		}
		c.UpdatedAt = now
		if err := tx.Cards.Update(ctx, c); err != nil {
			return fmt.Errorf("erro ao atualizar valor do cartão: %w", err)
		}

		// 3. Registra evento no cartão
		if err := tx.CardEvents.Create(ctx, CardEventDTO{
			ID:        uuid.New().String(),
			CardID:    cmd.CardID,
			Kind:      "proposal_created",
			ActorType: "owner",
			Payload: map[string]any{
				"proposal_id":  proposalID,
				"amount_cents": cmd.AmountCents,
			},
			At: now,
		}); err != nil {
			return fmt.Errorf("erro ao registrar evento: %w", err)
		}

		// 4. Audit log
		return tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.New().String(),
			OwnerID:    principal.ID,
			Action:     "proposal.created",
			ActorType:  "owner",
			TargetType: "proposal",
			TargetID:   &proposalID,
			Meta: map[string]any{
				"card_id":      cmd.CardID,
				"amount_cents": cmd.AmountCents,
			},
			At: now,
		})
	})

	if err != nil {
		return nil, err
	}

	_ = uc.publisher.Publish(ctx, TopicCardsChanged, map[string]string{"card_id": cmd.CardID})
	_ = uc.publisher.Publish(ctx, TopicMoneyWaitingChanged, nil)

	return dto, nil
}

func (uc *ProposalUseCases) GetPublicProposal(ctx context.Context, rawToken string) (*PublicProposalViewDTO, error) {
	if rawToken == "" {
		return nil, fmt.Errorf("token inválido")
	}

	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	var view *PublicProposalViewDTO
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		v, _, err := tx.Proposals.GetByTokenHash(ctx, tokenHash)
		if err != nil {
			return err
		}
		view = v
		return nil
	})

	if err != nil {
		return nil, err
	}
	return view, nil
}

func (uc *ProposalUseCases) AcceptProposal(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return fmt.Errorf("token inválido")
	}

	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	var affectedCardID string
	now := time.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		// 1. Busca proposta
		view, cardID, err := tx.Proposals.GetByTokenHash(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("proposta não encontrada: %w", err)
		}
		if view.Status != "pendente" {
			return fmt.Errorf("proposta já foi %s anteriormente", view.Status)
		}

		affectedCardID = cardID

		// 2. Marca proposta como aceita
		if err := tx.Proposals.Accept(ctx, tokenHash); err != nil {
			return fmt.Errorf("falha ao aceitar proposta: %w", err)
		}

		// 3. Atualiza o cartão de proposta_aprovar para prompt_aprovar
		c, err := tx.Cards.GetByID(ctx, cardID)
		if err != nil {
			return fmt.Errorf("cartão não encontrado: %w", err)
		}

		c.Stage = card.StagePromptAprovar
		c.UpdatedAt = now
		if err := tx.Cards.Update(ctx, c); err != nil {
			return fmt.Errorf("falha ao avançar estágio do cartão: %w", err)
		}

		// 4. Registra evento no cartão
		if err := tx.CardEvents.Create(ctx, CardEventDTO{
			ID:        uuid.New().String(),
			CardID:    cardID,
			Kind:      "proposal_accepted",
			ActorType: "public_client",
			Payload: map[string]any{
				"from_stage": string(card.StagePropostaAprovar),
				"to_stage":   string(card.StagePromptAprovar),
				"message":    "Aceite registrado pelo cliente. Prompt técnico gerado.",
			},
			At: now,
		}); err != nil {
			return fmt.Errorf("falha ao registrar evento: %w", err)
		}

		// 5. Registra no audit log
		return tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.New().String(),
			OwnerID:    "00000000-0000-0000-0000-000000000001",
			Action:     "proposal.accepted",
			ActorType:  "client",
			TargetType: "card",
			TargetID:   &cardID,
			Meta: map[string]any{
				"amount_cents": view.AmountCents,
			},
			At: now,
		})
	})

	if err != nil {
		return err
	}

	_ = uc.publisher.Publish(ctx, TopicCardsChanged, map[string]string{"card_id": affectedCardID})
	_ = uc.publisher.Publish(ctx, TopicMoneyWaitingChanged, nil)

	return nil
}
