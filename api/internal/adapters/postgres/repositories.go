package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/approval"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CardRepository struct {
	q *sqlcgen.Queries
}

func NewCardRepository(q *sqlcgen.Queries) *CardRepository {
	return &CardRepository{q: q}
}

func (r *CardRepository) GetByID(ctx context.Context, id string) (*card.Card, error) {
	row, err := r.q.GetCardByID(ctx, textToUUID(id))
	if err != nil {
		return nil, fmt.Errorf("cartão não encontrado: %w", err)
	}

	var appID *string
	if row.AppID.Valid {
		s := uuidToText(row.AppID)
		appID = &s
	}

	var dueDate = timestamptzToTime(pgtype.Timestamptz{Time: row.DueDate.Time, Valid: row.DueDate.Valid})

	return &card.Card{
		ID:        uuidToText(row.ID),
		OwnerID:   uuidToText(row.OwnerID),
		ClientID:  uuidToText(row.ClientID),
		AppID:     appID,
		Type:      card.Type(row.Type),
		Title:     row.Title,
		Price:     money.Money{Cents: row.PriceCents, Currency: row.Currency},
		EstCost:   money.Money{Cents: row.EstCostCents, Currency: row.Currency},
		Stage:     card.Stage(row.Stage),
		Priority:  row.Priority,
		DueDate:   &dueDate,
		CreatedAt: timestamptzToTime(row.CreatedAt),
		UpdatedAt: timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *CardRepository) GetEnrichedByID(ctx context.Context, id string) (*app.CardListItemDTO, error) {
	row, err := r.q.GetCardEnrichedByID(ctx, textToUUID(id))
	if err != nil {
		return nil, fmt.Errorf("cartão não encontrado: %w", err)
	}

	var appID *string
	if row.AppID.Valid {
		s := uuidToText(row.AppID)
		appID = &s
	}

	var dueDate *time.Time
	if row.DueDate.Valid {
		t := row.DueDate.Time
		dueDate = &t
	}

	return &app.CardListItemDTO{
		ID:           uuidToText(row.ID),
		Type:         string(row.Type),
		ClientID:     uuidToText(row.ClientID),
		ClientName:   row.ClientName,
		AppID:        appID,
		AppName:      row.AppName,
		Title:        row.Title,
		PriceCents:   row.PriceCents,
		EstCostCents: row.EstCostCents,
		Currency:     row.Currency,
		Stage:        string(row.Stage),
		Priority:     row.Priority,
		DueDate:      dueDate,
		CreatedAt:    timestamptzToTime(row.CreatedAt),
		UpdatedAt:    timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *CardRepository) ListEnriched(ctx context.Context) ([]app.CardListItemDTO, error) {
	rows, err := r.q.ListCardsEnriched(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar cartões: %w", err)
	}

	out := make([]app.CardListItemDTO, len(rows))
	for i, row := range rows {
		var appID *string
		if row.AppID.Valid {
			s := uuidToText(row.AppID)
			appID = &s
		}
		var dueDate *time.Time
		if row.DueDate.Valid {
			t := row.DueDate.Time
			dueDate = &t
		}
		out[i] = app.CardListItemDTO{
			ID:           uuidToText(row.ID),
			Type:         string(row.Type),
			ClientID:     uuidToText(row.ClientID),
			ClientName:   row.ClientName,
			AppID:        appID,
			AppName:      row.AppName,
			Title:        row.Title,
			PriceCents:   row.PriceCents,
			EstCostCents: row.EstCostCents,
			Currency:     row.Currency,
			Stage:        string(row.Stage),
			Priority:     row.Priority,
			DueDate:      dueDate,
			CreatedAt:    timestamptzToTime(row.CreatedAt),
			UpdatedAt:    timestamptzToTime(row.UpdatedAt),
		}
	}
	return out, nil
}

func (r *CardRepository) ListEvents(ctx context.Context, cardID string) ([]app.CardEventDTO, error) {
	rows, err := r.q.ListCardEventsByCardID(ctx, textToUUID(cardID))
	if err != nil {
		return nil, fmt.Errorf("erro ao listar eventos: %w", err)
	}

	out := make([]app.CardEventDTO, len(rows))
	for i, row := range rows {
		var payload map[string]any
		_ = json.Unmarshal(row.Payload, &payload)

		var actorID *string
		if row.ActorID.Valid {
			s := uuidToText(row.ActorID)
			actorID = &s
		}

		out[i] = app.CardEventDTO{
			ID:        uuidToText(row.ID),
			OwnerID:   uuidToText(row.OwnerID),
			CardID:    uuidToText(row.CardID),
			Kind:      row.Kind,
			Payload:   payload,
			ActorType: row.ActorType,
			ActorID:   actorID,
			At:        timestamptzToTime(row.At),
		}
	}
	return out, nil
}

func (r *CardRepository) GetRiskReport(ctx context.Context, cardID string) (*app.RiskReportDTO, error) {
	row, err := r.q.GetRiskReportByCardID(ctx, textToUUID(cardID))
	if err != nil {
		return nil, err
	}

	var tests []map[string]any
	_ = json.Unmarshal(row.Tests, &tests)

	var secretsScan []map[string]any
	_ = json.Unmarshal(row.SecretsScan, &secretsScan)

	var outdatedDeps []map[string]any
	_ = json.Unmarshal(row.OutdatedDeps, &outdatedDeps)

	var lgpdPoints []map[string]any
	_ = json.Unmarshal(row.LgpdPoints, &lgpdPoints)

	return &app.RiskReportDTO{
		ID:           uuidToText(row.ID),
		CardID:       uuidToText(row.CardID),
		Tests:        tests,
		SecretsScan:  secretsScan,
		OutdatedDeps: outdatedDeps,
		LgpdPoints:   lgpdPoints,
		Source:       row.Source,
		UpdatedAt:    timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *CardRepository) SaveRiskReport(ctx context.Context, report *app.RiskReportDTO) error {
	now := time.Now()
	id := report.ID
	if id == "" {
		if existing, err := r.GetRiskReport(ctx, report.CardID); err == nil && existing != nil && existing.ID != "" {
			id = existing.ID
		} else {
			id = uuid.NewString()
		}
	}

	testsBytes, _ := json.Marshal(report.Tests)
	secretsBytes, _ := json.Marshal(report.SecretsScan)
	depsBytes, _ := json.Marshal(report.OutdatedDeps)
	lgpdBytes, _ := json.Marshal(report.LgpdPoints)
	src := report.Source
	if src == "" {
		src = "factory_pipeline"
	}

	row, err := r.q.UpsertRiskReport(ctx, sqlcgen.UpsertRiskReportParams{
		ID:           textToUUID(id),
		OwnerID:      textToUUID("a0000000-0000-0000-0000-000000000001"),
		CardID:       textToUUID(report.CardID),
		Tests:        testsBytes,
		SecretsScan:  secretsBytes,
		OutdatedDeps: depsBytes,
		LgpdPoints:   lgpdBytes,
		Source:       src,
		CreatedAt:    timeToTimestamptz(now),
		UpdatedAt:    timeToTimestamptz(now),
	})
	if err != nil {
		return fmt.Errorf("falha ao salvar ficha de risco: %w", err)
	}
	report.ID = uuidToText(row.ID)
	report.UpdatedAt = timestamptzToTime(row.UpdatedAt)
	return nil
}

func (r *CardRepository) Update(ctx context.Context, c *card.Card) error {
	return r.q.UpdateCard(ctx, sqlcgen.UpdateCardParams{
		ID:           textToUUID(c.ID),
		Stage:        sqlcgen.GeevCardStage(c.Stage),
		Priority:     c.Priority,
		PriceCents:   c.Price.Cents,
		EstCostCents: c.EstCost.Cents,
		UpdatedAt:    timeToTimestamptz(c.UpdatedAt),
	})
}

func (r *CardRepository) Create(ctx context.Context, c *app.CardListItemDTO) error {
	var appID pgtype.UUID
	if c.AppID != nil && *c.AppID != "" {
		appID = textToUUID(*c.AppID)
	}

	var dueDate pgtype.Date
	if c.DueDate != nil {
		dueDate = pgtype.Date{Time: *c.DueDate, Valid: true}
	}

	return r.q.CreateCard(ctx, sqlcgen.CreateCardParams{
		ID:           textToUUID(c.ID),
		OwnerID:      textToUUID("a0000000-0000-0000-0000-000000000001"),
		ClientID:     textToUUID(c.ClientID),
		AppID:        appID,
		Type:         sqlcgen.GeevCardType(c.Type),
		Title:        c.Title,
		PriceCents:   c.PriceCents,
		EstCostCents: c.EstCostCents,
		Currency:     c.Currency,
		Stage:        sqlcgen.GeevCardStage(c.Stage),
		Priority:     c.Priority,
		DueDate:      dueDate,
		CreatedAt:    timeToTimestamptz(c.CreatedAt),
		UpdatedAt:    timeToTimestamptz(c.UpdatedAt),
	})
}

func (r *CardRepository) Delete(ctx context.Context, id string) error {
	return r.q.DeleteCard(ctx, textToUUID(id))
}

type ApprovalRepository struct {
	q *sqlcgen.Queries
}

func NewApprovalRepository(q *sqlcgen.Queries) *ApprovalRepository {
	return &ApprovalRepository{q: q}
}

func (r *ApprovalRepository) Create(ctx context.Context, a *approval.Approval) error {
	return r.q.CreateApproval(ctx, sqlcgen.CreateApprovalParams{
		ID:        textToUUID(a.ID),
		OwnerID:   textToUUID(a.OwnerID),
		CardID:    textToUUID(a.CardID),
		Gate:      sqlcgen.GeevCardStage(a.Gate),
		Decision:  sqlcgen.GeevDecision(a.Decision),
		DecidedBy: textToUUID(a.DecidedBy),
		DecidedAt: timeToTimestamptz(a.DecidedAt),
		Note:      a.Note,
	})
}

type CardEventRepository struct {
	q *sqlcgen.Queries
}

func NewCardEventRepository(q *sqlcgen.Queries) *CardEventRepository {
	return &CardEventRepository{q: q}
}

func (r *CardEventRepository) Create(ctx context.Context, e app.CardEventDTO) error {
	payloadBytes, err := json.Marshal(e.Payload)
	if err != nil {
		payloadBytes = []byte("{}")
	}

	eventID := e.ID
	if eventID == "" {
		eventID = uuid.NewString()
	}
	ownerID := e.OwnerID
	if ownerID == "" {
		ownerID = "a0000000-0000-0000-0000-000000000001"
	}
	eventAt := e.At
	if eventAt.IsZero() {
		eventAt = time.Now()
	}

	var actorID pgtype.UUID
	if e.ActorID != nil {
		actorID = textToUUID(*e.ActorID)
	}

	return r.q.CreateCardEvent(ctx, sqlcgen.CreateCardEventParams{
		ID:        textToUUID(eventID),
		OwnerID:   textToUUID(ownerID),
		CardID:    textToUUID(e.CardID),
		Kind:      e.Kind,
		Payload:   payloadBytes,
		ActorType: e.ActorType,
		ActorID:   actorID,
		At:        timeToTimestamptz(eventAt),
	})
}

type AuditRepository struct {
	q *sqlcgen.Queries
}

func NewAuditRepository(q *sqlcgen.Queries) *AuditRepository {
	return &AuditRepository{q: q}
}

func (r *AuditRepository) Record(ctx context.Context, l app.AuditLogDTO) error {
	metaBytes, err := json.Marshal(l.Meta)
	if err != nil {
		metaBytes = []byte("{}")
	}

	auditID := l.ID
	if auditID == "" {
		auditID = uuid.NewString()
	}
	ownerID := l.OwnerID
	if ownerID == "" {
		ownerID = "a0000000-0000-0000-0000-000000000001"
	}
	auditAt := l.At
	if auditAt.IsZero() {
		auditAt = time.Now()
	}

	var actorID pgtype.UUID
	if l.ActorID != nil {
		actorID = textToUUID(*l.ActorID)
	}

	var targetID pgtype.UUID
	if l.TargetID != nil {
		targetID = textToUUID(*l.TargetID)
	}

	return r.q.RecordAudit(ctx, sqlcgen.RecordAuditParams{
		ID:         textToUUID(auditID),
		OwnerID:    textToUUID(ownerID),
		Action:     l.Action,
		ActorType:  l.ActorType,
		ActorID:    actorID,
		TargetType: l.TargetType,
		TargetID:   targetID,
		Meta:       metaBytes,
		At:         timeToTimestamptz(auditAt),
	})
}

type SystemStateRepository struct {
	q *sqlcgen.Queries
}

func NewSystemStateRepository(q *sqlcgen.Queries) *SystemStateRepository {
	return &SystemStateRepository{q: q}
}

func (r *SystemStateRepository) Get(ctx context.Context) (*system.SystemState, error) {
	row, err := r.q.GetSystemState(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler estado do sistema: %w", err)
	}

	var pausedBy *string
	if row.PausedBy.Valid {
		s := uuidToText(row.PausedBy)
		pausedBy = &s
	}

	var pausedAt = timestamptzToTime(row.PausedAt)

	return &system.SystemState{
		Paused:    row.Paused,
		PausedBy:  pausedBy,
		PausedAt:  &pausedAt,
		Reason:    row.Reason,
		UpdatedAt: timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *SystemStateRepository) Save(ctx context.Context, s *system.SystemState) error {
	var pausedBy pgtype.UUID
	if s.PausedBy != nil {
		pausedBy = textToUUID(*s.PausedBy)
	}

	var pausedAt pgtype.Timestamptz
	if s.PausedAt != nil {
		pausedAt = timeToTimestamptz(*s.PausedAt)
	}

	return r.q.SaveSystemState(ctx, sqlcgen.SaveSystemStateParams{
		Paused:    s.Paused,
		PausedBy:  pausedBy,
		PausedAt:  pausedAt,
		Reason:    s.Reason,
		UpdatedAt: timeToTimestamptz(s.UpdatedAt),
	})
}

type MoneyWaitingReader struct {
	q *sqlcgen.Queries
}

func NewMoneyWaitingReader(q *sqlcgen.Queries) *MoneyWaitingReader {
	return &MoneyWaitingReader{q: q}
}

func (r *MoneyWaitingReader) GetMoneyWaiting(ctx context.Context) (int64, int, error) {
	row, err := r.q.GetMoneyWaiting(ctx)
	if err != nil {
		return 0, 0, err
	}
	return row.TotalCents, int(row.Count), nil
}

// AppRepository implementação
type AppRepository struct {
	q *sqlcgen.Queries
}

func NewAppRepository(q *sqlcgen.Queries) *AppRepository {
	return &AppRepository{q: q}
}

func (r *AppRepository) ListWithFinancials(ctx context.Context) ([]app.AppDTO, error) {
	rows, err := r.q.ListAppsWithFinancials(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar apps: %w", err)
	}

	out := make([]app.AppDTO, len(rows))
	for i, row := range rows {
		var clientID *string
		if row.ClientID.Valid {
			s := uuidToText(row.ClientID)
			clientID = &s
		}
		margin := row.MonthlyRevenueCents - row.MonthlyCostCents
		isCostly := row.MonthlyCostCents > row.MonthlyRevenueCents

		out[i] = app.AppDTO{
			ID:                  uuidToText(row.ID),
			ClientID:            clientID,
			ClientName:          row.ClientName,
			Name:                row.Name,
			Hosting:             row.Hosting,
			Database:            row.Database,
			Repository:          row.Repository,
			Domain:              row.Domain,
			Status:              row.Status,
			MonthlyCostCents:    row.MonthlyCostCents,
			MonthlyRevenueCents: row.MonthlyRevenueCents,
			NetMarginCents:      margin,
			IsCostly:            isCostly,
			CreatedAt:           timestamptzToTime(row.CreatedAt),
			UpdatedAt:           timestamptzToTime(row.UpdatedAt),
		}
	}
	return out, nil
}

func (r *AppRepository) GetByID(ctx context.Context, id string) (*app.AppDTO, error) {
	row, err := r.q.GetAppByID(ctx, textToUUID(id))
	if err != nil {
		return nil, fmt.Errorf("app não encontrado: %w", err)
	}

	var clientID *string
	if row.ClientID.Valid {
		s := uuidToText(row.ClientID)
		clientID = &s
	}

	return &app.AppDTO{
		ID:         uuidToText(row.ID),
		ClientID:   clientID,
		Name:       row.Name,
		Hosting:    row.Hosting,
		Database:   row.Database,
		Repository: row.Repository,
		Domain:     row.Domain,
		Status:     row.Status,
		CreatedAt:  timestamptzToTime(row.CreatedAt),
		UpdatedAt:  timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *AppRepository) Create(ctx context.Context, cmd app.CreateAppCommand) (*app.AppDTO, error) {
	id := uuid.New().String()
	now := time.Now()

	var clientID pgtype.UUID
	if cmd.ClientID != nil && *cmd.ClientID != "" {
		clientID = textToUUID(*cmd.ClientID)
	}

	row, err := r.q.CreateApp(ctx, sqlcgen.CreateAppParams{
		ID:         textToUUID(id),
		OwnerID:    textToUUID("00000000-0000-0000-0000-000000000000"), // auth.uid()
		ClientID:   clientID,
		Name:       cmd.Name,
		Hosting:    cmd.Hosting,
		Database:   cmd.Database,
		Repository: cmd.Repository,
		Domain:     cmd.Domain,
		Status:     "ativo",
		CreatedAt:  timeToTimestamptz(now),
		UpdatedAt:  timeToTimestamptz(now),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar app: %w", err)
	}

	return &app.AppDTO{
		ID:         uuidToText(row.ID),
		Name:       row.Name,
		Hosting:    row.Hosting,
		Database:   row.Database,
		Repository: row.Repository,
		Domain:     row.Domain,
		Status:     row.Status,
		CreatedAt:  timestamptzToTime(row.CreatedAt),
		UpdatedAt:  timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *AppRepository) Update(ctx context.Context, cmd app.UpdateAppCommand) error {
	var clientID pgtype.UUID
	if cmd.ClientID != nil && *cmd.ClientID != "" {
		clientID = textToUUID(*cmd.ClientID)
	}

	return r.q.UpdateApp(ctx, sqlcgen.UpdateAppParams{
		ID:         textToUUID(cmd.ID),
		ClientID:   clientID,
		Name:       cmd.Name,
		Hosting:    cmd.Hosting,
		Database:   cmd.Database,
		Repository: cmd.Repository,
		Domain:     cmd.Domain,
		Status:     cmd.Status,
		UpdatedAt:  timeToTimestamptz(time.Now()),
	})
}

func (r *AppRepository) Delete(ctx context.Context, id string) error {
	return r.q.DeleteApp(ctx, textToUUID(id))
}

// FinanceRepository implementação
type FinanceRepository struct {
	q *sqlcgen.Queries
}

func NewFinanceRepository(q *sqlcgen.Queries) *FinanceRepository {
	return &FinanceRepository{q: q}
}

func (r *FinanceRepository) GetOverview(ctx context.Context) (*app.FinancesOverviewDTO, error) {
	totals, err := r.q.GetFinanceTotals(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler totais financeiros: %w", err)
	}

	services, err := r.q.ListCostByService(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar custos por serviço: %w", err)
	}

	byService := make([]app.CostByServiceDTO, len(services))
	for i, s := range services {
		name := "Outros"
		if s.ServiceName != "" {
			name = s.ServiceName
		}
		byService[i] = app.CostByServiceDTO{
			ServiceName: name,
			CostCents:   s.CostCents,
		}
	}

	entriesRows, err := r.q.ListLedgerEntries(ctx, 20)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar lançamentos: %w", err)
	}

	entries := make([]app.LedgerEntryDTO, len(entriesRows))
	for i, row := range entriesRows {
		var appID, clientID *string
		if row.AppID.Valid {
			s := uuidToText(row.AppID)
			appID = &s
		}
		if row.ClientID.Valid {
			s := uuidToText(row.ClientID)
			clientID = &s
		}
		entries[i] = app.LedgerEntryDTO{
			ID:          uuidToText(row.ID),
			AppID:       appID,
			AppName:     row.AppName,
			ClientID:    clientID,
			ClientName:  row.ClientName,
			Kind:        string(row.Kind),
			AmountCents: row.AmountCents,
			Currency:    row.Currency,
			OccurredOn:  row.OccurredOn.Time.Format("2006-01-02"),
			Service:     row.Service,
			Description: row.Description,
			CreatedAt:   timestamptzToTime(row.CreatedAt),
		}
	}

	return &app.FinancesOverviewDTO{
		TotalRevenueCents: totals.TotalRevenueCents,
		TotalCostCents:    totals.TotalCostCents,
		NetProfitCents:    totals.TotalRevenueCents - totals.TotalCostCents,
		CostsByService:    byService,
		RecentEntries:     entries,
	}, nil
}

func (r *FinanceRepository) CreateEntry(ctx context.Context, cmd app.CreateLedgerEntryCommand) (*app.LedgerEntryDTO, error) {
	id := uuid.New().String()
	now := time.Now()

	var appID pgtype.UUID
	if cmd.AppID != nil && *cmd.AppID != "" {
		appID = textToUUID(*cmd.AppID)
	}

	var clientID pgtype.UUID
	if cmd.ClientID != nil && *cmd.ClientID != "" {
		clientID = textToUUID(*cmd.ClientID)
	}

	occurredDate, err := time.Parse("2006-01-02", cmd.OccurredOn)
	if err != nil {
		occurredDate = now
	}

	row, err := r.q.CreateLedgerEntry(ctx, sqlcgen.CreateLedgerEntryParams{
		ID:          textToUUID(id),
		OwnerID:     textToUUID("00000000-0000-0000-0000-000000000000"),
		AppID:       appID,
		ClientID:    clientID,
		Kind:        sqlcgen.GeevEntryKind(cmd.Kind),
		AmountCents: cmd.AmountCents,
		Currency:    cmd.Currency,
		OccurredOn:  pgtype.Date{Time: occurredDate, Valid: true},
		Service:     cmd.Service,
		Description: cmd.Description,
		CreatedAt:   timeToTimestamptz(now),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao inserir lançamento: %w", err)
	}

	return &app.LedgerEntryDTO{
		ID:          uuidToText(row.ID),
		Kind:        string(row.Kind),
		AmountCents: row.AmountCents,
		Currency:    row.Currency,
		OccurredOn:  row.OccurredOn.Time.Format("2006-01-02"),
		Service:     row.Service,
		Description: row.Description,
		CreatedAt:   timestamptzToTime(row.CreatedAt),
	}, nil
}

// ConnectorRepository implementação
type ConnectorRepository struct {
	q *sqlcgen.Queries
}

func NewConnectorRepository(q *sqlcgen.Queries) *ConnectorRepository {
	return &ConnectorRepository{q: q}
}

var defaultConnectors = []string{
	"GitHub", "Vercel", "Render", "Neon", "Supabase", "Netlify", "Firebase", "Docker Hub", "Provedor de IA (agentes)",
}

func (r *ConnectorRepository) List(ctx context.Context) ([]app.ConnectorDTO, error) {
	rows, err := r.q.ListConnectors(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar conectores: %w", err)
	}

	// Se tabela estiver vazia, retorna a lista padrão fixa desligada (imutabilidade padrão)
	if len(rows) == 0 {
		out := make([]app.ConnectorDTO, len(defaultConnectors))
		for i, name := range defaultConnectors {
			out[i] = app.ConnectorDTO{
				ID:        uuid.NewMD5(uuid.Nil, []byte(name)).String(),
				Name:      name,
				Active:    false,
				SecretRef: "",
				Scope:     "read_only",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}
		return out, nil
	}

	out := make([]app.ConnectorDTO, len(rows))
	for i, row := range rows {
		out[i] = app.ConnectorDTO{
			ID:        uuidToText(row.ID),
			Name:      row.Name,
			Active:    row.Active,
			SecretRef: row.SecretRef,
			Scope:     row.Scope,
			CreatedAt: timestamptzToTime(row.CreatedAt),
			UpdatedAt: timestamptzToTime(row.UpdatedAt),
		}
	}
	return out, nil
}

func (r *ConnectorRepository) Toggle(ctx context.Context, id string, active bool) error {
	// Proibido ligar sem secret_ref por regra de segurança
	return nil
}

// ClientRepository implementação
type ClientRepository struct {
	q *sqlcgen.Queries
}

func NewClientRepository(q *sqlcgen.Queries) *ClientRepository {
	return &ClientRepository{q: q}
}

func (r *ClientRepository) FindByName(ctx context.Context, name string) (*app.ClientDTO, error) {
	row, err := r.q.FindClientByName(ctx, name)
	if err != nil {
		return nil, nil // Não encontrado
	}
	return &app.ClientDTO{
		ID:         uuidToText(row.ID),
		Name:       row.Name,
		ContactRef: row.ContactRef,
		CreatedAt:  timestamptzToTime(row.CreatedAt),
		UpdatedAt:  timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *ClientRepository) GetByID(ctx context.Context, id string) (*app.ClientDTO, error) {
	row, err := r.q.GetClientByID(ctx, textToUUID(id))
	if err != nil {
		return nil, fmt.Errorf("cliente não encontrado: %w", err)
	}
	return &app.ClientDTO{
		ID:         uuidToText(row.ID),
		Name:       row.Name,
		ContactRef: row.ContactRef,
		CreatedAt:  timestamptzToTime(row.CreatedAt),
		UpdatedAt:  timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *ClientRepository) Create(ctx context.Context, name, contactRef string) (*app.ClientDTO, error) {
	now := time.Now()
	newID := uuid.New()
	ownerUUID := textToUUID("00000000-0000-0000-0000-000000000001") // Mock/default owner
	row, err := r.q.CreateClient(ctx, sqlcgen.CreateClientParams{
		ID:         uuidToPgUUID(newID),
		OwnerID:    ownerUUID,
		Name:       name,
		ContactRef: contactRef,
		CreatedAt:  timeToTimestamptz(now),
		UpdatedAt:  timeToTimestamptz(now),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente: %w", err)
	}
	return &app.ClientDTO{
		ID:         uuidToText(row.ID),
		Name:       row.Name,
		ContactRef: row.ContactRef,
		CreatedAt:  timestamptzToTime(row.CreatedAt),
		UpdatedAt:  timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *ClientRepository) Delete(ctx context.Context, id string) error {
	return r.q.DeleteClient(ctx, textToUUID(id))
}

func (r *ClientRepository) Export(ctx context.Context, id string) (*app.ClientExportDTO, error) {
	client, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Busca cards do cliente
	allCards, err := r.q.ListCardsEnriched(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar cartões para exportação: %w", err)
	}
	var clientCards []app.CardListItemDTO
	for _, c := range allCards {
		if uuidToText(c.ClientID) == id {
			clientCards = append(clientCards, app.CardListItemDTO{
				ID:           uuidToText(c.ID),
				ClientID:     uuidToText(c.ClientID),
				ClientName:   c.ClientName,
				AppName:      c.AppName,
				Type:         string(c.Type),
				Title:        c.Title,
				PriceCents:   c.PriceCents,
				EstCostCents: c.EstCostCents,
				Currency:     c.Currency,
				Stage:        string(c.Stage),
				Priority:     c.Priority,
				CreatedAt:    timestamptzToTime(c.CreatedAt),
				UpdatedAt:    timestamptzToTime(c.UpdatedAt),
			})
		}
	}

	return &app.ClientExportDTO{
		Client:   *client,
		Cards:    clientCards,
		Ledger:   []app.LedgerEntryDTO{},
		ExportAt: time.Now(),
	}, nil
}

// ProposalRepository implementação
type ProposalRepository struct {
	q *sqlcgen.Queries
}

func NewProposalRepository(q *sqlcgen.Queries) *ProposalRepository {
	return &ProposalRepository{q: q}
}

func (r *ProposalRepository) Create(ctx context.Context, p *app.ProposalDTO) error {
	pID := uuid.New()
	if p.ID != "" {
		if parsed, err := uuid.Parse(p.ID); err == nil {
			pID = parsed
		}
	}
	p.ID = pID.String()
	now := time.Now()
	ownerUUID := textToUUID("00000000-0000-0000-0000-000000000001")

	_, err := r.q.CreateProposal(ctx, sqlcgen.CreateProposalParams{
		ID:              uuidToPgUUID(pID),
		OwnerID:         ownerUUID,
		CardID:          textToUUID(p.CardID),
		AmountCents:     p.AmountCents,
		Currency:        p.Currency,
		Body:            p.Body,
		AcceptTokenHash: textToPgText(p.AcceptTokenHash),
		Status:          sqlcgen.GeevProposalStatus(p.Status),
		CreatedAt:       timeToTimestamptz(now),
		UpdatedAt:       timeToTimestamptz(now),
	})
	return err
}

func (r *ProposalRepository) GetByTokenHash(ctx context.Context, hash string) (*app.PublicProposalViewDTO, string, error) {
	row, err := r.q.GetProposalByTokenHash(ctx, textToPgText(hash))
	if err != nil {
		return nil, "", fmt.Errorf("proposta não encontrada: %w", err)
	}

	view := &app.PublicProposalViewDTO{
		CardTitle:   row.CardTitle,
		ClientName:  row.ClientName,
		AmountCents: row.AmountCents,
		Currency:    row.Currency,
		Body:        row.Body,
		Status:      string(row.Status),
	}
	return view, uuidToText(row.CardID), nil
}

func (r *ProposalRepository) Accept(ctx context.Context, hash string) error {
	row, err := r.q.GetProposalByTokenHash(ctx, textToPgText(hash))
	if err != nil {
		return fmt.Errorf("proposta inválida para aceite: %w", err)
	}

	return r.q.AcceptProposal(ctx, sqlcgen.AcceptProposalParams{
		ID:         row.ID,
		AcceptedAt: timeToTimestamptz(time.Now()),
	})
}

// AgentRepository implementação
type AgentRepository struct {
	q *sqlcgen.Queries
}

func NewAgentRepository(q *sqlcgen.Queries) *AgentRepository {
	return &AgentRepository{q: q}
}

func (r *AgentRepository) GetByTokenHash(ctx context.Context, hash string) (*system.AgentPrincipal, error) {
	row, err := r.q.GetAgentByTokenHash(ctx, textToPgText(hash))
	if err != nil {
		return nil, fmt.Errorf("agente não encontrado: %w", err)
	}

	return &system.AgentPrincipal{
		ID:              uuidToText(row.ID),
		Role:            row.Role,
		Active:          row.Active,
		SpendLimitCents: row.SpendLimitCents,
	}, nil
}

func (r *AgentRepository) List(ctx context.Context) ([]app.AgentDTO, error) {
	rows, err := r.q.ListAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar agentes: %w", err)
	}

	out := make([]app.AgentDTO, len(rows))
	for i, row := range rows {
		out[i] = app.AgentDTO{
			ID:                uuidToText(row.ID),
			Role:              row.Role,
			Active:            row.Active,
			SpendLimitCents:   row.SpendLimitCents,
			SpendCurrentCents: row.SpendCurrentCents,
			Scope:             row.Scope,
			CreatedAt:         timestamptzToTime(row.CreatedAt),
			UpdatedAt:         timestamptzToTime(row.UpdatedAt),
		}
	}
	return out, nil
}

func (r *AgentRepository) Create(ctx context.Context, role, scope string, spendLimitCents int64, tokenHash string) (*app.AgentDTO, error) {
	now := time.Now()
	newID := uuid.New()
	ownerUUID := textToUUID("00000000-0000-0000-0000-000000000001")

	row, err := r.q.CreateAgent(ctx, sqlcgen.CreateAgentParams{
		ID:                uuidToPgUUID(newID),
		OwnerID:           ownerUUID,
		Role:              role,
		Active:            false, // Invariante: agentes sempre iniciam desligados
		SpendLimitCents:   spendLimitCents,
		SpendCurrentCents: 0,
		Scope:             scope,
		TokenHash:         textToPgText(tokenHash),
		CreatedAt:         timeToTimestamptz(now),
		UpdatedAt:         timeToTimestamptz(now),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao cadastrar agente: %w", err)
	}

	return &app.AgentDTO{
		ID:                uuidToText(row.ID),
		Role:              row.Role,
		Active:            row.Active,
		SpendLimitCents:   row.SpendLimitCents,
		SpendCurrentCents: row.SpendCurrentCents,
		Scope:             row.Scope,
		CreatedAt:         timestamptzToTime(row.CreatedAt),
		UpdatedAt:         timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *AgentRepository) AddSpend(ctx context.Context, id string, cents int64) error {
	return r.q.UpdateAgentSpend(ctx, sqlcgen.UpdateAgentSpendParams{
		ID:                textToUUID(id),
		SpendCurrentCents: cents,
		UpdatedAt:         timeToTimestamptz(time.Now()),
	})
}

func (r *AgentRepository) Toggle(ctx context.Context, id string, active bool) error {
	return r.q.ToggleAgent(ctx, sqlcgen.ToggleAgentParams{
		ID:        textToUUID(id),
		Active:    active,
		UpdatedAt: timeToTimestamptz(time.Now()),
	})
}

// PaymentOrderRepository implementa app.PaymentOrderRepository via sqlc.
type PaymentOrderRepository struct {
	q *sqlcgen.Queries
}

func NewPaymentOrderRepository(q *sqlcgen.Queries) *PaymentOrderRepository {
	return &PaymentOrderRepository{q: q}
}

func (r *PaymentOrderRepository) Create(ctx context.Context, o *app.PaymentOrderDTO) error {
	return r.q.CreatePaymentOrder(ctx, sqlcgen.CreatePaymentOrderParams{
		ID:          textToUUID(o.ID),
		OwnerID:     textToUUID(o.OwnerID),
		CardID:      textToUUID(o.CardID),
		ClientID:    textToUUID(o.ClientID),
		AmountCents: o.AmountCents,
		Currency:    o.Currency,
		PixEmv:      o.PixEMV,
		PixTxid:     o.PixTxID,
		PixKey:      o.PixKey,
		Status:      o.Status,
		CreatedAt:   timeToTimestamptz(o.CreatedAt),
		UpdatedAt:   timeToTimestamptz(o.UpdatedAt),
	})
}

func (r *PaymentOrderRepository) GetByID(ctx context.Context, id string) (*app.PaymentOrderDTO, error) {
	row, err := r.q.GetPaymentOrderByID(ctx, textToUUID(id))
	if err != nil {
		return nil, err
	}
	return mapPaymentOrderRow(row), nil
}

func (r *PaymentOrderRepository) GetByTxID(ctx context.Context, txid string) (*app.PaymentOrderDTO, error) {
	row, err := r.q.GetPaymentOrderByTxID(ctx, txid)
	if err != nil {
		return nil, err
	}
	return mapPaymentOrderRow(row), nil
}

func (r *PaymentOrderRepository) GetByCardID(ctx context.Context, cardID string) (*app.PaymentOrderDTO, error) {
	row, err := r.q.GetPaymentOrderByCardID(ctx, textToUUID(cardID))
	if err != nil {
		return nil, err
	}
	return mapPaymentOrderRow(row), nil
}

func (r *PaymentOrderRepository) UpdateStatus(ctx context.Context, id, status string, paidAt *time.Time) error {
	var paidAtTs pgtype.Timestamptz
	if paidAt != nil {
		paidAtTs = timeToTimestamptz(*paidAt)
	}
	return r.q.UpdatePaymentOrderStatus(ctx, sqlcgen.UpdatePaymentOrderStatusParams{
		ID:        textToUUID(id),
		Status:    status,
		PaidAt:    paidAtTs,
		UpdatedAt: timeToTimestamptz(time.Now()),
	})
}

func mapPaymentOrderRow(row sqlcgen.GeevPaymentOrder) *app.PaymentOrderDTO {
	dto := &app.PaymentOrderDTO{
		ID:          uuidToText(row.ID),
		OwnerID:     uuidToText(row.OwnerID),
		CardID:      uuidToText(row.CardID),
		ClientID:    uuidToText(row.ClientID),
		AmountCents: row.AmountCents,
		Currency:    row.Currency,
		PixEMV:      row.PixEmv,
		PixTxID:     row.PixTxid,
		PixKey:      row.PixKey,
		Status:      row.Status,
		CreatedAt:   timestamptzToTime(row.CreatedAt),
		UpdatedAt:   timestamptzToTime(row.UpdatedAt),
	}
	if row.PaidAt.Valid {
		t := timestamptzToTime(row.PaidAt)
		dto.PaidAt = &t
	}
	return dto
}



