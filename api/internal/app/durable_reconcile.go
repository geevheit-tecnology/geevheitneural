package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

// DurableQueueReconciler reencontra tarefas salvas cujo sinal pode ter sido perdido.
// A chamada ainda exige o contexto autenticado do dono; não é um agendador.
type DurableQueueReconciler struct {
	uow    UnitOfWork
	signal DurableQueueSignal
}

func NewDurableQueueReconciler(uow UnitOfWork, signal DurableQueueSignal) *DurableQueueReconciler {
	return &DurableQueueReconciler{uow: uow, signal: signal}
}

// ReconcilePage reenvia sinais de até 100 tarefas pendentes, sem manter uma transação aberta
// durante chamadas à rede. Em erro, o chamador repete a mesma página; sinais são idempotentes.
// Uma varredura nova começa com cursor vazio e deve ser repetida periodicamente.
func (r *DurableQueueReconciler) ReconcilePage(ctx context.Context, owner system.OwnerPrincipal, oidc, cursor string, limit int32) (string, error) {
	if err := requireDurableOwner(owner); err != nil {
		return "", err
	}
	if r.uow == nil || r.signal == nil || strings.TrimSpace(oidc) == "" {
		return "", fmt.Errorf("reconciliação indisponível")
	}
	if limit < 1 || limit > 100 {
		return "", fmt.Errorf("limite de página inválido")
	}
	if cursor == "" {
		cursor = uuid.Nil.String()
	}
	id, err := uuid.Parse(cursor)
	if err != nil {
		return "", fmt.Errorf("cursor inválido")
	}
	var ids []string
	err = r.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.DurableFactory == nil {
			return fmt.Errorf("repositório da fábrica indisponível")
		}
		var err error
		ids, err = tx.DurableFactory.ListQueuedIDs(ctx, id.String(), limit)
		return err
	})
	if err != nil {
		return "", err
	}
	for _, jobID := range ids {
		if err := r.signal.Publish(ctx, oidc, jobID); err != nil {
			return "", fmt.Errorf("%w: sinalização de tarefa pendente falhou", ErrDurableDispatchPending)
		}
	}
	if len(ids) < int(limit) {
		return "", nil
	}
	return ids[len(ids)-1], nil
}
