package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/portal"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

var (
	ErrCardNotFound       = errors.New("cartão de projeto não encontrado")
	ErrInvalidPortalToken = errors.New("link de acesso inválido ou expirado")
)

type PortalUseCases struct {
	uow       UnitOfWork
	publisher EventPublisher
	baseURL   string
}

func NewPortalUseCases(uow UnitOfWork, publisher EventPublisher, baseURL string) *PortalUseCases {
	return &PortalUseCases{
		uow:       uow,
		publisher: publisher,
		baseURL:   baseURL,
	}
}

type PortalStatusDTO struct {
	CardID         string     `json:"card_id"`
	IsActive       bool       `json:"is_active"`
	PortalURL      string     `json:"portal_url,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	AcceptedBy     string     `json:"accepted_by,omitempty"`
}

// GeneratePortalLink gera um token criptograficamente seguro e retorna o link para o cliente
func (uc *PortalUseCases) GeneratePortalLink(ctx context.Context, principal system.OwnerPrincipal, cardID string, daysValid int) (string, string, error) {
	if cardID == "" {
		return "", "", fmt.Errorf("card_id é obrigatório")
	}
	if daysValid <= 0 {
		daysValid = 30
	}

	// Token aleatório criptograficamente seguro (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("falha ao gerar token aleatório: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	// Hash SHA-256 do token para salvar no banco
	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	now := time.Now()
	expiresAt := now.Add(time.Duration(daysValid) * 24 * time.Hour)

	var portalURL string

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		cardItem, err := tx.Cards.GetByID(ctx, cardID)
		if err != nil || cardItem == nil {
			return ErrCardNotFound
		}

		// Revoga tokens anteriores deste cartão
		_ = tx.PortalTokens.RevokeByCardID(ctx, cardID, now)

		portalToken := portal.NewPortalToken(cardID, cardItem.ClientID, principal.ID, tokenHash, expiresAt)
		if err := tx.PortalTokens.Create(ctx, portalToken); err != nil {
			return fmt.Errorf("erro ao registrar token do portal: %w", err)
		}

		portalURL = fmt.Sprintf("%s/portal?token=%s", uc.baseURL, rawToken)
		return nil
	})

	if err != nil {
		return "", "", err
	}

	return rawToken, portalURL, nil
}

// RevokePortalLink revoga o acesso ativo do cliente a este cartão
func (uc *PortalUseCases) RevokePortalLink(ctx context.Context, _ system.OwnerPrincipal, cardID string) error {
	now := time.Now()
	return uc.uow.RunInTx(ctx, func(tx Repositories) error {
		return tx.PortalTokens.RevokeByCardID(ctx, cardID, now)
	})
}

// GetPortalTokenStatus retorna os dados de status para visualização interna do dono
func (uc *PortalUseCases) GetPortalTokenStatus(ctx context.Context, _ system.OwnerPrincipal, cardID string) (*PortalStatusDTO, error) {
	var status *PortalStatusDTO
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		token, err := tx.PortalTokens.GetByCardID(ctx, cardID)
		if err != nil || token == nil {
			status = &PortalStatusDTO{
				CardID:   cardID,
				IsActive: false,
			}
			return nil
		}

		now := time.Now()
		status = &PortalStatusDTO{
			CardID:         token.CardID,
			IsActive:       token.IsActive(now),
			ExpiresAt:      &token.ExpiresAt,
			LastAccessedAt: token.LastAccessedAt,
			AcceptedAt:     token.AcceptedAt,
			AcceptedBy:     token.AcceptedBy,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return status, nil
}

// GetClientPortal retorna o DTO estritamente isolado para o cliente
func (uc *PortalUseCases) GetClientPortal(ctx context.Context, rawToken string, clientIP, userAgent string) (*portal.ClientPortalDTO, error) {
	if rawToken == "" {
		return nil, ErrInvalidPortalToken
	}

	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	var clientView *portal.ClientPortalDTO

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		token, err := tx.PortalTokens.GetByHash(ctx, tokenHash)
		if err != nil || token == nil {
			return ErrInvalidPortalToken
		}

		now := time.Now()
		if !token.IsActive(now) {
			return ErrInvalidPortalToken
		}

		// Atualiza último acesso
		token.Touch(now)
		_ = tx.PortalTokens.Update(ctx, token)

		cardItem, err := tx.Cards.GetByID(ctx, token.CardID)
		if err != nil || cardItem == nil {
			return ErrInvalidPortalToken
		}

		clientName := "Cliente"
		if clientDTO, err := tx.Clients.GetByID(ctx, cardItem.ClientID); err == nil && clientDTO != nil {
			clientName = clientDTO.Name
		}

		appName := "Projeto"
		if cardItem.AppID != nil {
			if appDTO, err := tx.Apps.GetByID(ctx, *cardItem.AppID); err == nil && appDTO != nil {
				appName = appDTO.Name
			}
		}

		clientView = buildClientPortalDTO(cardItem, clientName, appName, token)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return clientView, nil
}

// AcceptClientDelivery registra o aceite formal do cliente na etapa de entrega
func (uc *PortalUseCases) AcceptClientDelivery(ctx context.Context, rawToken string, acceptedBy, clientIP, userAgent string) (*portal.ClientPortalDTO, error) {
	if rawToken == "" {
		return nil, ErrInvalidPortalToken
	}
	if acceptedBy == "" {
		return nil, portal.ErrMissingAcceptedBy
	}

	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	var clientView *portal.ClientPortalDTO

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		token, err := tx.PortalTokens.GetByHash(ctx, tokenHash)
		if err != nil || token == nil {
			return ErrInvalidPortalToken
		}

		now := time.Now()
		if !token.IsActive(now) {
			return ErrInvalidPortalToken
		}

		cardItem, err := tx.Cards.GetByID(ctx, token.CardID)
		if err != nil || cardItem == nil {
			return ErrInvalidPortalToken
		}

		// Aceite só é permitido na etapa de Entrega (ou se já estiver entregue)
		if cardItem.Stage != card.StageEntregaAprovar && cardItem.Stage != card.StageEntregue {
			return portal.ErrDeliveryNotReady
		}

		if err := token.Accept(acceptedBy, clientIP, userAgent, string(cardItem.Stage), now); err != nil {
			return err
		}

		if err := tx.PortalTokens.Update(ctx, token); err != nil {
			return fmt.Errorf("falha ao salvar aceite no banco: %w", err)
		}

		// Grava auditoria imutável (append-only)
		auditLog := AuditLogDTO{
			ID:         uuid.NewString(),
			OwnerID:    token.OwnerID,
			Action:     "client_portal_delivery_accepted",
			ActorType:  "client",
			TargetType: "card",
			TargetID:   &token.CardID,
			Meta: map[string]any{
				"accepted_by":  acceptedBy,
				"client_ip":    clientIP,
				"user_agent":   userAgent,
				"state_hash":   token.AcceptedStateHash,
				"accepted_at":  now.Format(time.RFC3339),
			},
			At: now,
		}
		_ = tx.Audit.Record(ctx, auditLog)

		// Notifica atualização em tempo real
		if uc.publisher != nil {
			_ = uc.publisher.Publish(ctx, TopicCardsChanged, map[string]any{"card_id": token.CardID})
		}

		clientName := "Cliente"
		if clientDTO, err := tx.Clients.GetByID(ctx, cardItem.ClientID); err == nil && clientDTO != nil {
			clientName = clientDTO.Name
		}

		appName := "Projeto"
		if cardItem.AppID != nil {
			if appDTO, err := tx.Apps.GetByID(ctx, *cardItem.AppID); err == nil && appDTO != nil {
				appName = appDTO.Name
			}
		}

		clientView = buildClientPortalDTO(cardItem, clientName, appName, token)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return clientView, nil
}

// buildClientPortalDTO constrói a resposta restrita com as 4 etapas
func buildClientPortalDTO(cardItem *card.Card, clientName, appName string, token *portal.PortalToken) *portal.ClientPortalDTO {
	isAccepted := token.AcceptedAt != nil
	stage := cardItem.Stage

	stages := []portal.PortalStageView{
		{
			Key:         "requisitos",
			Label:       "Requisitos e Escopo",
			Description: "Definição das histórias de usuário e aprovação de escopo",
			Status:      "pendente",
			Order:       1,
		},
		{
			Key:         "construcao",
			Label:       "Desenvolvimento e Construção",
			Description: "Implementação da solução e arquitetura do software",
			Status:      "pendente",
			Order:       2,
		},
		{
			Key:         "testes",
			Label:       "Testes e Validação",
			Description: "Verificação de qualidade, testes automatizados e segurança",
			Status:      "pendente",
			Order:       3,
		},
		{
			Key:         "entrega",
			Label:       "Entrega e Aceite",
			Description: "Homologação final e liberação para produção",
			Status:      "pendente",
			Order:       4,
		},
	}

	progress := 25
	canAccept := false

	switch stage {
	case card.StageEntrada, card.StagePropostaAprovar, card.StageAguardandoCliente, card.StagePromptAprovar:
		stages[0].Status = "em_andamento"
		progress = 25
	case card.StageEmExecucao:
		stages[0].Status = "concluido"
		stages[1].Status = "em_andamento"
		progress = 60
	case card.StageEntregaAprovar:
		stages[0].Status = "concluido"
		stages[1].Status = "concluido"
		stages[2].Status = "concluido"
		stages[3].Status = "em_andamento"
		progress = 90
		canAccept = !isAccepted
	case card.StageEntregue:
		stages[0].Status = "concluido"
		stages[1].Status = "concluido"
		stages[2].Status = "concluido"
		stages[3].Status = "concluido"
		progress = 100
		canAccept = false
	default:
		stages[0].Status = "em_andamento"
	}

	stageLabels := map[card.Stage]string{
		card.StageEntrada:           "Levantamento Inicial",
		card.StagePropostaAprovar:   "Proposta em Elaboração",
		card.StageAguardandoCliente: "Aguardando Alinhamento",
		card.StagePromptAprovar:     "Planejamento Concluído",
		card.StageEmExecucao:        "Construção em Andamento",
		card.StageEntregaAprovar:    "Pronto para Homologação",
		card.StageEntregue:          "Projeto Entregue",
	}

	currentLabel := stageLabels[stage]
	if currentLabel == "" {
		currentLabel = "Em Andamento"
	}

	return &portal.ClientPortalDTO{
		ProjectTitle:       cardItem.Title,
		ClientName:         clientName,
		AppName:            appName,
		CurrentStageLabel:  currentLabel,
		ProgressPercentage: progress,
		Stages:             stages,
		CanAccept:          canAccept,
		IsAccepted:         isAccepted,
		AcceptedAt:         token.AcceptedAt,
		AcceptedBy:         token.AcceptedBy,
	}
}
