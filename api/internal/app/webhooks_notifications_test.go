package app_test

import (
	"context"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/notification"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/geevheit/neuralcenter/api/internal/domain/webhook"
)

type fakeWebhookRepo struct {
	endpoints  map[string]*webhook.Endpoint
	deliveries []*webhook.Delivery
}

func newFakeWebhookRepo() *fakeWebhookRepo {
	return &fakeWebhookRepo{
		endpoints:  make(map[string]*webhook.Endpoint),
		deliveries: make([]*webhook.Delivery, 0),
	}
}

func (r *fakeWebhookRepo) CreateEndpoint(ctx context.Context, ep *webhook.Endpoint) error {
	r.endpoints[ep.ID] = ep
	return nil
}

func (r *fakeWebhookRepo) GetEndpointByID(ctx context.Context, id string) (*webhook.Endpoint, error) {
	return r.endpoints[id], nil
}

func (r *fakeWebhookRepo) ListEndpoints(ctx context.Context, activeOnly bool) ([]webhook.Endpoint, error) {
	var list []webhook.Endpoint
	for _, ep := range r.endpoints {
		if activeOnly && !ep.Active {
			continue
		}
		list = append(list, *ep)
	}
	return list, nil
}

func (r *fakeWebhookRepo) DeleteEndpoint(ctx context.Context, id string) error {
	delete(r.endpoints, id)
	return nil
}

func (r *fakeWebhookRepo) RecordDelivery(ctx context.Context, d *webhook.Delivery) error {
	r.deliveries = append(r.deliveries, d)
	return nil
}

func (r *fakeWebhookRepo) ListDeliveries(ctx context.Context, webhookID *string, limit int) ([]webhook.Delivery, error) {
	var list []webhook.Delivery
	for _, d := range r.deliveries {
		if webhookID != nil && d.WebhookID != *webhookID {
			continue
		}
		list = append(list, *d)
	}
	return list, nil
}

type fakeNotificationRepo struct {
	prefs *notification.Preferences
}

func (r *fakeNotificationRepo) GetPreferences(ctx context.Context, ownerID string) (*notification.Preferences, error) {
	if r.prefs == nil {
		return &notification.Preferences{OwnerID: ownerID}, nil
	}
	return r.prefs, nil
}

func (r *fakeNotificationRepo) SavePreferences(ctx context.Context, prefs *notification.Preferences) error {
	r.prefs = prefs
	return nil
}

type fakeNotifier struct {
	messages []notification.Message
}

func (f *fakeNotifier) Send(ctx context.Context, msg notification.Message) error {
	f.messages = append(f.messages, msg)
	return nil
}

type fakeDispatcher struct {
	dispatched []map[string]any
}

func (f *fakeDispatcher) Dispatch(ctx context.Context, ep webhook.Endpoint, event string, payload map[string]any) (*webhook.Delivery, error) {
	f.dispatched = append(f.dispatched, payload)
	return &webhook.Delivery{
		ID:         "del-1",
		WebhookID:  ep.ID,
		Event:      event,
		StatusCode: 200,
		Success:    true,
	}, nil
}

func setupWebhookTest() (*app.NotificationWebhooksUseCases, *fakeWebhookRepo, *fakeNotifier, *fakeDispatcher) {
	webhookRepo := newFakeWebhookRepo()
	notifRepo := &fakeNotificationRepo{}
	auditRepo := &subFakeAuditRepo{}

	uow := &subFakeUoW{
		repos: app.Repositories{
			Webhooks:      webhookRepo,
			Notifications: notifRepo,
			Audit:         auditRepo,
		},
	}

	notifier := &fakeNotifier{}
	dispatcher := &fakeDispatcher{}

	uc := app.NewNotificationWebhooksUseCases(uow, notifier, dispatcher, false)
	return uc, webhookRepo, notifier, dispatcher
}

func TestWebhookUseCases_CreateAndSSRF(t *testing.T) {
	uc, _, _, _ := setupWebhookTest()
	principal := system.OwnerPrincipal{ID: "owner-1"}
	ctx := context.Background()

	// 1. Tentar cadastrar endpoint com IP privado (SSRF deve bloquear)
	_, err := uc.CreateWebhookEndpoint(ctx, principal, app.CreateWebhookInput{
		URL:         "http://192.168.1.100/hook",
		Description: "Ataque SSRF interno",
	})
	if err == nil {
		t.Fatal("Esperava bloqueio por SSRF para 192.168.1.100")
	}

	// 2. Cadastrar endpoint válido
	ep, err := uc.CreateWebhookEndpoint(ctx, principal, app.CreateWebhookInput{
		URL:         "https://webhook.site/teste-geev",
		Description: "Webhook n8n Produção",
		Events:      []string{"intake.received", "payment.pix_confirmed"},
	})
	if err != nil {
		t.Fatalf("Falha ao criar endpoint válido: %v", err)
	}

	if ep.SecretToken == "" {
		t.Error("SecretToken não foi gerado")
	}
	if len(ep.SecretToken) != 64 {
		t.Errorf("SecretToken deveria ter 64 caracteres hexadecimais, tem %d", len(ep.SecretToken))
	}
}

func TestWebhookUseCases_DispatchBusinessEvent(t *testing.T) {
	uc, webhookRepo, notifier, dispatcher := setupWebhookTest()
	principal := system.OwnerPrincipal{ID: "owner-1"}
	ctx := context.Background()

	// Cria endpoint ativo
	ep, _ := uc.CreateWebhookEndpoint(ctx, principal, app.CreateWebhookInput{
		URL:         "https://webhook.site/crm-pipe",
		Description: "CRM",
		Events:      []string{"intake.received"},
	})

	// Despacha evento de novo pedido
	payload := map[string]any{"client_name": "Acme Corp", "service": "App Mobile"}
	uc.DispatchBusinessEvent(ctx, "intake.received", "Novo Lead", "Acme solicitou orçamento", payload)

	// Valida notificação enviada ao operador
	if len(notifier.messages) != 1 {
		t.Fatalf("Esperava 1 mensagem no Notifier, obteve %d", len(notifier.messages))
	}
	if notifier.messages[0].Title != "Novo Lead" {
		t.Errorf("Título incorreto: %s", notifier.messages[0].Title)
	}

	// Verifica se o endpoint cadastrado existe no repo
	if len(webhookRepo.endpoints) != 1 || webhookRepo.endpoints[ep.ID] == nil {
		t.Error("Endpoint deveria estar cadastrado no repositório")
	}
	_ = dispatcher
}
