package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/jackc/pgx/v5"
)

type DurableFactoryRepository struct{ q *sqlcgen.Queries }

func NewDurableFactoryRepository(q *sqlcgen.Queries) *DurableFactoryRepository {
	return &DurableFactoryRepository{q: q}
}

func (r *DurableFactoryRepository) Queue(ctx context.Context, cardID, key string) (*app.DurableFactoryJob, error) {
	row, err := r.q.QueueFactoryJob(ctx, sqlcgen.QueueFactoryJobParams{CardID: textToUUID(cardID), IdempotencyKey: textToUUID(key)})
	if errors.Is(err, pgx.ErrNoRows) {
		existing, lookupErr := r.q.GetFactoryJobByKey(ctx, textToUUID(key))
		if lookupErr != nil {
			return nil, lookupErr
		}
		if uuidToText(existing.CardID) != cardID {
			return nil, fmt.Errorf("chave de idempotência já utilizada para outro cartão")
		}
		return &app.DurableFactoryJob{ID: uuidToText(existing.ID), CardID: uuidToText(existing.CardID), Status: existing.Status, Attempts: existing.Attempts}, nil
	}
	if err != nil {
		return nil, err
	}
	return &app.DurableFactoryJob{ID: uuidToText(row.ID), CardID: uuidToText(row.CardID), Status: row.Status, Attempts: row.Attempts, Created: true}, nil
}

func (r *DurableFactoryRepository) Claim(ctx context.Context, token string) (*app.DurableFactoryJob, error) {
	row, err := r.q.ClaimFactoryJob(ctx, textToUUID(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app.DurableFactoryJob{ID: uuidToText(row.ID), CardID: uuidToText(row.CardID), Status: "running", LeaseToken: uuidToText(row.LeaseToken), LeaseExpiresAt: timestamptzToTime(row.LeaseExpiresAt), Attempts: row.Attempts}, nil
}

func (r *DurableFactoryRepository) ClaimByID(ctx context.Context, id, token string) (*app.DurableFactoryJob, error) {
	row, err := r.q.ClaimFactoryJobByID(ctx, sqlcgen.ClaimFactoryJobByIDParams{ID: textToUUID(id), LeaseToken: textToUUID(token)})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app.DurableFactoryJob{ID: uuidToText(row.ID), CardID: uuidToText(row.CardID), Status: "running", LeaseToken: uuidToText(row.LeaseToken), LeaseExpiresAt: timestamptzToTime(row.LeaseExpiresAt), Attempts: row.Attempts}, nil
}

func (r *DurableFactoryRepository) GetByID(ctx context.Context, id string) (*app.DurableFactoryJob, error) {
	row, err := r.q.GetFactoryJobByID(ctx, textToUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app.DurableFactoryJob{ID: uuidToText(row.ID), CardID: uuidToText(row.CardID), Status: row.Status, LeaseExpiresAt: timestamptzToTime(row.LeaseExpiresAt), Attempts: row.Attempts}, nil
}

func (r *DurableFactoryRepository) ListQueuedIDs(ctx context.Context, after string, limit int32) ([]string, error) {
	ids, err := r.q.ListQueuedFactoryJobIDs(ctx, sqlcgen.ListQueuedFactoryJobIDsParams{ID: textToUUID(after), Limit: limit})
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, uuidToText(id))
	}
	return result, nil
}

func (r *DurableFactoryRepository) Finish(ctx context.Context, id, token, status string) (bool, error) {
	count, err := r.q.FinishFactoryJob(ctx, sqlcgen.FinishFactoryJobParams{ID: textToUUID(id), LeaseToken: textToUUID(token), Status: status})
	return count == 1, err
}

func (r *DurableFactoryRepository) MarkActivePipelineFailed(ctx context.Context, cardID string) error {
	return r.q.MarkActiveFactoryPipelineFailed(ctx, textToUUID(cardID))
}

func (r *DurableFactoryRepository) LockValidLease(ctx context.Context, id, token string) (bool, error) {
	_, err := r.q.LockValidFactoryJob(ctx, sqlcgen.LockValidFactoryJobParams{ID: textToUUID(id), LeaseToken: textToUUID(token)})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *DurableFactoryRepository) MarkExpired(ctx context.Context) (int64, error) {
	return r.q.MarkExpiredFactoryJobs(ctx)
}

func (r *DurableFactoryRepository) MarkExpiredForCard(ctx context.Context, cardID string) (int64, error) {
	return r.q.MarkExpiredFactoryJobForCard(ctx, textToUUID(cardID))
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

// ReplaceFiles deve ser chamado dentro da transação do UnitOfWork.
func (r *DurableFactoryRepository) ReplaceFiles(ctx context.Context, cardID string, files []app.ProjectFile) error {
	if len(files) == 0 {
		return fmt.Errorf("workspace sem arquivos")
	}
	seen := make(map[string]bool, len(files))
	for _, file := range files {
		if seen[file.Path] {
			return fmt.Errorf("arquivo duplicado no workspace: %s", file.Path)
		}
		seen[file.Path] = true
	}
	if _, err := r.q.LockFactoryWorkspace(ctx, textToUUID(cardID)); err != nil {
		return err
	}
	if err := r.q.DeleteFactoryWorkspaceFiles(ctx, textToUUID(cardID)); err != nil {
		return err
	}
	for _, file := range files {
		if err := r.PutFile(ctx, cardID, file); err != nil {
			return err
		}
	}
	return nil
}
