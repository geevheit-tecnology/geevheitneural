package postgres

import (
	"context"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type DurableFactoryRepository struct{ q *sqlcgen.Queries }

func NewDurableFactoryRepository(q *sqlcgen.Queries) *DurableFactoryRepository {
	return &DurableFactoryRepository{q: q}
}

func (r *DurableFactoryRepository) Queue(ctx context.Context, cardID, key string) (*app.DurableFactoryJob, error) {
	row, err := r.q.QueueFactoryJob(ctx, sqlcgen.QueueFactoryJobParams{CardID: textToUUID(cardID), IdempotencyKey: textToUUID(key)})
	if err != nil {
		return nil, err
	}
	return &app.DurableFactoryJob{ID: uuidToText(row.ID), CardID: uuidToText(row.CardID), Status: row.Status, Attempts: row.Attempts}, nil
}

func (r *DurableFactoryRepository) Claim(ctx context.Context, token string) (*app.DurableFactoryJob, error) {
	row, err := r.q.ClaimFactoryJob(ctx, textToUUID(token))
	if err != nil {
		return nil, err
	}
	return &app.DurableFactoryJob{ID: uuidToText(row.ID), CardID: uuidToText(row.CardID), Status: "running", LeaseToken: uuidToText(row.LeaseToken), LeaseExpiresAt: timestamptzToTime(row.LeaseExpiresAt), Attempts: row.Attempts}, nil
}

func (r *DurableFactoryRepository) Finish(ctx context.Context, id, token, status string) (bool, error) {
	count, err := r.q.FinishFactoryJob(ctx, sqlcgen.FinishFactoryJobParams{ID: textToUUID(id), LeaseToken: textToUUID(token), Status: status})
	return count == 1, err
}

func (r *DurableFactoryRepository) MarkExpired(ctx context.Context) (int64, error) {
	return r.q.MarkExpiredFactoryJobs(ctx)
}

func (r *DurableFactoryRepository) PutFile(ctx context.Context, cardID string, file app.ProjectFile) error {
	return r.q.PutFactoryWorkspaceFile(ctx, sqlcgen.PutFactoryWorkspaceFileParams{CardID: textToUUID(cardID), Path: file.Path, Content: file.Content})
}

func (r *DurableFactoryRepository) ListFiles(ctx context.Context, cardID string) ([]app.ProjectFile, error) {
	rows, err := r.q.ListFactoryWorkspaceFiles(ctx, textToUUID(cardID))
	if err != nil {
		return nil, err
	}
	files := make([]app.ProjectFile, 0, len(rows))
	for _, row := range rows {
		files = append(files, app.ProjectFile{Path: row.Path, Content: row.Content})
	}
	return files, nil
}
