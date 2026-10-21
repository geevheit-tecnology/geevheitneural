package payment_test

import (
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/payment"
)

func TestGeneratePixEMV(t *testing.T) {
	val := money.NewBRL(350000)

	emv := payment.GeneratePixEMV("financeiro@geevheit.com", "TESTETXID123", val, "GEEVHEIT", "SAO PAULO")

	if !strings.HasPrefix(emv, "000201") {
		t.Errorf("esperado prefixo 000201, obtido: %s", emv)
	}
	if !strings.Contains(emv, "br.gov.bcb.pix") {
		t.Errorf("esperado conter br.gov.bcb.pix, obtido: %s", emv)
	}
	if !strings.Contains(emv, "financeiro@geevheit.com") {
		t.Errorf("esperado conter chave pix, obtido: %s", emv)
	}
	if !strings.Contains(emv, "54073500.00") {
		t.Errorf("esperado valor formatado 54073500.00, obtido: %s", emv)
	}
	if !strings.Contains(emv, "TESTETXID123") {
		t.Errorf("esperado conter txid, obtido: %s", emv)
	}
	if !strings.Contains(emv, "6304") {
		t.Errorf("esperado conter tag CRC 6304, obtido: %s", emv)
	}
}

func TestOrderLifecycle(t *testing.T) {
	val := money.NewBRL(150000)
	now := time.Now()

	order, err := payment.NewOrder("ord-1", "own-1", "card-1", "cli-1", val, "pix@geevheit.com", "TX123", now)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if order.Status != payment.StatusPending {
		t.Errorf("esperado status pending, obtido: %s", order.Status)
	}

	paidAt := now.Add(5 * time.Minute)
	if err := order.MarkPaid(paidAt); err != nil {
		t.Fatalf("falha ao marcar pago: %v", err)
	}

	if order.Status != payment.StatusPaid {
		t.Errorf("esperado status paid, obtido: %s", order.Status)
	}

	// Não pode pagar duas vezes
	if err := order.MarkPaid(paidAt); err != payment.ErrAlreadyPaid {
		t.Errorf("esperado ErrAlreadyPaid, obtido: %v", err)
	}
}
