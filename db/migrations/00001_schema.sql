-- +goose Up
CREATE SCHEMA IF NOT EXISTS geev;

-- Enums
CREATE TYPE geev.card_type AS ENUM ('orcamento', 'bug', 'suporte', 'financeiro');
CREATE TYPE geev.card_stage AS ENUM (
    'entrada',
    'proposta_aprovar',
    'aguardando_cliente',
    'prompt_aprovar',
    'em_execucao',
    'entrega_aprovar',
    'entregue',
    'recusado'
);
CREATE TYPE geev.decision AS ENUM ('aprovado', 'mais_info', 'recusado');
CREATE TYPE geev.entry_kind AS ENUM ('receita', 'custo');
CREATE TYPE geev.proposal_status AS ENUM ('rascunho', 'enviada', 'aceita', 'recusada');

-- Função utilitária para atualização automática de updated_at
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION geev.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = clock_timestamp();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- 1. Clientes
CREATE TABLE geev.clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    name TEXT NOT NULL,
    contact_ref TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_clients_updated_at
BEFORE UPDATE ON geev.clients
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 2. Apps
CREATE TABLE geev.apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    client_id UUID REFERENCES geev.clients(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    hosting TEXT NOT NULL DEFAULT '',
    database TEXT NOT NULL DEFAULT '',
    repository TEXT NOT NULL DEFAULT '',
    domain TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'ativo',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_apps_updated_at
BEFORE UPDATE ON geev.apps
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 3. Cartões
CREATE TABLE geev.cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    type geev.card_type NOT NULL,
    client_id UUID NOT NULL REFERENCES geev.clients(id) ON DELETE CASCADE,
    app_id UUID REFERENCES geev.apps(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    price_cents BIGINT NOT NULL DEFAULT 0,
    est_cost_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'BRL',
    stage geev.card_stage NOT NULL DEFAULT 'entrada',
    priority SMALLINT NOT NULL DEFAULT 0,
    due_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_cards_updated_at
BEFORE UPDATE ON geev.cards
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 4. Eventos do Cartão (imutável)
CREATE TABLE geev.card_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    actor_type TEXT NOT NULL DEFAULT 'owner',
    actor_id UUID,
    at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- 5. Propostas
CREATE TABLE geev.proposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    amount_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'BRL',
    body TEXT NOT NULL DEFAULT '',
    accept_token_hash TEXT,
    status geev.proposal_status NOT NULL DEFAULT 'rascunho',
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_proposals_updated_at
BEFORE UPDATE ON geev.proposals
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 6. Aprovações de Portões (imutável)
CREATE TABLE geev.approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    gate geev.card_stage NOT NULL,
    decision geev.decision NOT NULL,
    decided_by UUID NOT NULL DEFAULT auth.uid(),
    decided_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    note TEXT NOT NULL DEFAULT ''
);

-- 7. Ficha de Risco
CREATE TABLE geev.risk_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    tests JSONB NOT NULL DEFAULT '[]'::jsonb,
    secrets_scan JSONB NOT NULL DEFAULT '[]'::jsonb,
    outdated_deps JSONB NOT NULL DEFAULT '[]'::jsonb,
    lgpd_points JSONB NOT NULL DEFAULT '[]'::jsonb,
    source TEXT NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_risk_reports_updated_at
BEFORE UPDATE ON geev.risk_reports
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 8. Ledger (Lançamentos financeiros)
CREATE TABLE geev.ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    app_id UUID REFERENCES geev.apps(id) ON DELETE SET NULL,
    client_id UUID REFERENCES geev.clients(id) ON DELETE SET NULL,
    kind geev.entry_kind NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'BRL',
    occurred_on DATE NOT NULL DEFAULT CURRENT_DATE,
    service TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- 9. Custos de Serviços (mensalidade por serviço e app)
CREATE TABLE geev.services_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    service TEXT NOT NULL,
    app_id UUID REFERENCES geev.apps(id) ON DELETE CASCADE,
    monthly_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'BRL',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_services_costs_updated_at
BEFORE UPDATE ON geev.services_costs
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 10. Conectores
CREATE TABLE geev.connectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    name TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT false,
    secret_ref TEXT NOT NULL DEFAULT '',
    scope TEXT NOT NULL DEFAULT 'read_only',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_connectors_updated_at
BEFORE UPDATE ON geev.connectors
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 11. Agentes
CREATE TABLE geev.agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    role TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT false,
    spend_limit_cents BIGINT NOT NULL DEFAULT 0,
    scope TEXT NOT NULL DEFAULT 'minimal',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_agents_updated_at
BEFORE UPDATE ON geev.agents
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 12. Audit Log (imutável)
CREATE TABLE geev.audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    action TEXT NOT NULL,
    actor_type TEXT NOT NULL DEFAULT 'owner',
    actor_id UUID,
    target_type TEXT NOT NULL DEFAULT '',
    target_id UUID,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- 13. System State (singleton para Botão Parar / Parada de Emergência)
CREATE TABLE geev.system_state (
    id BOOLEAN PRIMARY KEY DEFAULT true CHECK (id = true),
    owner_id UUID,
    paused BOOLEAN NOT NULL DEFAULT false,
    paused_by UUID,
    paused_at TIMESTAMPTZ,
    reason TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_system_state_updated_at
BEFORE UPDATE ON geev.system_state
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- Inicia o singleton do sistema em estado normal (não pausado)
INSERT INTO geev.system_state (id, paused, reason)
VALUES (true, false, '')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS geev.system_state CASCADE;
DROP TABLE IF EXISTS geev.audit_log CASCADE;
DROP TABLE IF EXISTS geev.agents CASCADE;
DROP TABLE IF EXISTS geev.connectors CASCADE;
DROP TABLE IF EXISTS geev.services_costs CASCADE;
DROP TABLE IF EXISTS geev.ledger CASCADE;
DROP TABLE IF EXISTS geev.risk_reports CASCADE;
DROP TABLE IF EXISTS geev.approvals CASCADE;
DROP TABLE IF EXISTS geev.proposals CASCADE;
DROP TABLE IF EXISTS geev.card_events CASCADE;
DROP TABLE IF EXISTS geev.cards CASCADE;
DROP TABLE IF EXISTS geev.apps CASCADE;
DROP TABLE IF EXISTS geev.clients CASCADE;
DROP FUNCTION IF EXISTS geev.set_updated_at CASCADE;
DROP TYPE IF EXISTS geev.proposal_status CASCADE;
DROP TYPE IF EXISTS geev.entry_kind CASCADE;
DROP TYPE IF EXISTS geev.decision CASCADE;
DROP TYPE IF EXISTS geev.card_stage CASCADE;
DROP TYPE IF EXISTS geev.card_type CASCADE;
DROP SCHEMA IF EXISTS geev CASCADE;
