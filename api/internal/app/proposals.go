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

// GenerateProposalDocument gera um documento executivo e comercial de proposta e contrato para o cartão
func (uc *ProposalUseCases) GenerateProposalDocument(ctx context.Context, principal system.OwnerPrincipal, cardID string) (*ProposalDocumentDTO, error) {
	if cardID == "" {
		return nil, fmt.Errorf("card_id é obrigatório")
	}

	var doc *ProposalDocumentDTO
	now := time.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		// 1. Busca dados completos do cartão
		cardItem, err := tx.Cards.GetEnrichedByID(ctx, cardID)
		if err != nil {
			return fmt.Errorf("cartão não encontrado: %w", err)
		}

		// 2. Determina escopo, entregáveis e cronograma
		scopeSummary := fmt.Sprintf("Desenvolvimento e entrega profissional de software para o módulo '%s'. Inclui arquitetura, implementação, testes de conformidade e deploy seguro.", cardItem.Title)
		deliverables := []string{
			"Arquitetura de software limpa (SOLID) com desacoplamento de banco e infraestrutura",
			"Implementação das funcionalidades com testes automatizados de regressão",
			"Segurança de dados e conformidade com RLS (Row Level Security) e LGPD",
			"Ambiente de prévia ao vivo (Sandbox) para homologação antes do deploy em produção",
			"Plano de rollback imediato e monitoramento pós-entrega",
		}

		timelineDays := 10
		switch cardItem.Type {
		case "bug":
			scopeSummary = fmt.Sprintf("Diagnóstico, correção emergencial de defeito e testes de validação para '%s'.", cardItem.Title)
			deliverables = []string{
				"Identificação de causa-raiz e mitigação do incidente",
				"Correção de código e testes de regressão automatizados",
				"Publicação segura em produção com monitoramento pós-deploy",
			}
			timelineDays = 3
		case "suporte":
			scopeSummary = fmt.Sprintf("Atendimento técnico e manutenção contínua para '%s'.", cardItem.Title)
			deliverables = []string{
				"Ajustes de configuração e otimização operacional",
				"Verificação de integridade de dados e auditoria de acessos",
				"Suporte direto ao dono da aplicação",
			}
			timelineDays = 2
		case "orcamento":
			timelineDays = 12
		}

		acceptURL := fmt.Sprintf("%s/aceite", uc.baseURL)

		// 3. Tenta enriquecer com especificações BDD se existirem
		if spec, err := tx.Specifications.GetByCardID(ctx, cardID); err == nil && spec != nil {
			if len(spec.UserStories) > 0 {
				var stories []string
				for _, us := range spec.UserStories {
					if title, ok := us["title"].(string); ok && title != "" {
						stories = append(stories, title)
					}
				}
				if len(stories) > 0 {
					deliverables = append([]string{fmt.Sprintf("User Stories BDD validadas (%d módulos mapeados)", len(stories))}, deliverables...)
				}
			}
			if spec.ScopeURL != "" {
				acceptURL = spec.ScopeURL
			}
		}

		// 4. Busca informações do PIX se existirem
		var pixEMV, pixTxID string
		if paymentOrder, err := tx.PaymentOrders.GetByCardID(ctx, cardID); err == nil && paymentOrder != nil {
			pixEMV = paymentOrder.PixEMV
			pixTxID = paymentOrder.PixTxID
		}

		// 5. Condições comerciais
		priceCents := cardItem.PriceCents
		currency := cardItem.Currency
		if currency == "" {
			currency = "BRL"
		}

		formattedPrice := fmt.Sprintf("%.2f", float64(priceCents)/100.0)
		formattedPriceBR := fmt.Sprintf("R$ %s", formattedPrice)
		if priceCents == 0 {
			formattedPriceBR = "A Definir / Sob Demanda"
		}

		paymentTerms := "Sinal de 50% via PIX no início do desenvolvimento e 50% na aprovação final da entrega em ambiente de prévia."
		if cardItem.Type == "bug" || cardItem.Type == "suporte" {
			paymentTerms = "Pagamento único de 100% via PIX após validação da solução em prévia."
		}

		warrantyDays := 90
		terms := "1. A Geevheit Software House garante a correção sem custos adicionais de qualquer inconformidade técnica ou bug pelo prazo improrrogável de 90 dias após a entrega.\n2. Alterações de escopo não contempladas nesta proposta serão tratadas como nova demanda com orçamento prévio.\n3. O início dos trabalhos ocorre após a confirmação do sinal financeiro e aceite formal da proposta."

		// 6. Mensagem formatada para WhatsApp
		deliverablesListText := ""
		for _, d := range deliverables {
			deliverablesListText += fmt.Sprintf("  • %s\n", d)
		}

		whatsAppMsg := fmt.Sprintf(`*PROPOSTA COMERCIAL — GEEVHEIT SOFTWARE HOUSE* 🚀
Olá, *%s*! Tudo bem?
Segue a proposta técnica e comercial referente ao projeto:

📋 *Projeto:* %s (%s)
⏱️ *Prazo de Execução:* %d dias úteis
🛡️ *Garantia Técnica:* %d dias com suporte incluso

📦 *Principais Entregas:*
%s
💰 *Investimento:* %s
💳 *Condição:* %s`,
			cardItem.ClientName,
			cardItem.Title,
			cardItem.AppName,
			timelineDays,
			warrantyDays,
			deliverablesListText,
			formattedPriceBR,
			paymentTerms,
		)

		if pixEMV != "" {
			whatsAppMsg += fmt.Sprintf("\n\n🔑 *Chave PIX Copia e Cola (Sinal):*\n```%s```", pixEMV)
		}

		whatsAppMsg += fmt.Sprintf(`

✍️ *Aprovação Formal:*
Para aprovar o projeto e acompanhar o status:
%s

Atenciosamente,
*Geevheit Software House*`, acceptURL)

		docID := uuid.New().String()
		doc = &ProposalDocumentDTO{
			ID:                 docID,
			CardID:             cardID,
			Title:              cardItem.Title,
			ClientName:         cardItem.ClientName,
			AppName:            cardItem.AppName,
			ScopeSummary:       scopeSummary,
			Deliverables:       deliverables,
			TimelineDays:       timelineDays,
			WarrantyDays:       warrantyDays,
			PriceCents:         priceCents,
			Currency:           currency,
			PaymentTerms:       paymentTerms,
			AcceptURL:          acceptURL,
			PixEMV:             pixEMV,
			PixTxID:            pixTxID,
			FormattedWhatsApp:  whatsAppMsg,
			TermsAndConditions: terms,
			GeneratedAt:        now,
		}

		// 7. Registra evento no cartão
		if err := tx.CardEvents.Create(ctx, CardEventDTO{
			ID:        uuid.New().String(),
			CardID:    cardID,
			Kind:      "proposal_document_generated",
			ActorType: "owner",
			Payload: map[string]any{
				"proposal_doc_id": docID,
				"price_cents":     priceCents,
				"timeline_days":   timelineDays,
			},
			At: now,
		}); err != nil {
			return fmt.Errorf("erro ao registrar evento: %w", err)
		}

		// 8. Registra no audit log
		return tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.New().String(),
			OwnerID:    principal.ID,
			Action:     "proposal.document_generated",
			ActorType:  "owner",
			TargetType: "card",
			TargetID:   &cardID,
			Meta: map[string]any{
				"price_cents":   priceCents,
				"timeline_days": timelineDays,
			},
			At: now,
		})
	})

	if err != nil {
		return nil, err
	}

	_ = uc.publisher.Publish(ctx, TopicCardsChanged, map[string]string{"card_id": cardID})

	return doc, nil
}

