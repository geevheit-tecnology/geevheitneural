package notification

import (
	"context"
	"fmt"
	"time"
)

type EventType string

const (
	EventIntakeReceived       EventType = "intake.received"
	EventPixConfirmed         EventType = "payment.pix_confirmed"
	EventPortalAccepted       EventType = "portal.delivery_accepted"
	EventSubscriptionOverdue  EventType = "subscription.overdue"
)

type Message struct {
	Event     EventType      `json:"event"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Data      map[string]any `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type Notifier interface {
	Send(ctx context.Context, msg Message) error
}

type ChannelPreferences struct {
	TelegramEnabled bool `json:"telegram_enabled"`
	WhatsAppEnabled bool `json:"whatsapp_enabled"`
	EmailEnabled    bool `json:"email_enabled"`
}

type EventPreferences struct {
	IntakeEnabled              bool `json:"intake_enabled"`
	PixEnabled                 bool `json:"pix_enabled"`
	PortalAcceptEnabled        bool `json:"portal_accept_enabled"`
	SubscriptionOverdueEnabled bool `json:"subscription_overdue_enabled"`
}

type Preferences struct {
	OwnerID  string             `json:"owner_id"`
	Channels ChannelPreferences `json:"channels"`
	Events   EventPreferences   `json:"events"`
}

// BuildTemplate cria uma mensagem limpa e padronizada para operadores humanos
func BuildTemplate(event EventType, title, body string, data map[string]any) Message {
	return Message{
		Event:     event,
		Title:     title,
		Body:      body,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// FormatForChat formata a mensagem para mensageiros como Telegram e WhatsApp
func (m Message) FormatForChat() string {
	icon := "📢"
	switch m.Event {
	case EventIntakeReceived:
		icon = "📥"
	case EventPixConfirmed:
		icon = "💰"
	case EventPortalAccepted:
		icon = "✅"
	case EventSubscriptionOverdue:
		icon = "⚠️"
	}

	text := fmt.Sprintf("%s *%s*\n\n%s", icon, m.Title, m.Body)
	return text
}
