package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const DefaultOwnerID = "a0000000-0000-0000-0000-000000000001"

// PipelineRepository implementa app.PipelineRepository
type PipelineRepository struct {
	q *sqlcgen.Queries
}

func NewPipelineRepository(q *sqlcgen.Queries) *PipelineRepository {
	return &PipelineRepository{q: q}
}

func (r *PipelineRepository) Upsert(ctx context.Context, p *app.PipelineDTO) error {
	now := time.Now()
	id := p.ID
	if id == "" {
		id = uuid.NewString()
		p.ID = id
	}

	row, err := r.q.UpsertCardPipeline(ctx, sqlcgen.UpsertCardPipelineParams{
		ID:              textToUUID(id),
		OwnerID:         ownerUUID(ctx, DefaultOwnerID),
		CardID:          textToUUID(p.CardID),
		Stage:           p.Stage,
		ActiveRole:      p.ActiveRole,
		SpendTotalCents: p.SpendTotalCents,
		SpendLimitCents: p.SpendLimitCents,
		Status:          p.Status,
		CurrentTask:     p.CurrentTask,
		CreatedAt:       timeToTimestamptz(now),
		UpdatedAt:       timeToTimestamptz(now),
	})
	if err != nil {
		return fmt.Errorf("falha ao salvar pipeline do cartão: %w", err)
	}

	p.CreatedAt = timestamptzToTime(row.CreatedAt)
	p.UpdatedAt = timestamptzToTime(row.UpdatedAt)
	return nil
}

func (r *PipelineRepository) GetByCardID(ctx context.Context, cardID string) (*app.PipelineDTO, error) {
	row, err := r.q.GetCardPipelineByCardID(ctx, textToUUID(cardID))
	if err != nil {
		return nil, fmt.Errorf("pipeline não encontrado para cartão %s: %w", cardID, err)
	}

	return &app.PipelineDTO{
		ID:              uuidToText(row.ID),
		CardID:          uuidToText(row.CardID),
		Stage:           row.Stage,
		ActiveRole:      row.ActiveRole,
		SpendTotalCents: row.SpendTotalCents,
		SpendLimitCents: row.SpendLimitCents,
		Status:          row.Status,
		CurrentTask:     row.CurrentTask,
		CreatedAt:       timestamptzToTime(row.CreatedAt),
		UpdatedAt:       timestamptzToTime(row.UpdatedAt),
	}, nil
}

func (r *PipelineRepository) List(ctx context.Context) ([]app.PipelineDTO, error) {
	rows, err := r.q.ListCardPipelines(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar pipelines: %w", err)
	}

	result := make([]app.PipelineDTO, len(rows))
	for i, row := range rows {
		result[i] = app.PipelineDTO{
			ID:              uuidToText(row.ID),
			CardID:          uuidToText(row.CardID),
			Stage:           row.Stage,
			ActiveRole:      row.ActiveRole,
			SpendTotalCents: row.SpendTotalCents,
			SpendLimitCents: row.SpendLimitCents,
			Status:          row.Status,
			CurrentTask:     row.CurrentTask,
			CreatedAt:       timestamptzToTime(row.CreatedAt),
			UpdatedAt:       timestamptzToTime(row.UpdatedAt),
		}
	}
	return result, nil
}

func (r *PipelineRepository) AddSpend(ctx context.Context, cardID string, cents int64) error {
	return r.q.AddPipelineSpend(ctx, sqlcgen.AddPipelineSpendParams{
		CardID:          textToUUID(cardID),
		SpendTotalCents: cents,
		UpdatedAt:       timeToTimestamptz(time.Now()),
	})
}

// SpecificationRepository implementa app.SpecificationRepository
type SpecificationRepository struct {
	q *sqlcgen.Queries
}

func NewSpecificationRepository(q *sqlcgen.Queries) *SpecificationRepository {
	return &SpecificationRepository{q: q}
}

func (r *SpecificationRepository) Upsert(ctx context.Context, spec *app.SpecificationDTO, tokenHash string) error {
	now := time.Now()
	id := spec.ID
	if id == "" {
		id = uuid.NewString()
		spec.ID = id
	}

	userStoriesJSON, err := json.Marshal(spec.UserStories)
	if err != nil {
		return fmt.Errorf("falha ao serializar user stories: %w", err)
	}

	var hashPg pgtype.Text
	if tokenHash != "" {
		hashPg = pgtype.Text{String: tokenHash, Valid: true}
	}

	row, err := r.q.UpsertSpecification(ctx, sqlcgen.UpsertSpecificationParams{
		ID:                 textToUUID(id),
		OwnerID:            ownerUUID(ctx, DefaultOwnerID),
		CardID:             textToUUID(spec.CardID),
		UserStories:        userStoriesJSON,
		TestsCode:          spec.TestsCode,
		TestsPassed:        spec.TestsPassed,
		ClientApproved:     spec.ClientApproved,
		ClientApprovedAt:   pgtype.Timestamptz{Valid: false},
		PaymentSignalCents: spec.PaymentSignalCents,
		PaymentSignalPaid:  spec.PaymentSignalPaid,
		ScopeTokenHash:     hashPg,
		CreatedAt:          timeToTimestamptz(now),
		UpdatedAt:          timeToTimestamptz(now),
	})
	if err != nil {
		return fmt.Errorf("falha ao salvar especificação: %w", err)
	}

	spec.ID = uuidToText(row.ID)
	return nil
}

func (r *SpecificationRepository) GetByCardID(ctx context.Context, cardID string) (*app.SpecificationDTO, error) {
	row, err := r.q.GetSpecificationByCardID(ctx, textToUUID(cardID))
	if err != nil {
		return nil, fmt.Errorf("especificação não encontrada para o cartão %s: %w", cardID, err)
	}

	var stories []map[string]any
	if len(row.UserStories) > 0 {
		_ = json.Unmarshal(row.UserStories, &stories)
	}

	return &app.SpecificationDTO{
		ID:                 uuidToText(row.ID),
		CardID:             uuidToText(row.CardID),
		UserStories:        stories,
		TestsCode:          row.TestsCode,
		TestsPassed:        row.TestsPassed,
		ClientApproved:     row.ClientApproved,
		PaymentSignalCents: row.PaymentSignalCents,
		PaymentSignalPaid:  row.PaymentSignalPaid,
	}, nil
}

func (r *SpecificationRepository) GetByTokenHash(ctx context.Context, hash string) (*app.SpecificationDTO, error) {
	row, err := r.q.GetSpecificationByScopeTokenHash(ctx, pgtype.Text{String: hash, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("especificação não encontrada por token: %w", err)
	}

	var stories []map[string]any
	if len(row.UserStories) > 0 {
		_ = json.Unmarshal(row.UserStories, &stories)
	}

	return &app.SpecificationDTO{
		ID:                 uuidToText(row.ID),
		CardID:             uuidToText(row.CardID),
		UserStories:        stories,
		TestsCode:          row.TestsCode,
		TestsPassed:        row.TestsPassed,
		ClientApproved:     row.ClientApproved,
		PaymentSignalCents: row.PaymentSignalCents,
		PaymentSignalPaid:  row.PaymentSignalPaid,
	}, nil
}

func (r *SpecificationRepository) ApproveByClient(ctx context.Context, hash string) error {
	return r.q.ApproveSpecificationByClient(ctx, sqlcgen.ApproveSpecificationByClientParams{
		ScopeTokenHash:   pgtype.Text{String: hash, Valid: true},
		ClientApprovedAt: timeToTimestamptz(time.Now()),
	})
}

// PlaybookRepository implementa app.PlaybookRepository
type PlaybookRepository struct {
	q *sqlcgen.Queries
}

func NewPlaybookRepository(q *sqlcgen.Queries) *PlaybookRepository {
	return &PlaybookRepository{q: q}
}

func (r *PlaybookRepository) ListPlaybooks(ctx context.Context) ([]app.PlaybookDTO, error) {
	rows, err := r.q.ListPlaybooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar playbooks: %w", err)
	}

	result := make([]app.PlaybookDTO, len(rows))
	for i, row := range rows {
		var stack map[string]any
		if len(row.Stack) > 0 {
			_ = json.Unmarshal(row.Stack, &stack)
		}
		var checklist []string
		if len(row.Checklist) > 0 {
			_ = json.Unmarshal(row.Checklist, &checklist)
		}

		result[i] = app.PlaybookDTO{
			ProductType:   row.ProductType,
			Title:         row.Title,
			Description:   row.Description,
			Stack:         stack,
			Checklist:     checklist,
			AutonomyLevel: int(row.AutonomyLevel),
		}
	}
	return result, nil
}

func (r *PlaybookRepository) GetPlaybook(ctx context.Context, productType string) (*app.PlaybookDTO, error) {
	row, err := r.q.GetPlaybookByType(ctx, productType)
	if err != nil {
		return nil, fmt.Errorf("playbook não encontrado para %s: %w", productType, err)
	}

	var stack map[string]any
	if len(row.Stack) > 0 {
		_ = json.Unmarshal(row.Stack, &stack)
	}
	var checklist []string
	if len(row.Checklist) > 0 {
		_ = json.Unmarshal(row.Checklist, &checklist)
	}

	return &app.PlaybookDTO{
		ProductType:   row.ProductType,
		Title:         row.Title,
		Description:   row.Description,
		Stack:         stack,
		Checklist:     checklist,
		AutonomyLevel: int(row.AutonomyLevel),
	}, nil
}

func (r *PlaybookRepository) UpsertPlaybook(ctx context.Context, p *app.PlaybookDTO) error {
	now := time.Now()
	stackJSON, _ := json.Marshal(p.Stack)
	checkJSON, _ := json.Marshal(p.Checklist)

	_, err := r.q.UpsertPlaybook(ctx, sqlcgen.UpsertPlaybookParams{
		ID:            textToUUID(uuid.NewString()),
		OwnerID:       ownerUUID(ctx, DefaultOwnerID),
		ProductType:   p.ProductType,
		Title:         p.Title,
		Description:   p.Description,
		Stack:         stackJSON,
		Checklist:     checkJSON,
		Prompts:       []byte(`{}`),
		AutonomyLevel: int32(p.AutonomyLevel),
		CreatedAt:     timeToTimestamptz(now),
		UpdatedAt:     timeToTimestamptz(now),
	})
	return err
}

func (r *PlaybookRepository) ListModules(ctx context.Context) ([]app.ModuleDTO, error) {
	rows, err := r.q.ListModules(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar módulos: %w", err)
	}

	result := make([]app.ModuleDTO, len(rows))
	for i, row := range rows {
		var templates map[string]any
		if len(row.CodeTemplates) > 0 {
			_ = json.Unmarshal(row.CodeTemplates, &templates)
		}

		result[i] = app.ModuleDTO{
			Name:          row.Name,
			Category:      row.Category,
			Description:   row.Description,
			CodeTemplates: templates,
			IsReviewed:    row.IsReviewed,
		}
	}
	return result, nil
}

func (r *PlaybookRepository) UpsertModule(ctx context.Context, m *app.ModuleDTO) error {
	now := time.Now()
	tplJSON, _ := json.Marshal(m.CodeTemplates)

	_, err := r.q.UpsertModule(ctx, sqlcgen.UpsertModuleParams{
		ID:            textToUUID(uuid.NewString()),
		OwnerID:       ownerUUID(ctx, DefaultOwnerID),
		Name:          m.Name,
		Category:      m.Category,
		Description:   m.Description,
		CodeTemplates: tplJSON,
		IsReviewed:    m.IsReviewed,
		CreatedAt:     timeToTimestamptz(now),
		UpdatedAt:     timeToTimestamptz(now),
	})
	return err
}

// PreviewRepository implementa app.PreviewRepository
type PreviewRepository struct {
	q *sqlcgen.Queries
}

func NewPreviewRepository(q *sqlcgen.Queries) *PreviewRepository {
	return &PreviewRepository{q: q}
}

func (r *PreviewRepository) Create(ctx context.Context, prev *app.PreviewDTO) error {
	now := time.Now()
	id := uuid.NewString()
	rollbackJSON, _ := json.Marshal(prev.RollbackPlan)

	row, err := r.q.CreatePreview(ctx, sqlcgen.CreatePreviewParams{
		ID:           textToUUID(id),
		OwnerID:      ownerUUID(ctx, DefaultOwnerID),
		CardID:       textToUUID(prev.CardID),
		PreviewUrl:   prev.PreviewURL,
		Status:       prev.Status,
		ExpiresAt:    timeToTimestamptz(prev.ExpiresAt),
		RollbackPlan: rollbackJSON,
		CreatedAt:    timeToTimestamptz(now),
		UpdatedAt:    timeToTimestamptz(now),
	})
	if err != nil {
		return fmt.Errorf("falha ao registrar preview: %w", err)
	}

	prev.ExpiresAt = timestamptzToTime(row.ExpiresAt)
	return nil
}

func (r *PreviewRepository) GetByCardID(ctx context.Context, cardID string) (*app.PreviewDTO, error) {
	row, err := r.q.GetPreviewByCardID(ctx, textToUUID(cardID))
	if err != nil {
		return nil, fmt.Errorf("preview não encontrado para cartão %s: %w", cardID, err)
	}

	var rollback map[string]any
	if len(row.RollbackPlan) > 0 {
		_ = json.Unmarshal(row.RollbackPlan, &rollback)
	}

	return &app.PreviewDTO{
		CardID:       uuidToText(row.CardID),
		PreviewURL:   row.PreviewUrl,
		Status:       row.Status,
		ExpiresAt:    timestamptzToTime(row.ExpiresAt),
		RollbackPlan: rollback,
	}, nil
}

// FactoryMetricsRepository implementa app.FactoryMetricsRepository
type FactoryMetricsRepository struct {
	q *sqlcgen.Queries
}

func NewFactoryMetricsRepository(q *sqlcgen.Queries) *FactoryMetricsRepository {
	return &FactoryMetricsRepository{q: q}
}

func (r *FactoryMetricsRepository) Record(ctx context.Context, cardID string, leadTimeSec, aiCost, infraCost, margin, decisionTimeSec int64, firstTimeApproved bool, defects int) error {
	now := time.Now()
	_, err := r.q.CreateFactoryMetric(ctx, sqlcgen.CreateFactoryMetricParams{
		ID:                       textToUUID(uuid.NewString()),
		OwnerID:                  ownerUUID(ctx, DefaultOwnerID),
		CardID:                   textToUUID(cardID),
		LeadTimeSeconds:          leadTimeSec,
		AiCostCents:              aiCost,
		InfraCostCents:           infraCost,
		FirstTimeApproval:        firstTimeApproved,
		PostDeliveryDefects:      int32(defects),
		OwnerDecisionTimeSeconds: decisionTimeSec,
		MarginCents:              margin,
		CreatedAt:                timeToTimestamptz(now),
	})
	return err
}

func (r *FactoryMetricsRepository) GetScoreboard(ctx context.Context) (*app.FactoryScoreboardDTO, error) {
	row, err := r.q.GetFactoryScoreboard(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar placar da fábrica: %w", err)
	}

	firstTimePct := 0.0
	if row.TotalProjects > 0 {
		firstTimePct = (float64(row.FirstTimeApprovals) / float64(row.TotalProjects)) * 100.0
	}

	return &app.FactoryScoreboardDTO{
		AvgLeadTimeHours:     float64(row.AvgLeadTimeSeconds) / 3600.0,
		TotalAICostCents:     row.TotalAiCostCents,
		TotalInfraCostCents:  row.TotalInfraCostCents,
		TotalMarginCents:     row.TotalMarginCents,
		AvgDecisionTimeHours: float64(row.AvgDecisionTimeSeconds) / 3600.0,
		TotalProjects:        row.TotalProjects,
		FirstTimeApprovalPct: firstTimePct,
		TotalDefects:         row.TotalDefects,
	}, nil
}

// ProjectMemoryRepository implementa app.ProjectMemoryRepository
type ProjectMemoryRepository struct {
	q *sqlcgen.Queries
}

func NewProjectMemoryRepository(q *sqlcgen.Queries) *ProjectMemoryRepository {
	return &ProjectMemoryRepository{q: q}
}

func (r *ProjectMemoryRepository) Upsert(ctx context.Context, mem *app.ProjectMemoryDTO) error {
	now := time.Now()
	id := mem.ID
	if id == "" {
		id = uuid.NewString()
		mem.ID = id
	}
	cat := mem.Category
	if cat == "" {
		cat = "nota"
	}
	return r.q.UpsertProjectMemory(ctx, sqlcgen.UpsertProjectMemoryParams{
		ID:        textToUUID(id),
		OwnerID:   ownerUUID(ctx, DefaultOwnerID),
		CardID:    textToUUID(mem.CardID),
		ClientID:  textToUUID(mem.ClientID),
		Key:       mem.Key,
		Category:  cat,
		Content:   mem.Content,
		CreatedAt: timeToTimestamptz(now),
	})
}

func (r *ProjectMemoryRepository) ListByCardID(ctx context.Context, cardID string) ([]app.ProjectMemoryDTO, error) {
	rows, err := r.q.ListProjectMemoryByCardID(ctx, textToUUID(cardID))
	if err != nil {
		return nil, fmt.Errorf("falha ao listar memória do projeto: %w", err)
	}
	out := make([]app.ProjectMemoryDTO, len(rows))
	for i, row := range rows {
		out[i] = app.ProjectMemoryDTO{
			ID:        uuidToText(row.ID),
			CardID:    uuidToText(row.CardID),
			ClientID:  uuidToText(row.ClientID),
			Key:       row.Key,
			Category:  row.Category,
			Content:   row.Content,
			CreatedAt: timestamptzToTime(row.CreatedAt),
		}
	}
	return out, nil
}
