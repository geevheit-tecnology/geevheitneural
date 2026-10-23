package subscription

import (
	"errors"
	"time"
)

var (
	ErrInvalidAmount       = errors.New("valor da assinatura deve ser maior que zero")
	ErrInvalidDueDay       = errors.New("dia de vencimento deve estar entre 1 e 31")
	ErrInvalidInterval     = errors.New("periodicidade inválida (use mensal, trimestral, semestral ou anual)")
	ErrInvalidStatus       = errors.New("status de assinatura inválido")
	ErrChargeAlreadyPaid   = errors.New("cobrança já foi liquidada")
	ErrSubscriptionPaused  = errors.New("assinatura está pausada")
)

type Status string

const (
	StatusAtiva        Status = "ativa"
	StatusPausada      Status = "pausada"
	StatusCancelada    Status = "cancelada"
	StatusInadimplente Status = "inadimplente"
)

type Interval string

const (
	IntervalMensal     Interval = "mensal"
	IntervalTrimestral Interval = "trimestral"
	IntervalSemestral  Interval = "semestral"
	IntervalAnual      Interval = "anual"
)

type ChargeStatus string

const (
	ChargeStatusPendente ChargeStatus = "pendente"
	ChargeStatusPaga     ChargeStatus = "paga"
	ChargeStatusAtrasada ChargeStatus = "atrasada"
)

type Subscription struct {
	ID          string     `json:"id"`
	OwnerID     string     `json:"owner_id"`
	ClientID    string     `json:"client_id"`
	AppID       *string    `json:"app_id,omitempty"`
	Plan        string     `json:"plan"`
	Description string     `json:"description"`
	AmountCents int64      `json:"amount_cents"`
	Currency    string     `json:"currency"`
	Interval    Interval   `json:"interval"`
	DueDay      int        `json:"due_day"`
	Status      Status     `json:"status"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	NextDueDate time.Time  `json:"next_due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type SubscriptionCharge struct {
	ID             string       `json:"id"`
	OwnerID        string       `json:"owner_id"`
	SubscriptionID string       `json:"subscription_id"`
	AmountCents    int64        `json:"amount_cents"`
	Currency       string       `json:"currency"`
	DueDate        time.Time    `json:"due_date"`
	Status         ChargeStatus `json:"status"`
	PaidAt         *time.Time   `json:"paid_at,omitempty"`
	PaymentMethod  string       `json:"payment_method"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type MRRMetrics struct {
	CurrentMRRCents      int64 `json:"current_mrr_cents"`
	ProjectedMRRCents    int64 `json:"projected_mrr_cents"`
	ActiveCount          int   `json:"active_count"`
	PausedCount          int   `json:"paused_count"`
	CanceledCount        int   `json:"canceled_count"`
	DelinquentCount      int   `json:"delinquent_count"`
	DelinquentAmountCents int64 `json:"delinquent_amount_cents"`
}

// NormalizedMonthlyCents calcula o valor mensalizado estritamente em centavos inteiros (sem float)
func (s *Subscription) NormalizedMonthlyCents() int64 {
	if s.AmountCents <= 0 {
		return 0
	}
	switch s.Interval {
	case IntervalTrimestral:
		return s.AmountCents / 3
	case IntervalSemestral:
		return s.AmountCents / 6
	case IntervalAnual:
		return s.AmountCents / 12
	case IntervalMensal:
		fallthrough
	default:
		return s.AmountCents
	}
}

// CalculateNextDueDate calcula a data do próximo vencimento respeitando o due_day e intervalo
func (s *Subscription) CalculateNextDueDate(currentDue time.Time) time.Time {
	var monthsToAdd int
	switch s.Interval {
	case IntervalTrimestral:
		monthsToAdd = 3
	case IntervalSemestral:
		monthsToAdd = 6
	case IntervalAnual:
		monthsToAdd = 12
	default:
		monthsToAdd = 1
	}

	year, month, _ := currentDue.Date()
	targetMonth := month + time.Month(monthsToAdd)
	targetYear := year

	for targetMonth > 12 {
		targetMonth -= 12
		targetYear++
	}

	// Ajusta o dia se o mês destino tiver menos dias que s.DueDay (ex: 31 em fevereiro ou abril)
	maxDays := daysInMonth(targetYear, targetMonth)
	day := s.DueDay
	if day > maxDays {
		day = maxDays
	}

	return time.Date(targetYear, targetMonth, day, 0, 0, 0, 0, currentDue.Location())
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// CalculateMRR processa uma lista de assinaturas para derivar as métricas de MRR
func CalculateMRR(subscriptions []Subscription) MRRMetrics {
	var metrics MRRMetrics

	for _, sub := range subscriptions {
		monthly := sub.NormalizedMonthlyCents()

		switch sub.Status {
		case StatusAtiva:
			metrics.ActiveCount++
			metrics.CurrentMRRCents += monthly
			metrics.ProjectedMRRCents += monthly
		case StatusPausada:
			metrics.PausedCount++
			// Assinatura pausada não entra no MRR atual, mas entra na projeção de retorno
			metrics.ProjectedMRRCents += monthly
		case StatusInadimplente:
			metrics.DelinquentCount++
			metrics.DelinquentAmountCents += sub.AmountCents
			// Inadimplente ainda conta na projeção esperada de regularização
			metrics.ProjectedMRRCents += monthly
		case StatusCancelada:
			metrics.CanceledCount++
		}
	}

	return metrics
}
