package money

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCurrencyMismatch = errors.New("moedas incompatíveis para operação aritmética")
)

// Money representa um valor monetário em centavos inteiros (int64) e sua moeda.
// Não utiliza ponto flutuante para evitar imprecisões financeiras.
type Money struct {
	Cents    int64
	Currency string
}

// NewBRL cria uma instância de Money em Real Brasileiro (BRL).
func NewBRL(cents int64) Money {
	return Money{
		Cents:    cents,
		Currency: "BRL",
	}
}

// Add soma dois valores monetários da mesma moeda.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != "" && other.Currency != "" && m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	curr := m.Currency
	if curr == "" {
		curr = other.Currency
	}
	return Money{
		Cents:    m.Cents + other.Cents,
		Currency: curr,
	}, nil
}

// Sub subtrai outro valor monetário da mesma moeda.
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != "" && other.Currency != "" && m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	curr := m.Currency
	if curr == "" {
		curr = other.Currency
	}
	return Money{
		Cents:    m.Cents - other.Cents,
		Currency: curr,
	}, nil
}

// FormatBRL formata o valor no padrão monetário brasileiro (ex: R$ 3.200,00 ou R$ -450,00).
func (m Money) FormatBRL() string {
	isNeg := m.Cents < 0
	centsAbs := m.Cents
	if isNeg {
		centsAbs = -centsAbs
	}

	reais := centsAbs / 100
	centavos := centsAbs % 100

	reaisStr := formatThousands(reais)
	formatted := fmt.Sprintf("R$ %s,%02d", reaisStr, centavos)
	if isNeg {
		return "-" + formatted
	}
	return formatted
}

func formatThousands(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var parts []string
	rem := len(s) % 3
	if rem > 0 {
		parts = append(parts, s[:rem])
		s = s[rem:]
	}

	for len(s) > 0 {
		parts = append(parts, s[:3])
		s = s[3:]
	}

	return strings.Join(parts, ".")
}
