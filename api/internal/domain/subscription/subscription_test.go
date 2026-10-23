package subscription_test

import (
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/subscription"
)

func TestNormalizedMonthlyCents(t *testing.T) {
	tests := []struct {
		name        string
		interval    subscription.Interval
		amountCents int64
		expected    int64
	}{
		{
			name:        "Mensal R$ 150,00",
			interval:    subscription.IntervalMensal,
			amountCents: 15000,
			expected:    15000,
		},
		{
			name:        "Trimestral R$ 300,00 (R$ 100/mes)",
			interval:    subscription.IntervalTrimestral,
			amountCents: 30000,
			expected:    10000,
		},
		{
			name:        "Semestral R$ 600,00 (R$ 100/mes)",
			interval:    subscription.IntervalSemestral,
			amountCents: 60000,
			expected:    10000,
		},
		{
			name:        "Anual R$ 1.200,00 (R$ 100/mes)",
			interval:    subscription.IntervalAnual,
			amountCents: 120000,
			expected:    10000,
		},
		{
			name:        "Anual com centavos fracionados (truncado sem float) 1000 cents / 12 = 83",
			interval:    subscription.IntervalAnual,
			amountCents: 1000,
			expected:    83,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := subscription.Subscription{
				Interval:    tt.interval,
				AmountCents: tt.amountCents,
			}
			if got := sub.NormalizedMonthlyCents(); got != tt.expected {
				t.Errorf("NormalizedMonthlyCents() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCalculateMRR(t *testing.T) {
	subs := []subscription.Subscription{
		{
			ID:          "sub-1",
			Status:      subscription.StatusAtiva,
			Interval:    subscription.IntervalMensal,
			AmountCents: 20000, // 200/mes
		},
		{
			ID:          "sub-2",
			Status:      subscription.StatusAtiva,
			Interval:    subscription.IntervalAnual,
			AmountCents: 120000, // 100/mes
		},
		{
			ID:          "sub-3",
			Status:      subscription.StatusPausada,
			Interval:    subscription.IntervalMensal,
			AmountCents: 5000, // 50/mes (não entra no current, entra no projected)
		},
		{
			ID:          "sub-4",
			Status:      subscription.StatusInadimplente,
			Interval:    subscription.IntervalTrimestral,
			AmountCents: 30000, // 100/mes (delinquente)
		},
		{
			ID:          "sub-5",
			Status:      subscription.StatusCancelada,
			Interval:    subscription.IntervalMensal,
			AmountCents: 40000, // cancelada (churn)
		},
	}

	metrics := subscription.CalculateMRR(subs)

	// CurrentMRRCents deve somar apenas as ativas: 200 + 100 = 300 (30000 cents)
	if metrics.CurrentMRRCents != 30000 {
		t.Errorf("CurrentMRRCents = %d, want 30000", metrics.CurrentMRRCents)
	}

	// ProjectedMRRCents inclui ativas (300) + pausada (50) + inadimplente (100) = 450 (45000 cents)
	if metrics.ProjectedMRRCents != 45000 {
		t.Errorf("ProjectedMRRCents = %d, want 45000", metrics.ProjectedMRRCents)
	}

	if metrics.ActiveCount != 2 {
		t.Errorf("ActiveCount = %d, want 2", metrics.ActiveCount)
	}
	if metrics.PausedCount != 1 {
		t.Errorf("PausedCount = %d, want 1", metrics.PausedCount)
	}
	if metrics.DelinquentCount != 1 {
		t.Errorf("DelinquentCount = %d, want 1", metrics.DelinquentCount)
	}
	if metrics.DelinquentAmountCents != 30000 {
		t.Errorf("DelinquentAmountCents = %d, want 30000", metrics.DelinquentAmountCents)
	}
	if metrics.CanceledCount != 1 {
		t.Errorf("CanceledCount = %d, want 1", metrics.CanceledCount)
	}
}

func TestCalculateNextDueDate(t *testing.T) {
	sub := subscription.Subscription{
		Interval: subscription.IntervalMensal,
		DueDay:   31,
	}

	// Data base: 31 de janeiro
	baseDate := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	nextDue := sub.CalculateNextDueDate(baseDate)

	// Fevereiro tem 28 dias em 2026 (ano não bissexto)
	if nextDue.Month() != time.February || nextDue.Day() != 28 {
		t.Errorf("NextDueDate para fev com DueDay 31 = %v, want 2026-02-28", nextDue)
	}
}
