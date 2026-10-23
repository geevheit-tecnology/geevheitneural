-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS geev.subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT 'a0000000-0000-0000-0000-000000000001',
    client_id UUID NOT NULL REFERENCES geev.clients(id) ON DELETE CASCADE,
    app_id UUID REFERENCES geev.apps(id) ON DELETE SET NULL,
    plan TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'BRL',
    interval TEXT NOT NULL DEFAULT 'mensal', -- mensal, trimestral, semestral, anual
    due_day INT NOT NULL DEFAULT 10,
    status TEXT NOT NULL DEFAULT 'ativa', -- ativa, pausada, cancelada, inadimplente
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,
    next_due_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE TRIGGER update_subscriptions_updated_at
BEFORE UPDATE ON geev.subscriptions
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

CREATE TABLE IF NOT EXISTS geev.subscription_charges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT 'a0000000-0000-0000-0000-000000000001',
    subscription_id UUID NOT NULL REFERENCES geev.subscriptions(id) ON DELETE CASCADE,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'BRL',
    due_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'pendente', -- pendente, paga, atrasada
    paid_at TIMESTAMPTZ,
    payment_method TEXT NOT NULL DEFAULT 'pix',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT uq_subscription_charge_due UNIQUE (subscription_id, due_date)
);

CREATE TRIGGER update_subscription_charges_updated_at
BEFORE UPDATE ON geev.subscription_charges
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- Habilita e força RLS
ALTER TABLE geev.subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.subscriptions FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.subscription_charges ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.subscription_charges FORCE ROW LEVEL SECURITY;

-- Políticas para subscriptions
CREATE POLICY subscriptions_owner ON geev.subscriptions
    FOR ALL TO authenticated
    USING (geev.is_owner());

CREATE POLICY subscriptions_api ON geev.subscriptions
    FOR ALL TO geev_api
    USING (true);

-- Políticas para subscription_charges
CREATE POLICY subscription_charges_owner ON geev.subscription_charges
    FOR ALL TO authenticated
    USING (geev.is_owner());

CREATE POLICY subscription_charges_api ON geev.subscription_charges
    FOR ALL TO geev_api
    USING (true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS geev.subscription_charges CASCADE;
DROP TABLE IF EXISTS geev.subscriptions CASCADE;
-- +goose StatementEnd
