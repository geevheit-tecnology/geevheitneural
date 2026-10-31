package command_test

import (
	"errors"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/domain/command"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input          string
		expectedIntent command.Intent
		expectedTarget string
		destructive    bool
		expectErr      bool
	}{
		{
			input:          "quanto gastei esse mês",
			expectedIntent: command.IntentFinances,
			destructive:    false,
		},
		{
			input:          "Lucro",
			expectedIntent: command.IntentFinances,
			destructive:    false,
		},
		{
			input:          "qual o status",
			expectedIntent: command.IntentStatus,
			destructive:    false,
		},
		{
			input:          "o que tem pra hoje",
			expectedIntent: command.IntentStatus,
			destructive:    false,
		},
		{
			input:          "aprovar Casa Verde",
			expectedIntent: command.IntentApprove,
			expectedTarget: "casa verde",
			destructive:    false,
		},
		{
			input:          "Aprovar o orçamento da Clínica Sorriso",
			expectedIntent: command.IntentApprove,
			expectedTarget: "orcamento da clinica sorriso",
			destructive:    false,
		},
		{
			input:          "construir Casa Verde",
			expectedIntent: command.IntentBuild,
			expectedTarget: "casa verde",
			destructive:    false,
		},
		{
			input:          "Jarvis, construir aplicativo de transporte",
			expectedIntent: command.IntentBuild,
			expectedTarget: "aplicativo de transporte",
			destructive:    false,
		},
		{
			input:          "criar projeto Plataforma E-commerce",
			expectedIntent: command.IntentCreate,
			expectedTarget: "plataforma e-commerce",
			destructive:    false,
		},
		{
			input:          "Jake, criar demanda Sistema de Gestão para Clínica Alpha",
			expectedIntent: command.IntentCreate,
			expectedTarget: "sistema de gestao para clinica alpha",
			destructive:    false,
		},
		{
			input:          "recusar Casa Verde",
			expectedIntent: command.IntentReject,
			expectedTarget: "casa verde",
			destructive:    true,
		},
		{
			input:          "rejeitar pedido da Padaria Central",
			expectedIntent: command.IntentReject,
			expectedTarget: "pedido da padaria central",
			destructive:    true,
		},
		{
			input:          "parar sistema",
			expectedIntent: command.IntentPauseSystem,
			destructive:    true,
		},
		{
			input:          "retomar",
			expectedIntent: command.IntentResumeSystem,
			destructive:    false,
		},
		{
			input:     "   ",
			expectErr: true,
		},
		{
			input:          "pode fazer uma landpage de agencia de veiculos simples",
			expectedIntent: command.IntentCreate,
			expectedTarget: "uma landpage de agencia de veiculos simples",
		},
		{
			input:          "Jarvis, por favor, criar projeto Loja de Carros",
			expectedIntent: command.IntentCreate,
			expectedTarget: "loja de carros",
		},
		{
			input:          "você poderia construir site da Auto Center",
			expectedIntent: command.IntentBuild,
			expectedTarget: "site da auto center",
		},
		{
			input:          "podemos conversar",
			expectedIntent: command.IntentUnknown,
		},
		{
			input:          "abobrinha aleatoria",
			expectedIntent: command.IntentUnknown,
			destructive:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res, err := command.ParseCommand(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("esperava erro, mas obteve nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if res.Intent != tt.expectedIntent {
				t.Errorf("intent esperada %s, obteve %s", tt.expectedIntent, res.Intent)
			}
			if tt.expectedTarget != "" && res.Target != tt.expectedTarget {
				t.Errorf("alvo esperado %q, obteve %q", tt.expectedTarget, res.Target)
			}
			if res.Destructive != tt.destructive {
				t.Errorf("destructive esperado %v, obteve %v", tt.destructive, res.Destructive)
			}
		})
	}
}

func TestResolveTarget_Ambiguity(t *testing.T) {
	available := []command.TargetMatch{
		{CardID: "1", ClientName: "Clínica Sorriso", Title: "Site institucional"},
		{CardID: "2", ClientName: "Casa Verde", Title: "Painel de estoque"},
		{CardID: "3", ClientName: "Padaria Central", Title: "Erro no cardápio"},
		{CardID: "4", ClientName: "Padaria Central", Title: "Ajuste de logo"},
	}

	t.Run("Exact match succeeds", func(t *testing.T) {
		m, err := command.ResolveTarget("Casa Verde", available)
		if err != nil {
			t.Fatalf("esperava sucesso, obteve: %v", err)
		}
		if m.CardID != "2" {
			t.Errorf("card id esperado 2, obteve %s", m.CardID)
		}
	})

	t.Run("Accent-insensitive match succeeds", func(t *testing.T) {
		m, err := command.ResolveTarget("clinica sorriso", available)
		if err != nil {
			t.Fatalf("esperava sucesso, obteve: %v", err)
		}
		if m.CardID != "1" {
			t.Errorf("card id esperado 1, obteve %s", m.CardID)
		}
	})

	t.Run("Ambiguous target returns ErrAmbiguousTarget", func(t *testing.T) {
		_, err := command.ResolveTarget("Padaria Central", available)
		if err == nil {
			t.Fatalf("esperava erro de ambiguidade, obteve nil")
		}
		if !errors.Is(err, command.ErrAmbiguousTarget) {
			t.Errorf("esperava ErrAmbiguousTarget, obteve: %v", err)
		}
	})

	t.Run("Non-existent target returns ErrTargetNotFound", func(t *testing.T) {
		_, err := command.ResolveTarget("Empresa Inexistente", available)
		if err == nil {
			t.Fatalf("esperava erro de não encontrado, obteve nil")
		}
		if !errors.Is(err, command.ErrTargetNotFound) {
			t.Errorf("esperava ErrTargetNotFound, obteve: %v", err)
		}
	})
}
