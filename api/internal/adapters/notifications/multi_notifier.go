package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/notification"
)

type MultiNotifier struct {
	logger          *slog.Logger
	httpClient      *http.Client
	telegramToken   string
	telegramChatID  string
	whatsAppAPIURL  string
	whatsAppToken   string
}

func NewMultiNotifier(logger *slog.Logger, telegramToken, telegramChatID, whatsAppURL, whatsAppToken string) *MultiNotifier {
	return &MultiNotifier{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		telegramToken:  telegramToken,
		telegramChatID: telegramChatID,
		whatsAppAPIURL: whatsAppURL,
		whatsAppToken:  whatsAppToken,
	}
}

// Send despacha a notificação para todos os canais configurados de forma resiliente
// Uma falha de canal NUNCA afeta outros canais nem quebra a operação de origem.
func (m *MultiNotifier) Send(ctx context.Context, msg notification.Message) error {
	text := msg.FormatForChat()

	// 1. Telegram
	if m.telegramToken != "" && m.telegramChatID != "" {
		go func() {
			tCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := m.sendTelegram(tCtx, text); err != nil {
				m.logger.Warn("telegram_notification_failed", slog.Any("error", err))
			}
		}()
	}

	// 2. WhatsApp
	if m.whatsAppAPIURL != "" {
		go func() {
			wCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := m.sendWhatsApp(wCtx, text); err != nil {
				m.logger.Warn("whatsapp_notification_failed", slog.Any("error", err))
			}
		}()
	}

	m.logger.Info("notification_dispatched",
		slog.String("event", string(msg.Event)),
		slog.String("title", msg.Title),
	)

	return nil
}

func (m *MultiNotifier) sendTelegram(ctx context.Context, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", m.telegramToken)
	payload := map[string]any{
		"chat_id":    m.telegramChatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram retornou status %d", resp.StatusCode)
	}

	return nil
}

func (m *MultiNotifier) sendWhatsApp(ctx context.Context, text string) error {
	payload := map[string]any{
		"message": text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.whatsAppAPIURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if m.whatsAppToken != "" {
		req.Header.Set("Authorization", "Bearer "+m.whatsAppToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("whatsapp provider retornou status %d", resp.StatusCode)
	}

	return nil
}
