package command_test

import (
	"errors"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/command"
)

func TestConfirmationService(t *testing.T) {
	svc := command.NewConfirmationService(50 * time.Millisecond)

	t.Run("Issue and consume token successfully", func(t *testing.T) {
		act, err := svc.IssueToken(command.ActionTypeRejectCard, "card-123", "Confirmar recusa?", map[string]any{"reason": "escopo"})
		if err != nil {
			t.Fatalf("falha ao emitir token: %v", err)
		}
		if len(act.Token) != 32 {
			t.Errorf("tamanho do token inesperado: %d", len(act.Token))
		}

		consumed, err := svc.Consume(act.Token)
		if err != nil {
			t.Fatalf("falha ao consumir token: %v", err)
		}
		if consumed.TargetID != "card-123" {
			t.Errorf("target id incorreto: %s", consumed.TargetID)
		}

		// Segunda tentativa deve falhar com ErrTokenNotFound (uso único)
		_, err = svc.Consume(act.Token)
		if !errors.Is(err, command.ErrTokenNotFound) && !errors.Is(err, command.ErrTokenAlreadyUsed) {
			t.Errorf("esperava erro de token já usado/não encontrado, obteve %v", err)
		}
	})

	t.Run("Expired token returns ErrTokenExpired", func(t *testing.T) {
		shortSvc := command.NewConfirmationService(10 * time.Millisecond)
		act, _ := shortSvc.IssueToken(command.ActionTypePauseSystem, "sys", "Pausar sistema?", nil)

		time.Sleep(20 * time.Millisecond)

		_, err := shortSvc.Consume(act.Token)
		if !errors.Is(err, command.ErrTokenExpired) {
			t.Errorf("esperava ErrTokenExpired, obteve %v", err)
		}
	})
}
