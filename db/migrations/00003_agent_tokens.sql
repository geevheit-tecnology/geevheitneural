-- +goose Up
-- SQL in this section is executed after the migration is applied.

ALTER TABLE geev.agents 
    ADD COLUMN IF NOT EXISTS token_hash TEXT UNIQUE,
    ADD COLUMN IF NOT EXISTS spend_current_cents BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_agents_token_hash ON geev.agents (token_hash) WHERE token_hash IS NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS geev.idx_agents_token_hash;

ALTER TABLE geev.agents 
    DROP COLUMN IF EXISTS spend_current_cents,
    DROP COLUMN IF EXISTS token_hash;
