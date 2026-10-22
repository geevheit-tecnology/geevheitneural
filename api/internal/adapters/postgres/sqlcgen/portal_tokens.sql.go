package sqlcgen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPortalToken = `-- name: CreatePortalToken :exec
INSERT INTO geev.client_portal_tokens (
    id, owner_id, card_id, client_id, token_hash, expires_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
`

type CreatePortalTokenParams struct {
	ID        pgtype.UUID
	OwnerID   pgtype.UUID
	CardID    pgtype.UUID
	ClientID  pgtype.UUID
	TokenHash string
	ExpiresAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (q *Queries) CreatePortalToken(ctx context.Context, arg CreatePortalTokenParams) error {
	_, err := q.db.Exec(ctx, createPortalToken,
		arg.ID,
		arg.OwnerID,
		arg.CardID,
		arg.ClientID,
		arg.TokenHash,
		arg.ExpiresAt,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const getPortalTokenByHash = `-- name: GetPortalTokenByHash :one
SELECT id, owner_id, card_id, client_id, token_hash, expires_at, revoked_at, last_accessed_at, accepted_at, accepted_by, accepted_ip, accepted_user_agent, accepted_state_hash, created_at, updated_at
FROM geev.client_portal_tokens
WHERE token_hash = $1
`

type GetPortalTokenByHashRow struct {
	ID                pgtype.UUID
	OwnerID           pgtype.UUID
	CardID            pgtype.UUID
	ClientID          pgtype.UUID
	TokenHash         string
	ExpiresAt         pgtype.Timestamptz
	RevokedAt         pgtype.Timestamptz
	LastAccessedAt    pgtype.Timestamptz
	AcceptedAt        pgtype.Timestamptz
	AcceptedBy        pgtype.Text
	AcceptedIp        pgtype.Text
	AcceptedUserAgent pgtype.Text
	AcceptedStateHash pgtype.Text
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
}

func (q *Queries) GetPortalTokenByHash(ctx context.Context, tokenHash string) (GetPortalTokenByHashRow, error) {
	row := q.db.QueryRow(ctx, getPortalTokenByHash, tokenHash)
	var i GetPortalTokenByHashRow
	err := row.Scan(
		&i.ID,
		&i.OwnerID,
		&i.CardID,
		&i.ClientID,
		&i.TokenHash,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.LastAccessedAt,
		&i.AcceptedAt,
		&i.AcceptedBy,
		&i.AcceptedIp,
		&i.AcceptedUserAgent,
		&i.AcceptedStateHash,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPortalTokenByCardID = `-- name: GetPortalTokenByCardID :one
SELECT id, owner_id, card_id, client_id, token_hash, expires_at, revoked_at, last_accessed_at, accepted_at, accepted_by, accepted_ip, accepted_user_agent, accepted_state_hash, created_at, updated_at
FROM geev.client_portal_tokens
WHERE card_id = $1
ORDER BY created_at DESC
LIMIT 1
`

func (q *Queries) GetPortalTokenByCardID(ctx context.Context, cardID pgtype.UUID) (GetPortalTokenByHashRow, error) {
	row := q.db.QueryRow(ctx, getPortalTokenByCardID, cardID)
	var i GetPortalTokenByHashRow
	err := row.Scan(
		&i.ID,
		&i.OwnerID,
		&i.CardID,
		&i.ClientID,
		&i.TokenHash,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.LastAccessedAt,
		&i.AcceptedAt,
		&i.AcceptedBy,
		&i.AcceptedIp,
		&i.AcceptedUserAgent,
		&i.AcceptedStateHash,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updatePortalToken = `-- name: UpdatePortalToken :exec
UPDATE geev.client_portal_tokens
SET revoked_at = $2, last_accessed_at = $3, accepted_at = $4, accepted_by = $5, accepted_ip = $6, accepted_user_agent = $7, accepted_state_hash = $8, updated_at = $9
WHERE id = $1
`

type UpdatePortalTokenParams struct {
	ID                pgtype.UUID
	RevokedAt         pgtype.Timestamptz
	LastAccessedAt    pgtype.Timestamptz
	AcceptedAt        pgtype.Timestamptz
	AcceptedBy        pgtype.Text
	AcceptedIp        pgtype.Text
	AcceptedUserAgent pgtype.Text
	AcceptedStateHash pgtype.Text
	UpdatedAt         pgtype.Timestamptz
}

func (q *Queries) UpdatePortalToken(ctx context.Context, arg UpdatePortalTokenParams) error {
	_, err := q.db.Exec(ctx, updatePortalToken,
		arg.ID,
		arg.RevokedAt,
		arg.LastAccessedAt,
		arg.AcceptedAt,
		arg.AcceptedBy,
		arg.AcceptedIp,
		arg.AcceptedUserAgent,
		arg.AcceptedStateHash,
		arg.UpdatedAt,
	)
	return err
}

const revokePortalTokensByCardID = `-- name: RevokePortalTokensByCardID :exec
UPDATE geev.client_portal_tokens
SET revoked_at = $2, updated_at = $2
WHERE card_id = $1 AND revoked_at IS NULL
`

type RevokePortalTokensByCardIDParams struct {
	CardID    pgtype.UUID
	RevokedAt pgtype.Timestamptz
}

func (q *Queries) RevokePortalTokensByCardID(ctx context.Context, arg RevokePortalTokensByCardIDParams) error {
	_, err := q.db.Exec(ctx, revokePortalTokensByCardID, arg.CardID, arg.RevokedAt)
	return err
}
