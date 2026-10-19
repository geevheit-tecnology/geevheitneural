package money_test

import (
	"errors"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/domain/money"
)

func TestMoney_Operations(t *testing.T) {
	t.Run("Add within same currency", func(t *testing.T) {
		m1 := money.NewBRL(1000) // R$ 10,00
		m2 := money.NewBRL(2550) // R$ 25,50
		res, err := m1.Add(m2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Cents != 3550 {
			t.Errorf("expected 3550, got %d", res.Cents)
		}
		if res.Currency != "BRL" {
			t.Errorf("expected BRL, got %s", res.Currency)
		}
	})

	t.Run("Sub within same currency", func(t *testing.T) {
		m1 := money.NewBRL(5000)
		m2 := money.NewBRL(1200)
		res, err := m1.Sub(m2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Cents != 3800 {
			t.Errorf("expected 3800, got %d", res.Cents)
		}
	})

	t.Run("Currency mismatch error", func(t *testing.T) {
		m1 := money.Money{Cents: 100, Currency: "BRL"}
		m2 := money.Money{Cents: 100, Currency: "USD"}
		_, err := m1.Add(m2)
		if !errors.Is(err, money.ErrCurrencyMismatch) {
			t.Errorf("expected ErrCurrencyMismatch, got %v", err)
		}
	})
}

func TestMoney_FormatBRL(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		expected string
	}{
		{"zero", 0, "R$ 0,00"},
		{"cents only", 50, "R$ 0,50"},
		{"single digit real", 500, "R$ 5,00"},
		{"thousands", 320000, "R$ 3.200,00"},
		{"millions", 150000000, "R$ 1.500.000,00"},
		{"negative value", -45000, "-R$ 450,00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := money.NewBRL(tt.cents)
			got := m.FormatBRL()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
