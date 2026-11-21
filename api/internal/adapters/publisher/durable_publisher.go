package publisher

import (
	"context"
	"fmt"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/app"
)

// DurablePublisher publica referências de prévia sem depender de disco local ou binário Git,
// para uso em runtime serverless onde os arquivos do workspace vivem no PostgreSQL.
type DurablePublisher struct {
	baseURL string
}

func NewDurablePublisher(baseURL string) *DurablePublisher {
	if baseURL == "" {
		baseURL = "http://localhost:8085"
	}
	return &DurablePublisher{baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *DurablePublisher) PublishWorkspace(_ context.Context, cardID, commitMsg string) (*app.PublishedPreviewDTO, error) {
	if strings.TrimSpace(cardID) == "" {
		return nil, fmt.Errorf("id do cartão obrigatório para publicação durável")
	}
	shortID := cardID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	return &app.PublishedPreviewDTO{
		PreviewURL: fmt.Sprintf("%s/v1/cards/%s/preview", p.baseURL, cardID),
		CommitHash: "durable-" + shortID,
		Branch:     "durable-pg",
		Message:    commitMsg,
	}, nil
}
