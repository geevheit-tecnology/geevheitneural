package postgres

import (
	"context"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/domain/portal"
	"github.com/jackc/pgx/v5/pgtype"
)

type PortalTokenRepository struct {
	q *sqlcgen.Queries
}

func NewPortalTokenRepository(q *sqlcgen.Queries) *PortalTokenRepository {
	return &PortalTokenRepository{q: q}
}

func (r *PortalTokenRepository) Create(ctx context.Context, t *portal.PortalToken) error {
	return r.q.CreatePortalToken(ctx, sqlcgen.CreatePortalTokenParams{
		ID:        textToUUID(t.ID),
		OwnerID:   textToUUID(t.OwnerID),
		CardID:    textToUUID(t.CardID),
		ClientID:  textToUUID(t.ClientID),
		TokenHash: t.TokenHash,
		ExpiresAt: timeToTimestamptz(t.ExpiresAt),
		CreatedAt: timeToTimestamptz(t.CreatedAt),
		UpdatedAt: timeToTimestamptz(t.UpdatedAt),
	})
}

func (r *PortalTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*portal.PortalToken, error) {
	row, err := r.q.GetPortalTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return mapPortalTokenRow(row), nil
}

func (r *PortalTokenRepository) GetByCardID(ctx context.Context, cardID string) (*portal.PortalToken, error) {
	row, err := r.q.GetPortalTokenByCardID(ctx, textToUUID(cardID))
	if err != nil {
		return nil, err
	}
	return mapPortalTokenRow(row), nil
}

func (r *PortalTokenRepository) Update(ctx context.Context, t *portal.PortalToken) error {
	var revokedAt pgtype.Timestamptz
	if t.RevokedAt != nil {
		revokedAt = timeToTimestamptz(*t.RevokedAt)
	}

	var lastAccessedAt pgtype.Timestamptz
	if t.LastAccessedAt != nil {
		lastAccessedAt = timeToTimestamptz(*t.LastAccessedAt)
	}

	var acceptedAt pgtype.Timestamptz
	if t.AcceptedAt != nil {
		acceptedAt = timeToTimestamptz(*t.AcceptedAt)
	}

	return r.q.UpdatePortalToken(ctx, sqlcgen.UpdatePortalTokenParams{
		ID:                textToUUID(t.ID),
		RevokedAt:         revokedAt,
		LastAccessedAt:    lastAccessedAt,
		AcceptedAt:        acceptedAt,
		AcceptedBy:        textToPgText(t.AcceptedBy),
		AcceptedIp:        textToPgText(t.AcceptedIP),
		AcceptedUserAgent: textToPgText(t.AcceptedUserAgent),
		AcceptedStateHash: textToPgText(t.AcceptedStateHash),
		UpdatedAt:         timeToTimestamptz(t.UpdatedAt),
	})
}

func (r *PortalTokenRepository) RevokeByCardID(ctx context.Context, cardID string, revokedAt time.Time) error {
	return r.q.RevokePortalTokensByCardID(ctx, sqlcgen.RevokePortalTokensByCardIDParams{
		CardID:    textToUUID(cardID),
		RevokedAt: timeToTimestamptz(revokedAt),
	})
}

func mapPortalTokenRow(row sqlcgen.GetPortalTokenByHashRow) *portal.PortalToken {
	t := &portal.PortalToken{
		ID:        uuidToText(row.ID),
		OwnerID:   uuidToText(row.OwnerID),
		CardID:    uuidToText(row.CardID),
		ClientID:  uuidToText(row.ClientID),
		TokenHash: row.TokenHash,
		ExpiresAt: timestamptzToTime(row.ExpiresAt),
		CreatedAt: timestamptzToTime(row.CreatedAt),
		UpdatedAt: timestamptzToTime(row.UpdatedAt),
	}

	if row.RevokedAt.Valid {
		rev := row.RevokedAt.Time
		t.RevokedAt = &rev
	}
	if row.LastAccessedAt.Valid {
		acc := row.LastAccessedAt.Time
		t.LastAccessedAt = &acc
	}
	if row.AcceptedAt.Valid {
		acpt := row.AcceptedAt.Time
		t.AcceptedAt = &acpt
	}
	if row.AcceptedBy.Valid {
		t.AcceptedBy = row.AcceptedBy.String
	}
	if row.AcceptedIp.Valid {
		t.AcceptedIP = row.AcceptedIp.String
	}
	if row.AcceptedUserAgent.Valid {
		t.AcceptedUserAgent = row.AcceptedUserAgent.String
	}
	if row.AcceptedStateHash.Valid {
		t.AcceptedStateHash = row.AcceptedStateHash.String
	}

	return t
}
