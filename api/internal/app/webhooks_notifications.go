package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/notification"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/geevheit/neuralcenter/api/internal/domain/webhook"
	"github.com/google/uuid"
)

var (
	ErrWebhookNotFound = errors.New("endpoint de webhook não encontrado")
)

type WebhookDispatcher interface {
	Dispatch(ctx context.Context, ep webhook.Endpoint, event string, payload map[string]any) (*webhook.Delivery, error)
}

type NotificationWebhooksUseCases struct {
	uow          UnitOfWork
	notifier     notification.Notifier
	dispatcher   WebhookDispatcher
	requireHTTPS bool
}

func NewNotificationWebhooksUseCases(
	uow UnitOfWork,
	notifier notification.Notifier,
	dispatcher WebhookDispatcher,
	requireHTTPS bool,
) *NotificationWebhooksUseCases {
	return &NotificationWebhooksUseCases{
		uow:          uow,
		notifier:     notifier,
		dispatcher:   dispatcher,
		requireHTTPS: requireHTTPS,
	}
}

type CreateWebhookInput struct {
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
}

// CreateWebhookEndpoint cadastra um novo webhook e gera o secret HMAC
func (uc *NotificationWebhooksUseCases) CreateWebhookEndpoint(ctx context.Context, principal system.OwnerPrincipal, in CreateWebhookInput) (*webhook.Endpoint, error) {
	if err := webhook.ValidateTargetURL(in.URL, uc.requireHTTPS); err != nil {
		return nil, err
	}

	if len(in.Events) == 0 {
		in.Events = []string{"*"}
	}

	// Gera segredo HMAC aleatório de 32 bytes (64 caracteres hexadecimais)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("falha ao gerar segredo aleatório: %w", err)
	}
	secret := hex.EncodeToString(secretBytes)

	now := time.Now().UTC()
	ep := &webhook.Endpoint{
		ID:          uuid.New().String(),
		OwnerID:     principal.ID,
		URL:         in.URL,
		Description: in.Description,
		SecretToken: secret,
		Events:      in.Events,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		if err := tx.Webhooks.CreateEndpoint(ctx, ep); err != nil {
			return fmt.Errorf("falha ao criar endpoint de webhook: %w", err)
		}

		if tx.Audit != nil {
			actorID := principal.ID
			_ = tx.Audit.Record(ctx, AuditLogDTO{
				Action:     "webhook_endpoint_created",
				ActorType:  "owner",
				ActorID:    &actorID,
				TargetType: "webhook_endpoint",
				TargetID:   &ep.ID,
				Meta: map[string]any{
					"url":         ep.URL,
					"description": ep.Description,
					"events":      ep.Events,
				},
				At: now,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return ep, nil
}

// ListWebhookEndpoints lista os endpoints cadastrados
func (uc *NotificationWebhooksUseCases) ListWebhookEndpoints(ctx context.Context, principal system.OwnerPrincipal) ([]webhook.Endpoint, error) {
	var endpoints []webhook.Endpoint
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		endpoints, err = tx.Webhooks.ListEndpoints(ctx, false)
		return err
	})
	return endpoints, err
}

// DeleteWebhookEndpoint remove um endpoint
func (uc *NotificationWebhooksUseCases) DeleteWebhookEndpoint(ctx context.Context, principal system.OwnerPrincipal, id string) error {
	now := time.Now().UTC()
	return uc.uow.RunInTx(ctx, func(tx Repositories) error {
		if err := tx.Webhooks.DeleteEndpoint(ctx, id); err != nil {
			return err
		}

		if tx.Audit != nil {
			actorID := principal.ID
			_ = tx.Audit.Record(ctx, AuditLogDTO{
				Action:     "webhook_endpoint_deleted",
				ActorType:  "owner",
				ActorID:    &actorID,
				TargetType: "webhook_endpoint",
				TargetID:   &id,
				At:         now,
			})
		}
		return nil
	})
}

// ListDeliveries lista histórico recente de disparos
func (uc *NotificationWebhooksUseCases) ListDeliveries(ctx context.Context, principal system.OwnerPrincipal, webhookID *string) ([]webhook.Delivery, error) {
	var deliveries []webhook.Delivery
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		deliveries, err = tx.Webhooks.ListDeliveries(ctx, webhookID, 50)
		return err
	})
	return deliveries, err
}

// Redeliver reencontra a entrega e dispara novamente
func (uc *NotificationWebhooksUseCases) Redeliver(ctx context.Context, principal system.OwnerPrincipal, webhookID string, event string, payload map[string]any) (*webhook.Delivery, error) {
	var ep *webhook.Endpoint
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		ep, err = tx.Webhooks.GetEndpointByID(ctx, webhookID)
		return err
	})
	if err != nil || ep == nil {
		return nil, ErrWebhookNotFound
	}

	if uc.dispatcher == nil {
		return nil, errors.New("dispatcher não configurado")
	}

	return uc.dispatcher.Dispatch(ctx, *ep, event, payload)
}

// DispatchBusinessEvent despacha eventos tanto para canais humanos quanto para webhooks de saída
func (uc *NotificationWebhooksUseCases) DispatchBusinessEvent(ctx context.Context, event string, title, body string, payload map[string]any) {
	// 1. Notificação Humana (Telegram/WhatsApp/Email)
	if uc.notifier != nil {
		msg := notification.BuildTemplate(notification.EventType(event), title, body, payload)
		_ = uc.notifier.Send(ctx, msg)
	}

	// 2. Webhooks de saída (n8n, CRM, etc.)
	if uc.dispatcher == nil {
		return
	}

	var endpoints []webhook.Endpoint
	_ = uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		endpoints, err = tx.Webhooks.ListEndpoints(ctx, true)
		return err
	})

	for _, ep := range endpoints {
		if ep.MatchesEvent(event) {
			endpointCopy := ep
			go func() {
				dispatchCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_, _ = uc.dispatcher.Dispatch(dispatchCtx, endpointCopy, event, payload)
			}()
		}
	}
}

// GetNotificationPreferences busca as preferências do dono
func (uc *NotificationWebhooksUseCases) GetNotificationPreferences(ctx context.Context, principal system.OwnerPrincipal) (*notification.Preferences, error) {
	var prefs *notification.Preferences
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		prefs, err = tx.Notifications.GetPreferences(ctx, principal.ID)
		return err
	})
	return prefs, err
}

// SaveNotificationPreferences atualiza as preferências do dono
func (uc *NotificationWebhooksUseCases) SaveNotificationPreferences(ctx context.Context, principal system.OwnerPrincipal, prefs notification.Preferences) error {
	prefs.OwnerID = principal.ID
	now := time.Now().UTC()

	return uc.uow.RunInTx(ctx, func(tx Repositories) error {
		if err := tx.Notifications.SavePreferences(ctx, &prefs); err != nil {
			return err
		}

		if tx.Audit != nil {
			actorID := principal.ID
			_ = tx.Audit.Record(ctx, AuditLogDTO{
				Action:     "notification_preferences_updated",
				ActorType:  "owner",
				ActorID:    &actorID,
				TargetType: "notification_preferences",
				At:         now,
			})
		}
		return nil
	})
}
