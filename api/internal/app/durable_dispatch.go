package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

var ErrDurableDispatchPending = errors.New("tarefa salva, mas o sinal para execução não foi confirmado")

type DurableQueueSignal interface {
	Publish(context.Context, string, string) error
}
type DurableBuildStarter interface {
	DurableQueueEnabled() bool
	Execute(context.Context, system.OwnerPrincipal, StartBuildCommand) (*PipelineDTO, error)
}

type DurableBuildDispatcher struct {
	starter DurableBuildStarter
	signal  DurableQueueSignal
}

func NewDurableBuildDispatcher(starter DurableBuildStarter, signal DurableQueueSignal) *DurableBuildDispatcher {
	return &DurableBuildDispatcher{starter: starter, signal: signal}
}

// Dispatch grava primeiro e sinaliza depois. Em falha o chamador deve repetir
// com a mesma chave: não se afirma que a tarefa foi iniciada.
func (d *DurableBuildDispatcher) Dispatch(ctx context.Context, owner system.OwnerPrincipal, cmd StartBuildCommand, oidc string) (*PipelineDTO, error) {
	if err := requireDurableOwner(owner); err != nil {
		return nil, err
	}
	if d.starter == nil || d.signal == nil || !d.starter.DurableQueueEnabled() {
		return nil, fmt.Errorf("fila durável indisponível")
	}
	if strings.TrimSpace(oidc) == "" {
		return nil, fmt.Errorf("identidade da Vercel indisponível")
	}
	result, err := d.starter.Execute(ctx, owner, cmd)
	if err != nil {
		return nil, err
	}
	if result == nil || result.JobID == "" {
		return result, fmt.Errorf("tarefa durável não retornada")
	}
	if err := d.signal.Publish(ctx, oidc, result.JobID); err != nil {
		return result, fmt.Errorf("%w: %v", ErrDurableDispatchPending, err)
	}
	return result, nil
}
