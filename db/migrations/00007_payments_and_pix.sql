-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS geev.payment_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT 'a0000000-0000-0000-0000-000000000001',
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES geev.clients(id),
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    currency TEXT NOT NULL DEFAULT 'BRL',
    pix_emv TEXT NOT NULL,
    pix_txid TEXT NOT NULL UNIQUE,
    pix_key TEXT NOT NULL DEFAULT 'pix@geevheit.com',
    status TEXT NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT payment_orders_status_check CHECK (status IN ('pending', 'paid', 'expired', 'cancelled'))
);

CREATE TRIGGER update_payment_orders_updated_at
BEFORE UPDATE ON geev.payment_orders
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

ALTER TABLE geev.payment_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.payment_orders FORCE ROW LEVEL SECURITY;

CREATE POLICY payment_orders_owner ON geev.payment_orders
    FOR ALL TO authenticated
    USING (geev.is_owner());

CREATE POLICY payment_orders_api ON geev.payment_orders
    FOR ALL TO geev_api
    USING (true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS geev.payment_orders CASCADE;
-- +goose StatementEnd
