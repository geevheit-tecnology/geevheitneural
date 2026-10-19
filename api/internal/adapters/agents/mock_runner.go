package agents

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
)

type MockAgentRunner struct{}

func NewMockAgentRunner() *MockAgentRunner {
	return &MockAgentRunner{}
}

func (r *MockAgentRunner) RunStep(ctx context.Context, role pipeline.Role, prompt string, contextMeta map[string]any) (*pipeline.RunnerResult, error) {
	switch role {
	case pipeline.RoleRequisitos:
		next := pipeline.RoleArquiteto
		return &pipeline.RunnerResult{
			Role:      role,
			CostCents: 45, // R$ 0,45 de tokens
			Output:    "User stories e critérios de aceite BDD (Given-When-Then) gerados.",
			Artifacts: map[string]any{
				"stories": []map[string]any{
					{
						"title":  "Autenticação Segura com TOTP",
						"given":  "Um usuário cadastrado no sistema",
						"when":   "Insere credenciais válidas e código 2FA",
						"then":   "Recebe token com privilégio AAL2",
						"passed": true,
					},
					{
						"title":  "Operação sob Sistema Pausado",
						"given":  "O sistema está marcado como pausado",
						"when":   "Qualquer agente tenta executar escrita",
						"then":   "A requisição é bloqueada com status 423 Locked",
						"passed": true,
					},
				},
				"test_code": `func TestSpecification_BDD(t *testing.T) { /* BDD tests auto-generated */ }`,
			},
			NextRole:     &next,
			RequiresGate: false,
		}, nil

	case pipeline.RoleArquiteto:
		next := pipeline.RoleDev
		return &pipeline.RunnerResult{
			Role:      role,
			CostCents: 80,
			Output:    "Arquitetura definida: Monorepo Go + Next.js PWA, RLS no Postgres e isolamento por tenant.",
			Artifacts: map[string]any{
				"architecture_notes": "Clean architecture, adapters e domain isolado sem I/O.",
				"playbook_selected":  "pwa_mobile",
			},
			NextRole:     &next,
			RequiresGate: false,
		}, nil

	case pipeline.RoleDev:
		next := pipeline.RoleQA
		return &pipeline.RunnerResult{
			Role:      role,
			CostCents: 150,
			Output:    "Código dos componentes e handlers implementado com passagem de todos os testes BDD.",
			Artifacts: map[string]any{
				"commits_created": 3,
				"pr_url":          "https://github.com/geevheit/neuralcenter/pull/42",
			},
			NextRole:     &next,
			RequiresGate: false,
		}, nil

	case pipeline.RoleQA:
		next := pipeline.RoleSeguranca
		return &pipeline.RunnerResult{
			Role:      role,
			CostCents: 60,
			Output:    "Suíte de testes executada com 100% de sucesso. Cobertura de 94.2%.",
			Artifacts: map[string]any{
				"tests_run": 24,
				"passed":    24,
				"coverage":  "94.2%",
			},
			NextRole:     &next,
			RequiresGate: false,
		}, nil

	case pipeline.RoleSeguranca:
		next := pipeline.RolePublicador
		return &pipeline.RunnerResult{
			Role:      role,
			CostCents: 90,
			Output:    "Varredura de vulnerabilidades SAST concluída. Zero segredos expostos, conformidade LGPD aprovada.",
			Artifacts: map[string]any{
				"secrets_found":  0,
				"outdated_deps":  0,
				"lgpd_compliant": true,
			},
			NextRole:     &next,
			RequiresGate: false,
		}, nil

	case pipeline.RolePublicador:
		next := pipeline.RoleSuporte
		return &pipeline.RunnerResult{
			Role:      role,
			CostCents: 40,
			Output:    "Prévia ao vivo gerada com sucesso na URL temporária. Cartão submetido ao Portão de Entrega.",
			Artifacts: map[string]any{
				"preview_url": fmt.Sprintf("https://preview.geevheit.internal/preview-%s", contextMeta["card_id"]),
				"rollback_plan": map[string]any{
					"strategy": "instant_dns_swap",
					"snapshot": "v1.0.0-prev",
				},
			},
			NextRole:     &next,
			RequiresGate: true, // Portão 3: exige aprovação do Dono antes do deploy final!
		}, nil

	case pipeline.RoleSuporte:
		return &pipeline.RunnerResult{
			Role:      role,
			CostCents: 20,
			Output:    "Monitoramento ativo iniciado. Zero erros nas últimas 24h.",
			Artifacts: map[string]any{
				"health_status": "optimal",
				"uptime":        "100%",
			},
			NextRole:     nil,
			RequiresGate: false,
		}, nil

	default:
		return nil, fmt.Errorf("papel de agente desconhecido: %s", role)
	}
}
