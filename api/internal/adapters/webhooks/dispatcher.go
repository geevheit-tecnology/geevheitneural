package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/webhook"
	"github.com/google/uuid"
)

type DeliveryRecorder interface {
	RecordDelivery(ctx context.Context, d *webhook.Delivery) error
}

type Dispatcher struct {
	logger       *slog.Logger
	httpClient   *http.Client
	recorder     DeliveryRecorder
	requireHTTPS bool
}

func NewDispatcher(logger *slog.Logger, recorder DeliveryRecorder, requireHTTPS bool) *Dispatcher {
	return &Dispatcher{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		recorder:     recorder,
		requireHTTPS: requireHTTPS,
	}
}

// Dispatch envia o evento para o endpoint com assinatura HMAC-SHA256 e retry exponencial
func (d *Dispatcher) Dispatch(ctx context.Context, ep webhook.Endpoint, event string, payload map[string]any) (*webhook.Delivery, error) {
	delivery := &webhook.Delivery{
		ID:        uuid.New().String(),
		OwnerID:   ep.OwnerID,
		WebhookID: ep.ID,
		Event:     event,
		Payload:   payload,
		Attempts:  0,
		Success:   false,
		CreatedAt: time.Now().UTC(),
	}

	// 1. Validação anti-SSRF
	if err := webhook.ValidateTargetURL(ep.URL, d.requireHTTPS); err != nil {
		delivery.StatusCode = 0
		delivery.ResponseBody = fmt.Sprintf("Bloqueado por segurança SSRF: %v", err)
		_ = d.record(delivery)
		return delivery, err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar payload do webhook: %w", err)
	}

	timestamp := time.Now().Unix()
	signature := webhook.ComputeHMACSignature(ep.SecretToken, timestamp, payloadBytes)

	// Tentativas com backoff (máximo 3)
	maxAttempts := 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		delivery.Attempts = attempt

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep.URL, bytes.NewReader(payloadBytes))
		if err != nil {
			lastErr = err
			break
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "GeevheitNeural-Webhook/1.0")
		req.Header.Set("X-Geevheit-Event", event)
		req.Header.Set("X-Geevheit-Timestamp", fmt.Sprintf("%d", timestamp))
		req.Header.Set("X-Geevheit-Signature", fmt.Sprintf("sha256=%s", signature))

		resp, err := d.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxAttempts {
				time.Sleep(time.Duration(attempt*100) * time.Millisecond) // backoff
				continue
			}
			delivery.ResponseBody = fmt.Sprintf("Falha de rede após %d tentativas: %v", attempt, err)
			break
		}

		delivery.StatusCode = resp.StatusCode
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024)) // lê até 1KB da resposta
		_ = resp.Body.Close()
		delivery.ResponseBody = string(bodyBytes)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			delivery.Success = true
			lastErr = nil
			break
		}

		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}
	}

	_ = d.record(delivery)
	return delivery, lastErr
}

func (d *Dispatcher) record(del *webhook.Delivery) error {
	if d.recorder == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return d.recorder.RecordDelivery(ctx, del)
}
