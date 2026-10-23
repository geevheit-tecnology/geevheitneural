-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS geev.webhook_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT 'a0000000-0000-0000-0000-000000000001',
    url TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    secret_token TEXT NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{"*"}',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE TRIGGER update_webhook_endpoints_updated_at
BEFORE UPDATE ON geev.webhook_endpoints
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

CREATE TABLE IF NOT EXISTS geev.webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT 'a0000000-0000-0000-0000-000000000001',
    webhook_id UUID NOT NULL REFERENCES geev.webhook_endpoints(id) ON DELETE CASCADE,
    event TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status_code INT NOT NULL DEFAULT 0,
    attempts INT NOT NULL DEFAULT 1,
    response_body TEXT NOT NULL DEFAULT '',
    success BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE IF NOT EXISTS geev.notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT 'a0000000-0000-0000-0000-000000000001' UNIQUE,
    channels JSONB NOT NULL DEFAULT '{"telegram":{"enabled":false},"whatsapp":{"enabled":false},"email":{"enabled":false}}'::jsonb,
    events JSONB NOT NULL DEFAULT '{"intake":true,"pix":true,"portal_accept":true,"subscription_overdue":true}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE TRIGGER update_notification_preferences_updated_at
BEFORE UPDATE ON geev.notification_preferences
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- Habilita e força RLS
ALTER TABLE geev.webhook_endpoints ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.webhook_endpoints FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.webhook_deliveries ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.webhook_deliveries FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.notification_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.notification_preferences FORCE ROW LEVEL SECURITY;

-- Políticas para webhook_endpoints
CREATE POLICY webhook_endpoints_owner ON geev.webhook_endpoints
    FOR ALL TO authenticated
    USING (geev.is_owner());

CREATE POLICY webhook_endpoints_api ON geev.webhook_endpoints
    FOR ALL TO geev_api
    USING (true);

-- Políticas para webhook_deliveries
CREATE POLICY webhook_deliveries_owner ON geev.webhook_deliveries
    FOR ALL TO authenticated
    USING (geev.is_owner());

CREATE POLICY webhook_deliveries_api ON geev.webhook_deliveries
    FOR ALL TO geev_api
    USING (true);

-- Políticas para notification_preferences
CREATE POLICY notification_preferences_owner ON geev.notification_preferences
    FOR ALL TO authenticated
    USING (geev.is_owner());

CREATE POLICY notification_preferences_api ON geev.notification_preferences
    FOR ALL TO geev_api
    USING (true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS geev.notification_preferences CASCADE;
DROP TABLE IF EXISTS geev.webhook_deliveries CASCADE;
DROP TABLE IF EXISTS geev.webhook_endpoints CASCADE;
-- +goose StatementEnd
