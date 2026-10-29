package app

import (
	"context"
	"errors"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/approval"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/notification"
	"github.com/geevheit/neuralcenter/api/internal/domain/portal"
	"github.com/geevheit/neuralcenter/api/internal/domain/subscription"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/geevheit/neuralcenter/api/internal/domain/webhook"
)

var ErrExecutionLocked = errors.New("execução bloqueada: sistema pausado ou agente inativo")

// CardEventDTO representa um evento imutável no histórico do cartão.
type CardEventDTO struct {
	ID        string         `json:"id"`
	OwnerID   string         `json:"owner_id"`
	CardID    string         `json:"card_id"`
	Kind      string         `json:"kind"`
	Payload   map[string]any `json:"payload"`
	ActorType string         `json:"actor_type"`
	ActorID   *string        `json:"actor_id,omitempty"`
	At        time.Time      `json:"at"`
}

// RiskReportDTO representa a ficha de risco exibida antes da entrega.
type RiskReportDTO struct {
	ID           string           `json:"id"`
	CardID       string           `json:"card_id"`
	Tests        []map[string]any `json:"tests"`
	SecretsScan  []map[string]any `json:"secrets_scan"`
	OutdatedDeps []map[string]any `json:"outdated_deps"`
	LgpdPoints   []map[string]any `json:"lgpd_points"`
	Source       string           `json:"source"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// CardListItemDTO representa o cartão enriquecido para visualização no Kanban e na aba Hoje.
type CardListItemDTO struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	ClientID     string     `json:"client_id"`
	ClientName   string     `json:"client_name"`
	AppID        *string    `json:"app_id,omitempty"`
	AppName      string     `json:"app_name"`
	Title        string     `json:"title"`
	PriceCents   int64      `json:"price_cents"`
	EstCostCents int64      `json:"est_cost_cents"`
	Currency     string     `json:"currency"`
	Stage        string     `json:"stage"`
	Priority     int16      `json:"priority"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CardDetailsDTO é o detalhamento exibido na bottom sheet.
type CardDetailsDTO struct {
	Card        CardListItemDTO  `json:"card"`
	Events      []CardEventDTO   `json:"events"`
	RiskReport  *RiskReportDTO   `json:"risk_report,omitempty"`
	IsGate      bool             `json:"is_gate"`
	Actions     []string         `json:"actions"`
	MarginCents int64            `json:"margin_cents"`
}

// AppDTO representa uma aplicação mantida pela software house.
type AppDTO struct {
	ID                  string    `json:"id"`
	ClientID            *string   `json:"client_id,omitempty"`
	ClientName          string    `json:"client_name"`
	Name                string    `json:"name"`
	Hosting             string    `json:"hosting"`
	Database            string    `json:"database"`
	Repository          string    `json:"repository"`
	Domain              string    `json:"domain"`
	Status              string    `json:"status"`
	MonthlyCostCents    int64     `json:"monthly_cost_cents"`
	MonthlyRevenueCents int64     `json:"monthly_revenue_cents"`
	NetMarginCents      int64     `json:"net_margin_cents"`
	IsCostly            bool      `json:"is_costly"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type CreateAppCommand struct {
	ClientID   *string `json:"client_id"`
	Name       string  `json:"name"`
	Hosting    string  `json:"hosting"`
	Database   string  `json:"database"`
	Repository string  `json:"repository"`
	Domain     string  `json:"domain"`
}

type UpdateAppCommand struct {
	ID         string  `json:"id"`
	ClientID   *string `json:"client_id"`
	Name       string  `json:"name"`
	Hosting    string  `json:"hosting"`
	Database   string  `json:"database"`
	Repository string  `json:"repository"`
	Domain     string  `json:"domain"`
	Status     string  `json:"status"`
}

type LedgerEntryDTO struct {
	ID          string    `json:"id"`
	AppID       *string   `json:"app_id,omitempty"`
	AppName     string    `json:"app_name"`
	ClientID    *string   `json:"client_id,omitempty"`
	ClientName  string    `json:"client_name"`
	Kind        string    `json:"kind"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
	OccurredOn  string    `json:"occurred_on"`
	Service     string    `json:"service"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateLedgerEntryCommand struct {
	AppID       *string `json:"app_id"`
	ClientID    *string `json:"client_id"`
	Kind        string  `json:"kind"`
	AmountCents int64   `json:"amount_cents"`
	Currency    string  `json:"currency"`
	OccurredOn  string  `json:"occurred_on"`
	Service     string  `json:"service"`
	Description string  `json:"description"`
}

type CostByServiceDTO struct {
	ServiceName string `json:"service_name"`
	CostCents   int64  `json:"cost_cents"`
}

type FinancesOverviewDTO struct {
	TotalRevenueCents int64              `json:"total_revenue_cents"`
	TotalCostCents    int64              `json:"total_cost_cents"`
	NetProfitCents    int64              `json:"net_profit_cents"`
	CostsByService    []CostByServiceDTO `json:"costs_by_service"`
	RecentEntries     []LedgerEntryDTO   `json:"recent_entries"`
	Suggestion        string             `json:"suggestion"`
}

type ConnectorDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	SecretRef string    `json:"secret_ref"`
	Scope     string    `json:"scope"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DTOs da Fase 4: Comercial e LGPD
type IntakeCommand struct {
	ClientName  string `json:"client_name"`
	Contact     string `json:"contact"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Website     string `json:"website"` // Honeypot antispam
	ClientIP    string `json:"-"`
}

type ProposalDTO struct {
	ID              string     `json:"id"`
	CardID          string     `json:"card_id"`
	AmountCents     int64      `json:"amount_cents"`
	Currency        string     `json:"currency"`
	Body            string     `json:"body"`
	AcceptTokenHash string     `json:"-"`
	AcceptTokenRaw  string     `json:"accept_token,omitempty"`
	AcceptURL       string     `json:"accept_url,omitempty"`
	Status          string     `json:"status"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type PublicProposalViewDTO struct {
	CardTitle   string `json:"card_title"`
	ClientName  string `json:"client_name"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	Body        string `json:"body"`
	Status      string `json:"status"`
}

type ProposalDocumentDTO struct {
	ID                 string    `json:"id"`
	CardID             string    `json:"card_id"`
	Title              string    `json:"title"`
	ClientName         string    `json:"client_name"`
	AppName            string    `json:"app_name"`
	ScopeSummary       string    `json:"scope_summary"`
	Deliverables       []string  `json:"deliverables"`
	TimelineDays       int       `json:"timeline_days"`
	WarrantyDays       int       `json:"warranty_days"`
	PriceCents         int64     `json:"price_cents"`
	Currency           string    `json:"currency"`
	PaymentTerms       string    `json:"payment_terms"`
	AcceptURL          string    `json:"accept_url,omitempty"`
	PixEMV             string    `json:"pix_emv,omitempty"`
	PixTxID            string    `json:"pix_txid,omitempty"`
	FormattedWhatsApp  string    `json:"formatted_whatsapp"`
	TermsAndConditions string    `json:"terms_and_conditions"`
	GeneratedAt        time.Time `json:"generated_at"`
}

type ClientDTO struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ContactRef string    `json:"contact_ref"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ClientExportDTO struct {
	Client  ClientDTO         `json:"client"`
	Cards   []CardListItemDTO `json:"cards"`
	Ledger  []LedgerEntryDTO  `json:"ledger"`
	ExportAt time.Time        `json:"exported_at"`
}

type AuditLogDTO struct {
	ID         string
	OwnerID    string
	Action     string
	ActorType  string
	ActorID    *string
	TargetType string
	TargetID   *string
	Meta       map[string]any
	At         time.Time
}

type CardRepository interface {
	GetByID(ctx context.Context, id string) (*card.Card, error)
	GetEnrichedByID(ctx context.Context, id string) (*CardListItemDTO, error)
	ListEnriched(ctx context.Context) ([]CardListItemDTO, error)
	ListEvents(ctx context.Context, cardID string) ([]CardEventDTO, error)
	GetRiskReport(ctx context.Context, cardID string) (*RiskReportDTO, error)
	Update(ctx context.Context, c *card.Card) error
	Create(ctx context.Context, c *CardListItemDTO) error
	Delete(ctx context.Context, id string) error
}

type AppRepository interface {
	ListWithFinancials(ctx context.Context) ([]AppDTO, error)
	GetByID(ctx context.Context, id string) (*AppDTO, error)
	Create(ctx context.Context, cmd CreateAppCommand) (*AppDTO, error)
	Update(ctx context.Context, cmd UpdateAppCommand) error
	Delete(ctx context.Context, id string) error
}

type FinanceRepository interface {
	GetOverview(ctx context.Context) (*FinancesOverviewDTO, error)
	CreateEntry(ctx context.Context, cmd CreateLedgerEntryCommand) (*LedgerEntryDTO, error)
}

type ConnectorRepository interface {
	List(ctx context.Context) ([]ConnectorDTO, error)
	Toggle(ctx context.Context, id string, active bool) error
}

type ClientRepository interface {
	FindByName(ctx context.Context, name string) (*ClientDTO, error)
	GetByID(ctx context.Context, id string) (*ClientDTO, error)
	Create(ctx context.Context, name, contactRef string) (*ClientDTO, error)
	Delete(ctx context.Context, id string) error
	Export(ctx context.Context, id string) (*ClientExportDTO, error)
}

type ProposalRepository interface {
	Create(ctx context.Context, p *ProposalDTO) error
	GetByTokenHash(ctx context.Context, hash string) (*PublicProposalViewDTO, string, error) // retorna view, cardID, erro
	Accept(ctx context.Context, hash string) error
}

// DTOs da Fase 5: Agentes de IA
type AgentDTO struct {
	ID                string    `json:"id"`
	Role              string    `json:"role"`
	Active            bool      `json:"active"`
	SpendLimitCents   int64     `json:"spend_limit_cents"`
	SpendCurrentCents int64     `json:"spend_current_cents"`
	Scope             string    `json:"scope"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AgentProgressCommand struct {
	Message string         `json:"message"`
	Cost    int64          `json:"cost_cents,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

type AgentSubmitDeliveryCommand struct {
	Summary string         `json:"summary"`
	PRURL   string         `json:"pr_url,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

type AgentRepository interface {
	GetByTokenHash(ctx context.Context, hash string) (*system.AgentPrincipal, error)
	List(ctx context.Context) ([]AgentDTO, error)
	Create(ctx context.Context, role, scope string, spendLimitCents int64, tokenHash string) (*AgentDTO, error)
	AddSpend(ctx context.Context, id string, cents int64) error
	Toggle(ctx context.Context, id string, active bool) error
}

type ApprovalRepository interface {
	Create(ctx context.Context, a *approval.Approval) error
}

type CardEventRepository interface {
	Create(ctx context.Context, event CardEventDTO) error
}

type AuditRepository interface {
	Record(ctx context.Context, log AuditLogDTO) error
}

type SystemStateRepository interface {
	Get(ctx context.Context) (*system.SystemState, error)
	Save(ctx context.Context, state *system.SystemState) error
}

type MoneyWaitingReader interface {
	GetMoneyWaiting(ctx context.Context) (totalCents int64, count int, err error)
}

// DTOs e interfaces do Sprint 02 (Fases 6 a 10)
type PipelineDTO struct {
	ID              string    `json:"id"`
	CardID          string    `json:"card_id"`
	Stage           string    `json:"stage"`
	ActiveRole      string    `json:"active_role"`
	SpendTotalCents int64     `json:"spend_total_cents"`
	SpendLimitCents int64     `json:"spend_limit_cents"`
	Status          string    `json:"status"`
	CurrentTask     string    `json:"current_task"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SpecificationDTO struct {
	ID                 string           `json:"id"`
	CardID             string           `json:"card_id"`
	UserStories        []map[string]any `json:"user_stories"`
	TestsCode          string           `json:"tests_code"`
	TestsPassed        bool             `json:"tests_passed"`
	ClientApproved     bool             `json:"client_approved"`
	PaymentSignalCents int64            `json:"payment_signal_cents"`
	PaymentSignalPaid  bool             `json:"payment_signal_paid"`
	ScopeTokenRaw      string           `json:"scope_token,omitempty"`
	ScopeURL           string           `json:"scope_url,omitempty"`
}

type PlaybookDTO struct {
	ProductType   string         `json:"product_type"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Stack         map[string]any `json:"stack"`
	Checklist     []string       `json:"checklist"`
	AutonomyLevel int            `json:"autonomy_level"`
}

type ModuleDTO struct {
	Name          string         `json:"name"`
	Category      string         `json:"category"`
	Description   string         `json:"description"`
	CodeTemplates map[string]any `json:"code_templates"`
	IsReviewed    bool           `json:"is_reviewed"`
}

type PreviewDTO struct {
	CardID       string         `json:"card_id"`
	PreviewURL   string         `json:"preview_url"`
	Status       string         `json:"status"`
	ExpiresAt    time.Time      `json:"expires_at"`
	RollbackPlan map[string]any `json:"rollback_plan"`
}

type ProjectMemoryDTO struct {
	ID        string    `json:"id"`
	CardID    string    `json:"card_id"`
	ClientID  string    `json:"client_id"`
	Key       string    `json:"key"`
	Category  string    `json:"category"` // requisito, decisao, pendencia, nota
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectMemoryRepository interface {
	Upsert(ctx context.Context, mem *ProjectMemoryDTO) error
	ListByCardID(ctx context.Context, cardID string) ([]ProjectMemoryDTO, error)
}

type RiskReportWriter interface {
	SaveRiskReport(ctx context.Context, report *RiskReportDTO) error
}

type FactoryScoreboardDTO struct {
	AvgLeadTimeHours     float64 `json:"avg_lead_time_hours"`
	TotalAICostCents     int64   `json:"total_ai_cost_cents"`
	TotalInfraCostCents  int64   `json:"total_infra_cost_cents"`
	TotalMarginCents     int64   `json:"total_margin_cents"`
	AvgDecisionTimeHours float64 `json:"avg_decision_time_hours"`
	TotalProjects        int64   `json:"total_projects"`
	FirstTimeApprovalPct float64 `json:"first_time_approval_pct"`
	TotalDefects         int64   `json:"total_defects"`
}

type PriceEstimateDTO struct {
	SuggestedPriceCents int64    `json:"suggested_price_cents"`
	EstCostCents        int64    `json:"est_cost_cents"`
	EstimatedDays       int      `json:"estimated_days"`
	RecommendedStack    []string `json:"recommended_stack"`
	Playbook            string   `json:"playbook"`
}

type PipelineRepository interface {
	Upsert(ctx context.Context, p *PipelineDTO) error
	GetByCardID(ctx context.Context, cardID string) (*PipelineDTO, error)
	List(ctx context.Context) ([]PipelineDTO, error)
	AddSpend(ctx context.Context, cardID string, cents int64) error
}

type SpecificationRepository interface {
	Upsert(ctx context.Context, spec *SpecificationDTO, tokenHash string) error
	GetByCardID(ctx context.Context, cardID string) (*SpecificationDTO, error)
	GetByTokenHash(ctx context.Context, hash string) (*SpecificationDTO, error)
	ApproveByClient(ctx context.Context, hash string) error
}

type PlaybookRepository interface {
	ListPlaybooks(ctx context.Context) ([]PlaybookDTO, error)
	GetPlaybook(ctx context.Context, productType string) (*PlaybookDTO, error)
	UpsertPlaybook(ctx context.Context, p *PlaybookDTO) error
	ListModules(ctx context.Context) ([]ModuleDTO, error)
	UpsertModule(ctx context.Context, m *ModuleDTO) error
}

type PreviewRepository interface {
	Create(ctx context.Context, prev *PreviewDTO) error
	GetByCardID(ctx context.Context, cardID string) (*PreviewDTO, error)
}

type FactoryMetricsRepository interface {
	Record(ctx context.Context, cardID string, leadTimeSec, aiCost, infraCost, margin, decisionTimeSec int64, firstTimeApproved bool, defects int) error
	GetScoreboard(ctx context.Context) (*FactoryScoreboardDTO, error)
}

type PaymentOrderDTO struct {
	ID          string     `json:"id"`
	OwnerID     string     `json:"owner_id"`
	CardID      string     `json:"card_id"`
	ClientID    string     `json:"client_id"`
	AmountCents int64      `json:"amount_cents"`
	Currency    string     `json:"currency"`
	PixEMV      string     `json:"pix_emv"`
	PixTxID     string     `json:"pix_txid"`
	PixKey      string     `json:"pix_key"`
	Status      string     `json:"status"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PaymentOrderRepository interface {
	Create(ctx context.Context, order *PaymentOrderDTO) error
	GetByID(ctx context.Context, id string) (*PaymentOrderDTO, error)
	GetByTxID(ctx context.Context, txid string) (*PaymentOrderDTO, error)
	GetByCardID(ctx context.Context, cardID string) (*PaymentOrderDTO, error)
	UpdateStatus(ctx context.Context, id, status string, paidAt *time.Time) error
}

type PortalTokenRepository interface {
	Create(ctx context.Context, token *portal.PortalToken) error
	GetByHash(ctx context.Context, tokenHash string) (*portal.PortalToken, error)
	GetByCardID(ctx context.Context, cardID string) (*portal.PortalToken, error)
	Update(ctx context.Context, token *portal.PortalToken) error
	RevokeByCardID(ctx context.Context, cardID string, revokedAt time.Time) error
}

type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, sub *subscription.Subscription) error
	GetSubscriptionByID(ctx context.Context, id string) (*subscription.Subscription, error)
	ListSubscriptions(ctx context.Context, clientID *string, status *string) ([]subscription.Subscription, error)
	UpdateSubscriptionStatus(ctx context.Context, id string, status subscription.Status, nextDue *time.Time) error
	UpdateSubscriptionNextDueDate(ctx context.Context, id string, nextDue time.Time) error

	CreateCharge(ctx context.Context, charge *subscription.SubscriptionCharge) error
	GetChargeByID(ctx context.Context, id string) (*subscription.SubscriptionCharge, error)
	ListCharges(ctx context.Context, status *string, dueBefore *time.Time) ([]subscription.SubscriptionCharge, error)
	UpdateChargeStatus(ctx context.Context, id string, status subscription.ChargeStatus, paidAt *time.Time, method string) error
}

type WebhookRepository interface {
	CreateEndpoint(ctx context.Context, ep *webhook.Endpoint) error
	GetEndpointByID(ctx context.Context, id string) (*webhook.Endpoint, error)
	ListEndpoints(ctx context.Context, activeOnly bool) ([]webhook.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id string) error
	RecordDelivery(ctx context.Context, d *webhook.Delivery) error
	ListDeliveries(ctx context.Context, webhookID *string, limit int) ([]webhook.Delivery, error)
}

type NotificationRepository interface {
	GetPreferences(ctx context.Context, ownerID string) (*notification.Preferences, error)
	SavePreferences(ctx context.Context, prefs *notification.Preferences) error
}

// ProjectFile representa um arquivo de código gerado no workspace
type ProjectFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// ExecResult representa o resultado da execução de um comando ou teste na sandbox
type ExecResult struct {
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

// SandboxManager abstrai o ambiente isolado de compilação, escrita e teste de código
type SandboxManager interface {
	InitWorkspace(cardID string) (string, error)
	WriteFiles(cardID string, files []ProjectFile) error
	ReadFiles(cardID string) ([]ProjectFile, error)
	RunValidation(ctx context.Context, cardID string, cmd string, args ...string) (*ExecResult, error)
	GetWorkspacePath(cardID string) string
	CleanupWorkspace(cardID string) error
}

// PublishedPreviewDTO representa o resultado de publicação do workspace
type PublishedPreviewDTO struct {
	CommitHash string `json:"commit_hash"`
	PreviewURL string `json:"preview_url"`
	Branch     string `json:"branch"`
	Message    string `json:"message"`
}

// WorkspacePublisher abstrai o versionador de código e gerador de link de prévia
type WorkspacePublisher interface {
	PublishWorkspace(ctx context.Context, cardID string, commitMsg string) (*PublishedPreviewDTO, error)
}



type Repositories struct {
	Cards          CardRepository
	Apps           AppRepository
	Finances       FinanceRepository
	Connectors     ConnectorRepository
	Clients        ClientRepository
	Proposals      ProposalRepository
	Agents         AgentRepository
	Pipelines      PipelineRepository
	Specifications SpecificationRepository
	Playbooks      PlaybookRepository
	Previews       PreviewRepository
	ProjectMemory  ProjectMemoryRepository
	FactoryMetrics FactoryMetricsRepository
	PaymentOrders  PaymentOrderRepository
	PortalTokens   PortalTokenRepository
	Subscriptions  SubscriptionRepository
	Webhooks       WebhookRepository
	Notifications  NotificationRepository
	Approvals      ApprovalRepository
	CardEvents     CardEventRepository
	Audit          AuditRepository
	SystemState    SystemStateRepository
}

type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(tx Repositories) error) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}

const (
	TopicMoneyWaitingChanged = "money_waiting_changed"
	TopicSystemStateChanged  = "system_state_changed"
	TopicCardsChanged        = "cards_changed"
	TopicAppsChanged         = "apps_changed"
	TopicFinancesChanged     = "finances_changed"
	TopicPipelineChanged     = "pipeline_changed"
	TopicScoreboardChanged   = "scoreboard_changed"
)
