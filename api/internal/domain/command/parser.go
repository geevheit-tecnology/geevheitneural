package command

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Intent representa a intenção do comando interpretado pelo parser determinístico.
type Intent string

const (
	IntentStatus       Intent = "status"
	IntentFinances     Intent = "finances"
	IntentApprove      Intent = "approve"
	IntentReject       Intent = "reject"
	IntentDelete       Intent = "delete"
	IntentBuild        Intent = "build"
	IntentCreate       Intent = "create"
	IntentPauseSystem  Intent = "pause_system"
	IntentResumeSystem Intent = "resume_system"
	IntentUnknown      Intent = "unknown"
)

// ParsedCommand representa a estrutura resultante do parser.
type ParsedCommand struct {
	RawInput   string
	Normalized string
	Intent     Intent
	Target     string // nome do cliente, app ou identificador informado
	Destructive bool   // true para ações irreversíveis que exigem confirmação (F3-02)
}

var (
	ErrAmbiguousTarget = errors.New("ambiguous_target")
	ErrTargetNotFound  = errors.New("target_not_found")
	ErrEmptyCommand    = errors.New("empty_command")
)

// NormalizeText remove acentuação, converte para minúsculas e remove espaços extras.
func NormalizeText(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return strings.ToLower(strings.TrimSpace(result))
}

// ParseCommand interpreta o texto em linguagem natural sem depender de IA externa (ADR-001).
func ParseCommand(raw string) (*ParsedCommand, error) {
	normalized := NormalizeText(raw)
	if normalized == "" {
		return nil, ErrEmptyCommand
	}

	// Remove saudações/invocações direcionadas a assistentes (Jarvis/Jake)
	for _, prefix := range []string{"jarvis,", "jarvis:", "jarvis", "jake,", "jake:", "jake"} {
		if strings.HasPrefix(normalized, prefix) {
			normalized = strings.TrimSpace(strings.TrimPrefix(normalized, prefix))
			break
		}
	}
	if normalized == "" {
		return nil, ErrEmptyCommand
	}

	// Remove cortesias e pedidos indiretos ("pode fazer...", "por favor, criar...")
	normalized = stripPoliteness(normalized)
	if normalized == "" {
		return nil, ErrEmptyCommand
	}

	cmd := &ParsedCommand{
		RawInput:   raw,
		Normalized: normalized,
		Intent:     IntentUnknown,
	}

	// 1. Finanças e Lucro
	if matchAny(normalized, "lucro", "quanto gastei", "gastos", "faturamento", "receita", "custos", "financeiro", "financas") {
		cmd.Intent = IntentFinances
		return cmd, nil
	}

	// 2. Status e Resumo Operacional
	if matchAny(normalized, "status", "resumo", "pendente", "pendentes", "decisoes", "dinheiro parado", "o que tem pra hoje", "hoje") {
		cmd.Intent = IntentStatus
		return cmd, nil
	}

	// 3. Parar de emergência
	if matchAny(normalized, "parar sistema", "parar tudo", "emergencia", "pausar", "parar") && !strings.Contains(normalized, "aprovar") {
		cmd.Intent = IntentPauseSystem
		cmd.Destructive = true
		return cmd, nil
	}

	// 4. Retomar
	if matchAny(normalized, "retomar", "despausar", "continuar", "liberar sistema") {
		cmd.Intent = IntentResumeSystem
		return cmd, nil
	}

	// 5. Recusar (destrutivo)
	if strings.HasPrefix(normalized, "recusar") || strings.HasPrefix(normalized, "rejeitar") || strings.HasPrefix(normalized, "cancelar") {
		cmd.Intent = IntentReject
		cmd.Destructive = true
		cmd.Target = extractTarget(normalized, "recusar", "rejeitar", "cancelar")
		return cmd, nil
	}

	// 5.1 Excluir / Deletar demanda (destrutivo)
	if strings.HasPrefix(normalized, "excluir") || strings.HasPrefix(normalized, "deletar") || strings.HasPrefix(normalized, "remover") || strings.HasPrefix(normalized, "apagar") {
		cmd.Intent = IntentDelete
		cmd.Destructive = true
		cmd.Target = extractTarget(normalized, "excluir", "deletar", "remover", "apagar")
		return cmd, nil
	}

	// 6. Aprovar
	if strings.HasPrefix(normalized, "aprovar") || strings.HasPrefix(normalized, "aceitar") || strings.HasPrefix(normalized, "liberar") || strings.HasPrefix(normalized, "entregar") {
		cmd.Intent = IntentApprove
		cmd.Target = extractTarget(normalized, "aprovar", "aceitar", "liberar", "entregar")
		return cmd, nil
	}

	// 7. Construir / Pipeline da Fábrica Autônoma (Fase 6 - sprit02)
	if strings.HasPrefix(normalized, "construir") || strings.HasPrefix(normalized, "desenvolver") || strings.HasPrefix(normalized, "compilar") || strings.HasPrefix(normalized, "executar") {
		cmd.Intent = IntentBuild
		target := extractTarget(normalized, "construir", "desenvolver", "compilar", "executar")
		target = strings.TrimPrefix(target, "este projeto ")
		target = strings.TrimPrefix(target, "projeto ")
		target = strings.TrimPrefix(target, "demanda ")
		target = strings.TrimPrefix(target, "app ")
		cmd.Target = strings.TrimSpace(target)
		return cmd, nil
	}

	// 8. Criar / Cadastrar Nova Demanda ou Projeto
	if strings.HasPrefix(normalized, "criar") || strings.HasPrefix(normalized, "cadastrar") || strings.HasPrefix(normalized, "novo projeto") || strings.HasPrefix(normalized, "nova demanda") || strings.HasPrefix(normalized, "novo app") || strings.HasPrefix(normalized, "fazer") {
		cmd.Intent = IntentCreate
		target := extractTarget(normalized, "criar", "cadastrar", "novo projeto", "nova demanda", "novo app", "fazer")
		target = strings.TrimPrefix(target, "um projeto ")
		target = strings.TrimPrefix(target, "uma demanda ")
		target = strings.TrimPrefix(target, "um app ")
		target = strings.TrimPrefix(target, "projeto ")
		target = strings.TrimPrefix(target, "demanda ")
		target = strings.TrimPrefix(target, "app ")
		cmd.Target = strings.TrimSpace(target)
		return cmd, nil
	}

	return cmd, nil
}

// politePrefixes são expressões de cortesia ou pedido indireto que antecedem o verbo.
// Ordenadas da mais longa para a mais curta para que a mais específica vença.
var politePrefixes = []string{
	"voce poderia", "voce pode", "quero que voce", "gostaria de",
	"por favor", "poderia", "preciso", "podes", "pode", "quero", "vamos", "favor",
}

// stripPoliteness remove cortesias do início do comando, respeitando limites de palavra.
func stripPoliteness(text string) string {
	for {
		trimmed := strings.TrimLeft(text, " ,.:;-")
		stripped := false
		for _, p := range politePrefixes {
			if !strings.HasPrefix(trimmed, p) {
				continue
			}
			rest := trimmed[len(p):]
			if rest == "" || strings.ContainsRune(" ,.:;", rune(rest[0])) {
				text = strings.TrimLeft(rest, " ,.:;-")
				stripped = true
				break
			}
		}
		if !stripped {
			return strings.TrimSpace(trimmed)
		}
	}
}

func matchAny(text string, keywords ...string) bool {
	for _, kw := range keywords {
		if text == kw || strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func extractTarget(text string, verbs ...string) string {
	for _, v := range verbs {
		if strings.HasPrefix(text, v) {
			target := strings.TrimPrefix(text, v)
			target = strings.TrimPrefix(target, " o ")
			target = strings.TrimPrefix(target, " a ")
			target = strings.TrimPrefix(target, " de ")
			target = strings.TrimPrefix(target, " da ")
			target = strings.TrimPrefix(target, " do ")
			return strings.TrimSpace(target)
		}
	}
	return strings.TrimSpace(text)
}

// TargetMatch representa o cartão correspondente encontrado após busca difusa.
type TargetMatch struct {
	CardID     string
	ClientName string
	Title      string
	Stage      string
	PriceCents int64
}

// ResolveTarget busca o alvo entre os cartões ativos e trata ambiguidades.
func ResolveTarget(target string, available []TargetMatch) (*TargetMatch, error) {
	normTarget := NormalizeText(target)
	if normTarget == "" {
		return nil, fmt.Errorf("especifique o nome do cliente ou do cartão: %w", ErrTargetNotFound)
	}

	var exactMatches []TargetMatch
	var partialMatches []TargetMatch

	for _, c := range available {
		normClient := NormalizeText(c.ClientName)
		normTitle := NormalizeText(c.Title)

		// Correspondência exata
		if normClient == normTarget || normTitle == normTarget {
			exactMatches = append(exactMatches, c)
			continue
		}

		// Correspondência parcial por palavras relevantes (>= 3 chars)
		if strings.Contains(normClient, normTarget) || strings.Contains(normTitle, normTarget) {
			partialMatches = append(partialMatches, c)
			continue
		}

		words := strings.Fields(normTarget)
		for _, w := range words {
			if len(w) >= 3 && (strings.Contains(normClient, w) || strings.Contains(normTitle, w)) {
				partialMatches = append(partialMatches, c)
				break
			}
		}
	}

	matches := exactMatches
	if len(matches) == 0 {
		matches = partialMatches
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("nenhum cartão encontrado para '%s': %w", target, ErrTargetNotFound)
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = fmt.Sprintf("%s (%s)", m.ClientName, m.Title)
		}
		return nil, fmt.Errorf("comando ambíguo, encontrei %d cartões: %s: %w", len(matches), strings.Join(names, ", "), ErrAmbiguousTarget)
	}

	return &matches[0], nil
}
