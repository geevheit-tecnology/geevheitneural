package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
)

var (
	ErrGeminiKeyMissing   = errors.New("chave da API Gemini não configurada (GEMINI_API_KEY)")
	ErrGeminiAPIError     = errors.New("erro retornado pela API do Google Gemini")
	ErrGeminiInvalidReply = errors.New("resposta da API Gemini em formato inesperado")
)

// GeminiConfig define as opções de conexão com a API do Google AI Studio (Free Tier)
type GeminiConfig struct {
	APIKey      string
	Model       string        // ex: "gemini-1.5-flash" ou "gemini-2.5-flash"
	BaseURL     string        // URL base, configurável para mocks de teste
	HTTPClient  *http.Client
	MinInterval time.Duration // Intervalo mínimo entre chamadas para respeitar o limite gratuito de 15 RPM (ex: 4s)
}

// GeminiAgentRunner implementa pipeline.AgentRunner conectando-se ao Google Gemini
type GeminiAgentRunner struct {
	cfg        GeminiConfig
	mu         sync.Mutex
	lastCallAt time.Time
}

// NewGeminiAgentRunner instancia o executor do Gemini com proteção de taxa (Rate Limiter)
func NewGeminiAgentRunner(cfg GeminiConfig) *GeminiAgentRunner {
	if cfg.Model == "" {
		cfg.Model = "gemini-1.5-flash"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}
	if cfg.MinInterval <= 0 {
		cfg.MinInterval = 4100 * time.Millisecond // ~14.6 req/minuto (dentro dos 15 RPM free)
	}

	return &GeminiAgentRunner{
		cfg: cfg,
	}
}

// geminiRequest representa a estrutura da requisição para a API do Gemini
type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature      float64 `json:"temperature"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

// geminiResponse representa a resposta da API do Gemini
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// rateLimitSleep aguarda o intervalo necessário para não violar os 15 RPM da camada gratuita
func (r *GeminiAgentRunner) rateLimitSleep(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastCallAt)
	if elapsed < r.cfg.MinInterval {
		waitDuration := r.cfg.MinInterval - elapsed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
		}
	}
	r.lastCallAt = time.Now()
	return nil
}

// RunStep executa um passo do pipeline do agente correspondente ao papel recebido
func (r *GeminiAgentRunner) RunStep(ctx context.Context, role pipeline.Role, prompt string, contextMeta map[string]any) (*pipeline.RunnerResult, error) {
	if r.cfg.APIKey == "" {
		return nil, ErrGeminiKeyMissing
	}

	// 1. Aplicar controle de taxa para camada gratuita (15 RPM)
	if err := r.rateLimitSleep(ctx); err != nil {
		return nil, err
	}

	// 2. Construir prompt instrucional especializado para o papel
	systemInstructions := buildSystemInstruction(role, contextMeta)
	fullPrompt := fmt.Sprintf("%s\n\nInstrução do Passo: %s\nContexto: %s",
		systemInstructions, prompt, formatMeta(contextMeta))

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: fullPrompt},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature:      0.2, // Baixa temperatura para saídas determinísticas e estruturadas
			ResponseMimeType: "application/json",
		},
	}

	rawJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar payload para Gemini: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		strings.TrimRight(r.cfg.BaseURL, "/"), r.cfg.Model, r.cfg.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawJSON))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição HTTP: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := r.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("falha de rede ao conectar com Google Gemini: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler resposta da API Gemini: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d - %s", ErrGeminiAPIError, resp.StatusCode, string(bodyBytes))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON retornado pelo Gemini: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("%w: [%d %s] %s", ErrGeminiAPIError, geminiResp.Error.Code, geminiResp.Error.Status, geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, ErrGeminiInvalidReply
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	responseText = cleanJSONMarkdown(responseText)

	// 3. Estruturar saída em artefatos padronizados
	return parseRoleOutput(role, responseText, contextMeta)
}

// cleanJSONMarkdown remove cercas de código markdown se presentes
func cleanJSONMarkdown(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}

func formatMeta(meta map[string]any) string {
	b, _ := json.Marshal(meta)
	return string(b)
}

// buildSystemInstruction injeta a identidade e as regras de engenharia de cada agente
func buildSystemInstruction(role pipeline.Role, meta map[string]any) string {
	switch role {
	case pipeline.RoleRequisitos:
		return `Você é o Agente de Requisitos da Fábrica de Software GeevheitNeural.
Sua missão é analisar a demanda do cliente e estruturar:
1. Resumo executivo da proposta
2. Histórias de usuário com critérios de aceite BDD (Given-When-Then)
3. Sugestão de escopo
Responda EXCLUSIVAMENTE em formato JSON com o seguinte esquema:
{
  "output": "Resumo em texto claro para o Dono aprovar",
  "stories": [
    {
      "title": "Título da história",
      "given": "Dado que...",
      "when": "Quando...",
      "then": "Então...",
      "passed": true
    }
  ],
  "scope_summary": "Descrição sintética do escopo"
}`

	case pipeline.RoleArquiteto:
		return `Você é o Agente Arquiteto de Software da GeevheitNeural.
Com base nas histórias e escopo aprovados, defina a arquitetura técnica, estrutura de arquivos e playbook recomendado.
Responda EXCLUSIVAMENTE em formato JSON com o seguinte esquema:
{
  "output": "Explicação técnica da arquitetura",
  "playbook_selected": "landing_page|crud_api|pwa_mobile",
  "files_plan": ["index.html", "style.css", "app.js"],
  "tech_stack": "HTML5, TailwindCSS, Vanilla JS"
}`

	case pipeline.RoleDev:
		return `Você é o Agente Desenvolvedor Sênior da GeevheitNeural.
Sua missão é implementar o código COMPLETO, funcional, profissional e sem placeholders (nada de '// TODO' ou código truncado).
Responda EXCLUSIVAMENTE em formato JSON com o seguinte esquema:
{
  "output": "Relatório das implementações concluídas",
  "files": [
    {
      "path": "caminho/do/arquivo.ext",
      "content": "conteúdo completo do arquivo"
    }
  ]
}`

	case pipeline.RoleQA:
		return `Você é o Agente de QA (Garantia de Qualidade) da GeevheitNeural.
Analise os artefatos de código produzidos e valide conformidade com as histórias BDD, sintaxe, e usabilidade.
Responda EXCLUSIVAMENTE em formato JSON com o seguinte esquema:
{
  "output": "Relatório de QA com resultados dos testes",
  "tests_run": 10,
  "passed": 10,
  "coverage": "95%",
  "ready_for_review": true
}`

	case pipeline.RoleSeguranca:
		return `Você é o Agente de Segurança e Conformidade da GeevheitNeural.
Varra os arquivos à procura de segredos expostos, injeções, vulnerabilidades OWASP e adequação à LGPD.
Responda EXCLUSIVAMENTE em formato JSON com o seguinte esquema:
{
  "output": "Varredura de segurança concluída com sucesso.",
  "secrets_found": 0,
  "vulnerabilities_detected": 0,
  "lgpd_compliant": true
}`

	case pipeline.RolePublicador:
		return `Você é o Agente Publicador / Deployer da GeevheitNeural.
Gere a estratégia de entrega, configuração de ambiente de prévia e plano de rollback.
Responda EXCLUSIVAMENTE em formato JSON com o seguinte esquema:
{
  "output": "Aplicação preparada e pronta para aprovação no Portão de Entrega.",
  "preview_url": "URL gerada da prévia",
  "rollback_strategy": "instant_rollback"
}`

	case pipeline.RoleSuporte:
		return `Você é o Agente de Suporte Pós-Entrega da GeevheitNeural.
Inicie o monitoramento ativo e relatórios de integridade.
Responda EXCLUSIVAMENTE em formato JSON com o seguinte esquema:
{
  "output": "Monitoramento e integridade operacionais ativados.",
  "health_status": "optimal",
  "uptime": "100%"
}`

	default:
		return "Você é um agente da GeevheitNeural. Responda em JSON válido estruturado."
	}
}

// parseRoleOutput traduz a resposta JSON do modelo nos artefatos específicos do pipeline
func parseRoleOutput(role pipeline.Role, responseText string, contextMeta map[string]any) (*pipeline.RunnerResult, error) {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(responseText), &parsed); err != nil {
		// Se não foi possível deserializar direto, encapsula o texto como output simples
		parsed = map[string]any{
			"output": responseText,
		}
	}

	outputText, _ := parsed["output"].(string)
	if outputText == "" {
		outputText = fmt.Sprintf("Passo do agente %s concluído com sucesso via Google Gemini.", role)
	}

	result := &pipeline.RunnerResult{
		Role:         role,
		Output:       outputText,
		CostCents:    0, // Free Tier: Custo zero (R$ 0,00)
		Artifacts:    parsed,
		RequiresGate: false,
	}

	switch role {
	case pipeline.RoleRequisitos:
		next := pipeline.RoleArquiteto
		result.NextRole = &next

	case pipeline.RoleArquiteto:
		next := pipeline.RoleDev
		result.NextRole = &next

	case pipeline.RoleDev:
		next := pipeline.RoleQA
		result.NextRole = &next

	case pipeline.RoleQA:
		next := pipeline.RoleSeguranca
		result.NextRole = &next

	case pipeline.RoleSeguranca:
		next := pipeline.RolePublicador
		result.NextRole = &next

	case pipeline.RolePublicador:
		next := pipeline.RoleSuporte
		result.NextRole = &next
		result.RequiresGate = true // Portão 3: Aprovação de Entrega pelo Dono!

	case pipeline.RoleSuporte:
		result.NextRole = nil
	}

	return result, nil
}
