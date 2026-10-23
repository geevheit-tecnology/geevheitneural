package postgres

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const ClaimsContextKey contextKey = "jwt_claims_raw"

// ContextWithClaims insere o JSON bruto dos claims JWT no contexto para aplicação de RLS por transação.
func ContextWithClaims(ctx context.Context, claimsJSON string) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, claimsJSON)
}

type UnitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{pool: pool}
}

// RunInTx executa uma função atômica com transação PostgreSQL aplicando papel e claims RLS (ADR-005).
func (u *UnitOfWork) RunInTx(ctx context.Context, fn func(tx app.Repositories) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação no postgres: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Aplica RLS se claims do JWT estiverem presentes no contexto
	if claimsRaw, ok := ctx.Value(ClaimsContextKey).(string); ok && claimsRaw != "" {
		_, err = tx.Exec(ctx, "SET LOCAL ROLE authenticated;")
		if err != nil {
			return fmt.Errorf("falha ao aplicar SET LOCAL ROLE authenticated: %w", err)
		}
		_, err = tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true);", claimsRaw)
		if err != nil {
			return fmt.Errorf("falha ao aplicar claims JWT da transação: %w", err)
		}
	}

	q := sqlcgen.New(tx)
	repos := app.Repositories{
		Cards:          NewCardRepository(q),
		Apps:           NewAppRepository(q),
		Finances:       NewFinanceRepository(q),
		Connectors:     NewConnectorRepository(q),
		Clients:        NewClientRepository(q),
		Proposals:      NewProposalRepository(q),
		Agents:         NewAgentRepository(q),
		Pipelines:      NewPipelineRepository(q),
		Specifications: NewSpecificationRepository(q),
		Playbooks:      NewPlaybookRepository(q),
		Previews:       NewPreviewRepository(q),
		FactoryMetrics: NewFactoryMetricsRepository(q),
		PaymentOrders:  NewPaymentOrderRepository(q),
		PortalTokens:   NewPortalTokenRepository(q),
		Subscriptions:  NewSubscriptionRepository(q),
		Webhooks:       NewWebhookRepository(q),
		Notifications:  NewNotificationRepository(q),
		Approvals:      NewApprovalRepository(q),
		CardEvents:     NewCardEventRepository(q),
		Audit:          NewAuditRepository(q),
		SystemState:    NewSystemStateRepository(q),
	}

	if err := fn(repos); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("falha ao commitar transação: %w", err)
	}

	return nil
}
