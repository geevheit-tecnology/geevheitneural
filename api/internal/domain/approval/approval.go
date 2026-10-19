package approval

import (
	"errors"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
)

var (
	ErrNotAGate = errors.New("o estágio informado não é um portão do dono")
)

// Approval registra a decisão explícita e imutável do dono em um portão.
type Approval struct {
	ID        string
	OwnerID   string
	CardID    string
	Gate      card.Stage
	Decision  card.Decision
	DecidedBy string
	DecidedAt time.Time
	Note      string
}

// New cria um registro de aprovação validando que o estágio é de fato um portão.
func New(id, ownerID, cardID string, gate card.Stage, dec card.Decision, decidedBy string, note string, at time.Time) (*Approval, error) {
	if !gate.IsGate() {
		return nil, ErrNotAGate
	}
	if !dec.IsValid() {
		return nil, errors.New("decisão inválida")
	}

	return &Approval{
		ID:        id,
		OwnerID:   ownerID,
		CardID:    cardID,
		Gate:      gate,
		Decision:  dec,
		DecidedBy: decidedBy,
		DecidedAt: at,
		Note:      note,
	}, nil
}
