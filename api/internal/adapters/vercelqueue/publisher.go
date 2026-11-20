// Package vercelqueue implementa a publicação de identificadores de tarefas.
// A fila externa é somente um sinal: o estado autoritativo fica no PostgreSQL.
package vercelqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var validName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type Publisher struct {
	endpoint string
	client   *http.Client
}

// NewPublisher aceita uma região e tópico, nunca uma URL arbitrária.
func NewPublisher(region, topic string, client *http.Client) (*Publisher, error) {
	if len(region) == 0 || len(region) > 16 || !validName.MatchString(region) || len(topic) == 0 || len(topic) > 256 || !validName.MatchString(topic) {
		return nil, fmt.Errorf("região ou tópico inválido para fila")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Publisher{endpoint: "https://" + region + ".vercel-queue.com/api/v3/topic/" + topic, client: client}, nil
}

// Publish envia somente o ID, usa-o como chave de idempotência e nunca persiste JWT do dono.
// O token OIDC é obtido na invocação da Vercel e não é registrado ou incluído no corpo.
func (p *Publisher) Publish(ctx context.Context, oidcToken, jobID string) error {
	if strings.TrimSpace(oidcToken) == "" {
		return fmt.Errorf("identidade OIDC da Vercel indisponível")
	}
	id, err := uuid.Parse(jobID)
	if err != nil || id == uuid.Nil {
		return fmt.Errorf("identificador de tarefa inválido")
	}
	body, err := json.Marshal(struct {
		JobID string `json:"job_id"`
	}{JobID: id.String()})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("falha ao preparar mensagem")
	}
	req.Header.Set("Authorization", "Bearer "+oidcToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Vqs-Idempotency-Key", id.String())
	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("não foi possível enviar sinal à fila")
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("fila recusou o sinal (HTTP %d)", res.StatusCode)
	}
	return nil
}
