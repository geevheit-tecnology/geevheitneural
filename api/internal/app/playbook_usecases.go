package app

import (
	"context"
)

type ListPlaybooksUseCase struct {
	uow UnitOfWork
}

func NewListPlaybooksUseCase(uow UnitOfWork) *ListPlaybooksUseCase {
	return &ListPlaybooksUseCase{uow: uow}
}

func (uc *ListPlaybooksUseCase) Execute(ctx context.Context) ([]PlaybookDTO, error) {
	var list []PlaybookDTO
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		items, err := tx.Playbooks.ListPlaybooks(ctx)
		if err != nil {
			return err
		}
		list = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

type ListModulesUseCase struct {
	uow UnitOfWork
}

func NewListModulesUseCase(uow UnitOfWork) *ListModulesUseCase {
	return &ListModulesUseCase{uow: uow}
}

func (uc *ListModulesUseCase) Execute(ctx context.Context) ([]ModuleDTO, error) {
	var list []ModuleDTO
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		items, err := tx.Playbooks.ListModules(ctx)
		if err != nil {
			return err
		}
		list = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

type CreatePlaybookUseCase struct {
	uow UnitOfWork
}

func NewCreatePlaybookUseCase(uow UnitOfWork) *CreatePlaybookUseCase {
	return &CreatePlaybookUseCase{uow: uow}
}

func (uc *CreatePlaybookUseCase) Execute(ctx context.Context, p PlaybookDTO) (*PlaybookDTO, error) {
	if p.AutonomyLevel < 1 || p.AutonomyLevel > 3 {
		p.AutonomyLevel = 2
	}
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		return tx.Playbooks.UpsertPlaybook(ctx, &p)
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type CreateModuleUseCase struct {
	uow UnitOfWork
}

func NewCreateModuleUseCase(uow UnitOfWork) *CreateModuleUseCase {
	return &CreateModuleUseCase{uow: uow}
}

func (uc *CreateModuleUseCase) Execute(ctx context.Context, m ModuleDTO) (*ModuleDTO, error) {
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		return tx.Playbooks.UpsertModule(ctx, &m)
	})
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// SeedPlaybooksAndModules inicializa os playbooks e módulos reutilizáveis da fábrica autônoma
func SeedPlaybooksAndModules(ctx context.Context, uow UnitOfWork) error {
	return uow.RunInTx(ctx, func(tx Repositories) error {
		// Playbooks padrão
		playbooks := []PlaybookDTO{
			{
				ProductType:   "saas_b2b",
				Title:         "SaaS B2B Multi-tenant",
				Description:   "Playbook para sistemas empresariais web com isolamento de dados, painel administrativo e cobrança recorrente.",
				Stack: map[string]any{
					"backend":  "Go 1.22 + Clean Architecture + PostgreSQL",
					"frontend": "Next.js 14 App Router + TailwindCSS",
					"auth":     "Supabase Auth + TOTP 2FA",
					"billing":  "Stripe / Asaas Assinaturas",
				},
				Checklist: []string{
					"Configuração de RLS no PostgreSQL",
					"Isolamento multi-tenant por client_id",
					"Webhooks de faturamento idempotentes",
					"Exportação de dados para LGPD",
				},
				AutonomyLevel: 2,
			},
			{
				ProductType:   "mobile_app",
				Title:         "App Mobile Flutter / PWA",
				Description:   "Playbook para aplicações móveis conectadas via API REST e WebSocket para atualizações em tempo real.",
				Stack: map[string]any{
					"backend":  "Go 1.22 REST API",
					"mobile":   "Next.js PWA Mobile-First / Flutter",
					"realtime": "SSE (Server-Sent Events)",
				},
				Checklist: []string{
					"Layout responsivo calibrado para 360px",
					"Feedback táctil e micro-animações",
					"Modo offline-friendly com cache local",
				},
				AutonomyLevel: 2,
			},
			{
				ProductType:   "landing_page",
				Title:         "Landing Page de Alta Conversão",
				Description:   "Página única com formulário intake, SEO técnico e captura de leads com proteção antispam.",
				Stack: map[string]any{
					"frontend": "Next.js Static HTML + Tailwind CSS",
					"hosting":  "Vercel / Cloudflare Pages",
				},
				Checklist: []string{
					"Honeypot e rate limit no endpoint de captura",
					"Metatags OpenGraph e JSON-LD",
					"Performance Lighthouse > 95",
				},
				AutonomyLevel: 3,
			},
		}

		for _, p := range playbooks {
			_ = tx.Playbooks.UpsertPlaybook(ctx, &p)
		}

		// Módulos reutilizáveis catalogados
		modules := []ModuleDTO{
			{
				Name:        "auth_supabase_totp",
				Category:    "security",
				Description: "Autenticação segura com JWT e TOTP 2FA obrigatório (aal2).",
				CodeTemplates: map[string]any{
					"middleware": "internal/adapters/http/middleware/auth.go",
				},
				IsReviewed: true,
			},
			{
				Name:        "payment_webhook_idempotent",
				Category:    "billing",
				Description: "Recepção de webhooks de pagamento com deduplicação por hash e atualização de status em transação.",
				CodeTemplates: map[string]any{
					"handler": "internal/adapters/http/handlers/webhooks.go",
				},
				IsReviewed: true,
			},
			{
				Name:        "audit_trail_immutable",
				Category:    "compliance",
				Description: "Trilha de auditoria append-only protegida por trigger SQL contra modificações.",
				CodeTemplates: map[string]any{
					"sql": "db/migrations/00001_initial_schema.sql",
				},
				IsReviewed: true,
			},
			{
				Name:        "kanban_bottom_sheet_mobile",
				Category:    "ui",
				Description: "Componente de gaveta deslizante mobile-first com animação suave e suporte a temas.",
				CodeTemplates: map[string]any{
					"react": "src/features/kanban/components/CardBottomSheet.tsx",
				},
				IsReviewed: true,
			},
		}

		for _, m := range modules {
			_ = tx.Playbooks.UpsertModule(ctx, &m)
		}

		return nil
	})
}
