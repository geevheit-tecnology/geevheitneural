-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS geev.client_portal_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT 'a0000000-0000-0000-0000-000000000001',
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES geev.clients(id),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    last_accessed_at TIMESTAMPTZ,
    accepted_at TIMESTAMPTZ,
    accepted_by TEXT,
    accepted_ip TEXT,
    accepted_user_agent TEXT,
    accepted_state_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER update_client_portal_tokens_updated_at
BEFORE UPDATE ON geev.client_portal_tokens
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

ALTER TABLE geev.client_portal_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.client_portal_tokens FORCE ROW LEVEL SECURITY;

CREATE POLICY client_portal_tokens_owner ON geev.client_portal_tokens
    FOR ALL TO authenticated
    USING (geev.is_owner());

CREATE POLICY client_portal_tokens_api ON geev.client_portal_tokens
    FOR ALL TO geev_api
    USING (true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS geev.client_portal_tokens CASCADE;
-- +goose StatementEnd
