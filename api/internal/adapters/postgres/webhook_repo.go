package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/domain/notification"
	"github.com/geevheit/neuralcenter/api/internal/domain/webhook"
	"github.com/jackc/pgx/v5/pgtype"
)

type WebhookRepository struct {
	q *sqlcgen.Queries
}

func NewWebhookRepository(q *sqlcgen.Queries) *WebhookRepository {
	return &WebhookRepository{q: q}
}

func (r *WebhookRepository) CreateEndpoint(ctx context.Context, ep *webhook.Endpoint) error {
	return r.q.CreateWebhookEndpoint(ctx, sqlcgen.CreateWebhookEndpointParams{
		ID:          textToUUID(ep.ID),
		OwnerID:     textToUUID(ep.OwnerID),
		Url:         ep.URL,
		Description: ep.Description,
		SecretToken: ep.SecretToken,
		Events:      ep.Events,
		Active:      ep.Active,
		CreatedAt:   timeToTimestamptz(ep.CreatedAt),
		UpdatedAt:   timeToTimestamptz(ep.UpdatedAt),
	})
}

func (r *WebhookRepository) GetEndpointByID(ctx context.Context, id string) (*webhook.Endpoint, error) {
	row, err := r.q.GetWebhookEndpointByID(ctx, textToUUID(id))
	if err != nil {
		return nil, err
	}
	return &webhook.Endpoint{
		ID:          uuidToText(row.ID),
		OwnerID:     uuidToText(row.OwnerID),
		URL:         row.Url,
		Description: row.Description,
		SecretToken: row.SecretToken,
		Events:      row.Events,
		Active:      row.Active,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *WebhookRepository) ListEndpoints(ctx context.Context, activeOnly bool) ([]webhook.Endpoint, error) {
	rows, err := r.q.ListWebhookEndpoints(ctx, activeOnly)
	if err != nil {
		return nil, err
	}

	endpoints := make([]webhook.Endpoint, len(rows))
	for i, row := range rows {
		endpoints[i] = webhook.Endpoint{
			ID:          uuidToText(row.ID),
			OwnerID:     uuidToText(row.OwnerID),
			URL:         row.Url,
			Description: row.Description,
			SecretToken: row.SecretToken,
			Events:      row.Events,
			Active:      row.Active,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
	}
	return endpoints, nil
}

func (r *WebhookRepository) DeleteEndpoint(ctx context.Context, id string) error {
	return r.q.DeleteWebhookEndpoint(ctx, textToUUID(id))
}

func (r *WebhookRepository) RecordDelivery(ctx context.Context, d *webhook.Delivery) error {
	payloadBytes, _ := json.Marshal(d.Payload)

	return r.q.CreateWebhookDelivery(ctx, sqlcgen.CreateWebhookDeliveryParams{
		ID:           textToUUID(d.ID),
		OwnerID:      textToUUID(d.OwnerID),
		WebhookID:    textToUUID(d.WebhookID),
		Event:        d.Event,
		Payload:      payloadBytes,
		StatusCode:   int32(d.StatusCode),
		Attempts:     int32(d.Attempts),
		ResponseBody: d.ResponseBody,
		Success:      d.Success,
		CreatedAt:    timeToTimestamptz(d.CreatedAt),
	})
}

func (r *WebhookRepository) ListDeliveries(ctx context.Context, webhookID *string, limit int) ([]webhook.Delivery, error) {
	if limit <= 0 {
		limit = 50
	}

	var wID pgtype.UUID
	if webhookID != nil {
		wID = textToUUID(*webhookID)
	}

	rows, err := r.q.ListWebhookDeliveries(ctx, sqlcgen.ListWebhookDeliveriesParams{
		Column1: wID,
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, err
	}

	deliveries := make([]webhook.Delivery, len(rows))
	for i, row := range rows {
		var payload map[string]any
		_ = json.Unmarshal(row.Payload, &payload)

		deliveries[i] = webhook.Delivery{
			ID:           uuidToText(row.ID),
			OwnerID:      uuidToText(row.OwnerID),
			WebhookID:    uuidToText(row.WebhookID),
			Event:        row.Event,
			Payload:      payload,
			StatusCode:   int(row.StatusCode),
			Attempts:     int(row.Attempts),
			ResponseBody: row.ResponseBody,
			Success:      row.Success,
			CreatedAt:    row.CreatedAt.Time,
		}
	}
	return deliveries, nil
}

type NotificationRepository struct {
	q *sqlcgen.Queries
}

func NewNotificationRepository(q *sqlcgen.Queries) *NotificationRepository {
	return &NotificationRepository{q: q}
}

func (r *NotificationRepository) GetPreferences(ctx context.Context, ownerID string) (*notification.Preferences, error) {
	row, err := r.q.GetNotificationPreferences(ctx, textToUUID(ownerID))
	if err != nil {
		// Retorna defaults se não cadastrado ainda
		return &notification.Preferences{
			OwnerID: ownerID,
			Channels: notification.ChannelPreferences{
				TelegramEnabled: false,
				WhatsAppEnabled: false,
				EmailEnabled:    false,
			},
			Events: notification.EventPreferences{
				IntakeEnabled:              true,
				PixEnabled:                 true,
				PortalAcceptEnabled:        true,
				SubscriptionOverdueEnabled: true,
			},
		}, nil
	}

	var channels notification.ChannelPreferences
	_ = json.Unmarshal(row.Channels, &channels)

	var events notification.EventPreferences
	_ = json.Unmarshal(row.Events, &events)

	return &notification.Preferences{
		OwnerID:  ownerID,
		Channels: channels,
		Events:   events,
	}, nil
}

func (r *NotificationRepository) SavePreferences(ctx context.Context, prefs *notification.Preferences) error {
	channelsBytes, _ := json.Marshal(prefs.Channels)
	eventsBytes, _ := json.Marshal(prefs.Events)

	return r.q.UpsertNotificationPreferences(ctx, sqlcgen.UpsertNotificationPreferencesParams{
		ID:        textToUUID(prefs.OwnerID), // usa ownerID como seed
		OwnerID:   textToUUID(prefs.OwnerID),
		Channels:  channelsBytes,
		Events:    eventsBytes,
		CreatedAt: timeToTimestamptz(timeToTimestamptz(time.Now()).Time),
		UpdatedAt: timeToTimestamptz(time.Now()),
	})
}
