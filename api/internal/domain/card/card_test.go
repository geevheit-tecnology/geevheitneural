package card_test

import (
	"errors"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
)

func TestCard_GatesAndTransitions(t *testing.T) {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)

	t.Run("Gates cannot advance via AdvanceManual", func(t *testing.T) {
		c, _ := card.New("1", "owner-1", "cli-1", card.TypeOrcamento, "Site", money.NewBRL(1000), money.NewBRL(100), now)
		c.Stage = card.StagePropostaAprovar

		_, err := c.AdvanceManual(false, now)
		if !errors.Is(err, card.ErrGateRequiresOwner) {
			t.Errorf("expected ErrGateRequiresOwner, got %v", err)
		}
	})

	t.Run("Proposta gate approval advances to aguardando_cliente", func(t *testing.T) {
		c, _ := card.New("1", "owner-1", "cli-1", card.TypeOrcamento, "Site", money.NewBRL(1000), money.NewBRL(100), now)
		c.Stage = card.StagePropostaAprovar

		next, err := c.DecideGate(card.DecisionAprovado, false, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next != card.StageAguardandoCliente || c.Stage != card.StageAguardandoCliente {
			t.Errorf("expected stage aguardando_cliente, got %s", next)
		}
	})

	t.Run("Prompt gate blocked if system is paused", func(t *testing.T) {
		c, _ := card.New("1", "owner-1", "cli-1", card.TypeOrcamento, "Site", money.NewBRL(1000), money.NewBRL(100), now)
		c.Stage = card.StagePromptAprovar

		_, err := c.DecideGate(card.DecisionAprovado, true, now)
		if !errors.Is(err, card.ErrSystemPaused) {
			t.Errorf("expected ErrSystemPaused, got %v", err)
		}
	})

	t.Run("Prompt gate approval when not paused advances to em_execucao", func(t *testing.T) {
		c, _ := card.New("1", "owner-1", "cli-1", card.TypeOrcamento, "Site", money.NewBRL(1000), money.NewBRL(100), now)
		c.Stage = card.StagePromptAprovar

		next, err := c.DecideGate(card.DecisionAprovado, false, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next != card.StageEmExecucao {
			t.Errorf("expected em_execucao, got %s", next)
		}
	})

	t.Run("Entrega gate mais_info sends back to em_execucao", func(t *testing.T) {
		c, _ := card.New("1", "owner-1", "cli-1", card.TypeOrcamento, "Site", money.NewBRL(1000), money.NewBRL(100), now)
		c.Stage = card.StageEntregaAprovar

		next, err := c.DecideGate(card.DecisionMaisInfo, false, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next != card.StageEmExecucao {
			t.Errorf("expected return to em_execucao, got %s", next)
		}
	})

	t.Run("Any gate recusado transitions to recusado", func(t *testing.T) {
		c, _ := card.New("1", "owner-1", "cli-1", card.TypeOrcamento, "Site", money.NewBRL(1000), money.NewBRL(100), now)
		c.Stage = card.StageEntregaAprovar

		next, err := c.DecideGate(card.DecisionRecusado, false, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next != card.StageRecusado {
			t.Errorf("expected recusado, got %s", next)
		}

		// Can be restored back to entrada
		if err := c.RestoreRecusado(now); err != nil {
			t.Fatalf("unexpected restore error: %v", err)
		}
		if c.Stage != card.StageEntrada {
			t.Errorf("expected restored to entrada, got %s", c.Stage)
		}
	})
}
