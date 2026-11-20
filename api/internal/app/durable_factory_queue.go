package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

var ErrDurableJobPending = errors.New("tarefa ainda está em andamento ou aguardando reconciliação")
var ErrDurableJobUnavailable = errors.New("tarefa não encontrada ou indisponível")

// DurableFactoryProcessor só pode indicar sucesso depois da persistência do resultado.
type DurableFactoryProcessor interface {
	Process(context.Context, system.OwnerPrincipal, DurableFactoryJob) error
}

type DurableFactoryQueue struct {
	uow       UnitOfWork
	processor DurableFactoryProcessor
}

func NewDurableFactoryQueue(uow UnitOfWork, processor DurableFactoryProcessor) *DurableFactoryQueue {
	return &DurableFactoryQueue{uow: uow, processor: processor}
}

func requireDurableOwner(owner system.OwnerPrincipal) error {
	if owner.ID == "" || owner.AAL != "aal2" {
		return system.ErrOwnerRequired
	}
	return nil
}

// Enqueue só aceita cartão que já passou pelos portões do dono.
func (q *DurableFactoryQueue) Enqueue(ctx context.Context, owner system.OwnerPrincipal, cardID, requestKey string) (*DurableFactoryJob, error) {
	if err := requireDurableOwner(owner); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(cardID); err != nil {
		return nil, fmt.Errorf("cartão inválido")
	}
	if _, err := uuid.Parse(requestKey); err != nil {
		return nil, fmt.Errorf("chave de solicitação inválida")
	}
	var job *DurableFactoryJob
	err := q.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.SystemState == nil || tx.Cards == nil || tx.DurableFactory == nil {
			return fmt.Errorf("repositórios da fábrica indisponíveis")
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
		c, err := tx.Cards.GetByID(ctx, cardID)
		if err != nil {
			return err
		}
		if c == nil || c.Stage != card.StageEmExecucao {
			return fmt.Errorf("cartão precisa estar autorizado para execução")
		}
		job, err = tx.DurableFactory.Queue(ctx, cardID, requestKey)
		return err
	})
	return job, err
}

// MarkInterrupted identifica execuções expiradas, sem reexecutar agentes.
// O agendamento e a revisão pelo dono são responsabilidades separadas.
func (q *DurableFactoryQueue) MarkInterrupted(ctx context.Context, owner system.OwnerPrincipal) (int64, error) {
	if err := requireDurableOwner(owner); err != nil {
		return 0, err
	}
	if q.uow == nil {
		return 0, fmt.Errorf("repositório da fábrica indisponível")
	}
	var count int64
	err := q.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.DurableFactory == nil {
			return fmt.Errorf("repositório da fábrica indisponível")
		}
		var err error
		count, err = tx.DurableFactory.MarkExpired(ctx)
		return err
	})
	return count, err
}

// RunNext processa uma tarefa em uma invocação autenticada; não cria goroutine.
// Uma invocação cancelada deixa o lease expirar e exige recuperação supervisionada.
func (q *DurableFactoryQueue) RunNext(ctx context.Context, owner system.OwnerPrincipal) (*DurableFactoryJob, error) {
	return q.run(ctx, owner, "")
}

// RunJob só assume a tarefa identificada; entregas duplicadas não executam outra tarefa.
func (q *DurableFactoryQueue) RunJob(ctx context.Context, owner system.OwnerPrincipal, jobID string) (*DurableFactoryJob, error) {
	if _, err := uuid.Parse(jobID); err != nil {
		return nil, fmt.Errorf("tarefa inválida")
	}
	return q.run(ctx, owner, jobID)
}

func (q *DurableFactoryQueue) run(ctx context.Context, owner system.OwnerPrincipal, jobID string) (*DurableFactoryJob, error) {
	if err := requireDurableOwner(owner); err != nil {
		return nil, err
	}
	if q.processor == nil {
		return nil, fmt.Errorf("processador durável indisponível")
	}
	var job *DurableFactoryJob
	err := q.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.SystemState == nil || tx.DurableFactory == nil {
			return fmt.Errorf("repositórios da fábrica indisponíveis")
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
		if jobID == "" {
			job, err = tx.DurableFactory.Claim(ctx, uuid.NewString())
		} else {
			job, err = tx.DurableFactory.ClaimByID(ctx, jobID, uuid.NewString())
			if err != nil || job != nil {
				return err
			}
			existing, lookupErr := tx.DurableFactory.GetByID(ctx, jobID)
			if lookupErr != nil {
				return lookupErr
			}
			if existing == nil {
				return ErrDurableJobUnavailable
			}
			if existing.Status == "running" {
				if _, err := tx.DurableFactory.MarkExpiredForCard(ctx, existing.CardID); err != nil {
					return err
				}
				existing, lookupErr = tx.DurableFactory.GetByID(ctx, jobID)
				if lookupErr != nil {
					return lookupErr
				}
				if existing == nil {
					return ErrDurableJobUnavailable
				}
			}
			switch existing.Status {
			case "completed", "failed", "interrupted":
				return nil
			default:
				return ErrDurableJobPending
			}
		}
		return err
	})
	if err != nil || job == nil {
		return job, err
	}
	processErr := q.processor.Process(ctx, owner, *job)
	if ctx.Err() != nil {
		return job, fmt.Errorf("execução interrompida; confirmação pendente: %w", ctx.Err())
	}
	if errors.Is(processErr, context.DeadlineExceeded) || errors.Is(processErr, context.Canceled) {
		return job, fmt.Errorf("execução interrompida; confirmação pendente: %w", processErr)
	}
	status := "completed"
	if processErr != nil {
		status = "failed"
	}
	err = q.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.DurableFactory == nil {
			return fmt.Errorf("repositório da fábrica indisponível")
		}
		ok, err := tx.DurableFactory.Finish(ctx, job.ID, job.LeaseToken, status)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("lease expirou ou foi substituído; resultado não confirmado")
		}
		// A falha do worker pode acontecer antes de o pipeline atualizar seu estado.
		// Confirma ambos na mesma transação para que o painel não continue em execução.
		if processErr != nil {
			if err := tx.DurableFactory.MarkActivePipelineFailed(ctx, job.CardID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return job, err
	}
	job.Status = status
	if processErr != nil {
		return job, fmt.Errorf("falha no processamento da fábrica: %w", processErr)
	}
	return job, nil
}
