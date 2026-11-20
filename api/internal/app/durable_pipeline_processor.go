package app

import (
	"context"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

// leaseGuardedUOW mantém o lease bloqueado enquanto cada transação é aplicada.
// A verificação final impede confirmar gravações depois da expiração.
type leaseGuardedUOW struct {
	base UnitOfWork
	job  DurableFactoryJob
}

func (u leaseGuardedUOW) RunInTx(ctx context.Context, fn func(Repositories) error) error {
	return u.base.RunInTx(ctx, func(tx Repositories) error {
		check := func() error {
			if tx.DurableFactory == nil || tx.SystemState == nil {
				return fmt.Errorf("estado da execução indisponível")
			}
			valid, err := tx.DurableFactory.LockValidLease(ctx, u.job.ID, u.job.LeaseToken)
			if err != nil {
				return err
			}
			if !valid {
				return fmt.Errorf("lease da execução expirou ou foi substituído")
			}
			st, err := tx.SystemState.Get(ctx)
			if err != nil {
				return err
			}
			if st == nil {
				return fmt.Errorf("estado do sistema indisponível")
			}
			if st.Paused {
				return system.ErrSystemPaused
			}
			return nil
		}
		if err := check(); err != nil {
			return err
		}
		if err := fn(tx); err != nil {
			return err
		}
		return check()
	})
}

// DurablePipelineProcessor executa o pipeline dentro da invocação do worker.
// A entrega de mensagens e a recuperação supervisionada são configuradas separadamente.
type DurablePipelineProcessor struct{ build *StartBuildUseCase }

func NewDurablePipelineProcessor(build *StartBuildUseCase) *DurablePipelineProcessor {
	return &DurablePipelineProcessor{build: build}
}

func (p *DurablePipelineProcessor) Process(ctx context.Context, owner system.OwnerPrincipal, job DurableFactoryJob) error {
	if err := requireDurableOwner(owner); err != nil {
		return err
	}
	if p.build == nil || p.build.runner == nil || !p.build.persistGeneratedFiles {
		return fmt.Errorf("pipeline durável não configurado")
	}
	if job.ID == "" || job.CardID == "" || job.LeaseToken == "" {
		return fmt.Errorf("lease inválido")
	}
	deadline := job.LeaseExpiresAt.Add(-5 * time.Second)
	if !deadline.After(time.Now()) {
		return fmt.Errorf("lease sem prazo suficiente")
	}
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	if availability, ok := p.build.runner.(interface{ Ready(context.Context) error }); ok {
		if err := availability.Ready(ctx); err != nil {
			return err
		}
	}
	guarded := leaseGuardedUOW{base: p.build.uow, job: job}
	var title, clientID string
	var price, limit int64
	err := guarded.RunInTx(ctx, func(tx Repositories) error {
		if tx.Cards == nil || tx.Pipelines == nil {
			return fmt.Errorf("repositórios de cartão indisponíveis")
		}
		c, err := tx.Cards.GetByID(ctx, job.CardID)
		if err != nil {
			return err
		}
		if c == nil || c.Stage != card.StageEmExecucao {
			return fmt.Errorf("cartão não está autorizado para execução")
		}
		state, err := tx.Pipelines.GetByCardID(ctx, job.CardID)
		if err != nil {
			return err
		}
		if state == nil || state.SpendLimitCents <= 0 || state.SpendTotalCents != 0 {
			return fmt.Errorf("pipeline requer revisão antes de executar novamente")
		}
		title, clientID, price, limit = c.Title, c.ClientID, c.Price.Cents, state.SpendLimitCents
		state.Status = string(pipeline.StatusRunning)
		state.CurrentTask = "Execução iniciada pela fila"
		return tx.Pipelines.Upsert(ctx, state)
	})
	if err != nil {
		return err
	}
	build := &StartBuildUseCase{uow: guarded, events: p.build.events, guard: p.build.guard, runner: p.build.runner, sandbox: p.build.sandbox, publisher: p.build.publisher, persistGeneratedFiles: true}
	build.runFactoryPipeline(ctx, job.CardID, title, clientID, price, limit)
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("execução interrompida: %w", err)
	}
	return guarded.RunInTx(ctx, func(tx Repositories) error {
		state, err := tx.Pipelines.GetByCardID(ctx, job.CardID)
		if err != nil {
			return err
		}
		if state == nil || state.Status != string(pipeline.StatusCompleted) {
			return fmt.Errorf("pipeline não terminou com sucesso")
		}
		return nil
	})
}
