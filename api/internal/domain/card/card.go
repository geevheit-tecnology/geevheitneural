package card

import (
	"errors"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/money"
)

var (
	ErrInvalidStage       = errors.New("estágio de cartão inválido")
	ErrInvalidTransition  = errors.New("transição de estágio inválida")
	ErrGateRequiresOwner  = errors.New("portões de aprovação exigem decisão do dono")
	ErrSystemPaused       = errors.New("sistema pausado: nenhuma execução pode ser iniciada")
	ErrCardRecusado       = errors.New("cartão está recusado")
	ErrInvalidCardType    = errors.New("tipo de cartão inválido")
	ErrEmptyTitle         = errors.New("título do cartão não pode ser vazio")
	ErrInvalidPrice       = errors.New("preço cobrado não pode ser negativo")
)

type Type string

const (
	TypeOrcamento  Type = "orcamento"
	TypeBug        Type = "bug"
	TypeSuporte    Type = "suporte"
	TypeFinanceiro Type = "financeiro"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeOrcamento, TypeBug, TypeSuporte, TypeFinanceiro:
		return true
	default:
		return false
	}
}

type Stage string

const (
	StageEntrada           Stage = "entrada"
	StagePropostaAprovar   Stage = "proposta_aprovar"   // Portão 1
	StageAguardandoCliente Stage = "aguardando_cliente"
	StagePromptAprovar     Stage = "prompt_aprovar"     // Portão 2
	StageEmExecucao        Stage = "em_execucao"
	StageEntregaAprovar    Stage = "entrega_aprovar"    // Portão 3
	StageEntregue          Stage = "entregue"
	StageRecusado          Stage = "recusado"
)

func (s Stage) IsValid() bool {
	switch s {
	case StageEntrada, StagePropostaAprovar, StageAguardandoCliente,
		StagePromptAprovar, StageEmExecucao, StageEntregaAprovar,
		StageEntregue, StageRecusado:
		return true
	default:
		return false
	}
}

// IsGate indica se o estágio é um portão que exige decisão explícita do dono.
func (s Stage) IsGate() bool {
	return s == StagePropostaAprovar || s == StagePromptAprovar || s == StageEntregaAprovar
}

// TouchesExecution verifica se o estágio representa ou libera execução ativa.
func (s Stage) TouchesExecution() bool {
	return s == StageEmExecucao
}

type Decision string

const (
	DecisionAprovado Decision = "aprovado"
	DecisionMaisInfo Decision = "mais_info"
	DecisionRecusado Decision = "recusado"
)

func (d Decision) IsValid() bool {
	switch d {
	case DecisionAprovado, DecisionMaisInfo, DecisionRecusado:
		return true
	default:
		return false
	}
}

// Card representa a entidade central do fluxo operacional.
type Card struct {
	ID           string
	OwnerID      string
	ClientID     string
	AppID        *string
	Type         Type
	Title        string
	Price        money.Money
	EstCost      money.Money
	Stage        Stage
	Priority     int16
	DueDate      *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// New cria um novo cartão validando invariantes iniciais.
func New(id, ownerID, clientID string, cardType Type, title string, price, estCost money.Money, now time.Time) (*Card, error) {
	if !cardType.IsValid() {
		return nil, ErrInvalidCardType
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if price.Cents < 0 {
		return nil, ErrInvalidPrice
	}

	return &Card{
		ID:        id,
		OwnerID:   ownerID,
		ClientID:  clientID,
		Type:      cardType,
		Title:     title,
		Price:     price,
		EstCost:   estCost,
		Stage:     StageEntrada,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// DecideGate avalia e aplica a decisão do dono em um estágio de portão.
func (c *Card) DecideGate(dec Decision, systemPaused bool, now time.Time) (Stage, error) {
	if !c.Stage.IsGate() {
		return c.Stage, ErrInvalidTransition
	}
	if !dec.IsValid() {
		return c.Stage, errors.New("decisão inválida")
	}

	if dec == DecisionRecusado {
		c.Stage = StageRecusado
		c.UpdatedAt = now
		return c.Stage, nil
	}

	switch c.Stage {
	case StagePropostaAprovar:
		if dec == DecisionAprovado {
			c.Stage = StageAguardandoCliente
		}
		// Se mais_info, mantém StagePropostaAprovar

	case StagePromptAprovar:
		if dec == DecisionAprovado {
			if systemPaused {
				return c.Stage, ErrSystemPaused
			}
			c.Stage = StageEmExecucao
		}
		// Se mais_info, mantém StagePromptAprovar

	case StageEntregaAprovar:
		if dec == DecisionAprovado {
			c.Stage = StageEntregue
		} else if dec == DecisionMaisInfo {
			if systemPaused {
				return c.Stage, ErrSystemPaused
			}
			// Pedir mais informações na entrega volta para em_execucao
			c.Stage = StageEmExecucao
		}
	}

	c.UpdatedAt = now
	return c.Stage, nil
}

// AdvanceManual avança estágios que não são portões.
func (c *Card) AdvanceManual(systemPaused bool, now time.Time) (Stage, error) {
	if c.Stage.IsGate() {
		return c.Stage, ErrGateRequiresOwner
	}

	switch c.Stage {
	case StageEntrada:
		c.Stage = StagePropostaAprovar

	case StageAguardandoCliente:
		c.Stage = StagePromptAprovar

	case StageEmExecucao:
		c.Stage = StageEntregaAprovar

	case StageRecusado:
		return c.Stage, ErrCardRecusado

	case StageEntregue:
		return c.Stage, errors.New("cartão já finalizado")

	default:
		return c.Stage, ErrInvalidTransition
	}

	c.UpdatedAt = now
	return c.Stage, nil
}

// RestoreRecusado restaura um cartão recusado de volta para Entrada.
func (c *Card) RestoreRecusado(now time.Time) error {
	if c.Stage != StageRecusado {
		return errors.New("apenas cartões recusados podem ser restaurados")
	}
	c.Stage = StageEntrada
	c.UpdatedAt = now
	return nil
}
