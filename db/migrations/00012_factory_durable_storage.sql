-- +goose Up
-- Relacionamentos compostos impedem associar uma execução ou arquivo a cartão de outro dono.
CREATE UNIQUE INDEX uq_cards_owner_id_id ON geev.cards (owner_id, id);

CREATE TABLE geev.factory_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL,
    idempotency_key UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'completed', 'failed', 'interrupted')),
    lease_token UUID,
    lease_expires_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT fk_factory_job_card_owner FOREIGN KEY (owner_id, card_id) REFERENCES geev.cards (owner_id, id) ON DELETE CASCADE,
    CONSTRAINT uq_factory_job_request UNIQUE (owner_id, idempotency_key),
    CONSTRAINT factory_job_lease_consistent CHECK (
        (status = 'running' AND lease_token IS NOT NULL AND lease_expires_at IS NOT NULL)
        OR (status <> 'running' AND lease_token IS NULL AND lease_expires_at IS NULL)
    )
);
CREATE UNIQUE INDEX uq_factory_job_active_card ON geev.factory_jobs (owner_id, card_id) WHERE status IN ('queued', 'running', 'interrupted');
CREATE INDEX ix_factory_jobs_queue ON geev.factory_jobs (owner_id, created_at) WHERE status = 'queued';

CREATE TABLE geev.factory_workspace_files (
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL,
    path TEXT NOT NULL CHECK (path <> '' AND octet_length(path) <= 1024 AND path !~ '(^/|\\|(^|/)\.\.(/|$))'),
    content TEXT NOT NULL CHECK (octet_length(content) <= 1048576),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (owner_id, card_id, path),
    CONSTRAINT fk_factory_file_card_owner FOREIGN KEY (owner_id, card_id) REFERENCES geev.cards (owner_id, id) ON DELETE CASCADE
);

ALTER TABLE geev.factory_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.factory_jobs FORCE ROW LEVEL SECURITY;
ALTER TABLE geev.factory_workspace_files ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.factory_workspace_files FORCE ROW LEVEL SECURITY;
CREATE POLICY p_owner_factory_jobs ON geev.factory_jobs FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());
CREATE POLICY p_owner_factory_workspace_files ON geev.factory_workspace_files FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());
GRANT SELECT, INSERT, UPDATE, DELETE ON geev.factory_jobs, geev.factory_workspace_files TO authenticated;

-- +goose Down
DROP TABLE geev.factory_workspace_files;
DROP TABLE geev.factory_jobs;
DROP INDEX geev.uq_cards_owner_id_id;
