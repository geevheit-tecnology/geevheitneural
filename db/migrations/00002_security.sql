-- +goose Up

-- 1. Criação das roles da aplicação se não existirem
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'geev_api') THEN
        CREATE ROLE geev_api WITH LOGIN NOINHERIT;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'geev_intake') THEN
        CREATE ROLE geev_intake WITH LOGIN NOINHERIT;
    END IF;
END $$;
-- +goose StatementEnd

-- Permite que a role da API assuma authenticated e anon para aplicar RLS com contexto de claims
GRANT authenticated, anon TO geev_api;
GRANT USAGE ON SCHEMA geev TO authenticated, anon, geev_api, geev_intake;
GRANT ALL ON ALL TABLES IN SCHEMA geev TO authenticated, geev_api;
GRANT ALL ON ALL SEQUENCES IN SCHEMA geev TO authenticated, geev_api;
ALTER DEFAULT PRIVILEGES IN SCHEMA geev GRANT ALL ON TABLES TO authenticated, geev_api;
ALTER DEFAULT PRIVILEGES IN SCHEMA geev GRANT ALL ON SEQUENCES TO authenticated, geev_api;

-- 2. Função de validação do dono autenticado com 2FA (AAL2)
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION geev.is_owner()
RETURNS BOOLEAN AS $$
BEGIN
    RETURN (
        auth.uid() IS NOT NULL AND
        COALESCE(auth.jwt()->>'aal', '') = 'aal2'
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;
-- +goose StatementEnd

-- 3. Função de verificação para execução de conectores e agentes (ExecutionGuard no banco)
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION geev.can_execute(p_actor_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
    v_paused BOOLEAN;
    v_active BOOLEAN;
BEGIN
    SELECT paused INTO v_paused FROM geev.system_state WHERE id = true;
    IF COALESCE(v_paused, true) THEN
        RETURN FALSE;
    END IF;

    -- Verifica se o ator é um agente ativo
    SELECT active INTO v_active FROM geev.agents WHERE id = p_actor_id;
    IF v_active IS TRUE THEN
        RETURN TRUE;
    END IF;

    -- Verifica se o ator é um conector ativo
    SELECT active INTO v_active FROM geev.connectors WHERE id = p_actor_id;
    IF v_active IS TRUE THEN
        RETURN TRUE;
    END IF;

    RETURN FALSE;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;
-- +goose StatementEnd

-- 4. Função de bloqueio de UPDATE e DELETE para tabelas imutáveis
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION geev.prevent_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'A tabela % é imutável. UPDATE e DELETE são proibidos.', TG_TABLE_NAME
        USING ERRCODE = '23505';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- 5. Aplicação dos triggers de imutabilidade
CREATE TRIGGER trg_immutable_card_events
BEFORE UPDATE OR DELETE ON geev.card_events
FOR EACH ROW EXECUTE FUNCTION geev.prevent_mutation();

CREATE TRIGGER trg_immutable_approvals
BEFORE UPDATE OR DELETE ON geev.approvals
FOR EACH ROW EXECUTE FUNCTION geev.prevent_mutation();

CREATE TRIGGER trg_immutable_audit_log
BEFORE UPDATE OR DELETE ON geev.audit_log
FOR EACH ROW EXECUTE FUNCTION geev.prevent_mutation();

-- 6. Habilitação e Forçamento de Row Level Security (RLS) em TODAS as tabelas
ALTER TABLE geev.clients ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.clients FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.apps ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.apps FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.cards ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.cards FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.card_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.card_events FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.proposals ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.proposals FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.approvals ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.approvals FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.risk_reports ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.risk_reports FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.ledger ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.ledger FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.services_costs ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.services_costs FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.connectors ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.connectors FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.agents FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.audit_log FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.system_state ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.system_state FORCE ROW LEVEL SECURITY;

-- 7. Policies para o Dono (autenticado com 2FA)

-- Clientes
CREATE POLICY p_owner_clients ON geev.clients
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Apps
CREATE POLICY p_owner_apps ON geev.apps
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Cards
CREATE POLICY p_owner_cards ON geev.cards
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Card Events (Somente SELECT e INSERT)
CREATE POLICY p_owner_card_events_select ON geev.card_events
    FOR SELECT TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid());

CREATE POLICY p_owner_card_events_insert ON geev.card_events
    FOR INSERT TO authenticated
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Proposals
CREATE POLICY p_owner_proposals ON geev.proposals
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Approvals (Somente SELECT e INSERT)
CREATE POLICY p_owner_approvals_select ON geev.approvals
    FOR SELECT TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid());

CREATE POLICY p_owner_approvals_insert ON geev.approvals
    FOR INSERT TO authenticated
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Risk Reports
CREATE POLICY p_owner_risk_reports ON geev.risk_reports
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Ledger
CREATE POLICY p_owner_ledger ON geev.ledger
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Services Costs
CREATE POLICY p_owner_services_costs ON geev.services_costs
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Connectors
CREATE POLICY p_owner_connectors ON geev.connectors
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Agents
CREATE POLICY p_owner_agents ON geev.agents
    FOR ALL TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid())
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- Audit Log (Somente SELECT e INSERT)
CREATE POLICY p_owner_audit_log_select ON geev.audit_log
    FOR SELECT TO authenticated
    USING (geev.is_owner() AND owner_id = auth.uid());

CREATE POLICY p_owner_audit_log_insert ON geev.audit_log
    FOR INSERT TO authenticated
    WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- System State
CREATE POLICY p_owner_system_state ON geev.system_state
    FOR ALL TO authenticated
    USING (geev.is_owner())
    WITH CHECK (geev.is_owner());

-- +goose Down
DROP POLICY IF EXISTS p_owner_system_state ON geev.system_state;
DROP POLICY IF EXISTS p_owner_audit_log_insert ON geev.audit_log;
DROP POLICY IF EXISTS p_owner_audit_log_select ON geev.audit_log;
DROP POLICY IF EXISTS p_owner_agents ON geev.agents;
DROP POLICY IF EXISTS p_owner_connectors ON geev.connectors;
DROP POLICY IF EXISTS p_owner_services_costs ON geev.services_costs;
DROP POLICY IF EXISTS p_owner_ledger ON geev.ledger;
DROP POLICY IF EXISTS p_owner_risk_reports ON geev.risk_reports;
DROP POLICY IF EXISTS p_owner_approvals_insert ON geev.approvals;
DROP POLICY IF EXISTS p_owner_approvals_select ON geev.approvals;
DROP POLICY IF EXISTS p_owner_proposals ON geev.proposals;
DROP POLICY IF EXISTS p_owner_card_events_insert ON geev.card_events;
DROP POLICY IF EXISTS p_owner_card_events_select ON geev.card_events;
DROP POLICY IF EXISTS p_owner_cards ON geev.cards;
DROP POLICY IF EXISTS p_owner_apps ON geev.apps;
DROP POLICY IF EXISTS p_owner_clients ON geev.clients;

DROP TRIGGER IF EXISTS trg_immutable_audit_log ON geev.audit_log;
DROP TRIGGER IF EXISTS trg_immutable_approvals ON geev.approvals;
DROP TRIGGER IF EXISTS trg_immutable_card_events ON geev.card_events;
DROP FUNCTION IF EXISTS geev.prevent_mutation CASCADE;
DROP FUNCTION IF EXISTS geev.can_execute CASCADE;
DROP FUNCTION IF EXISTS geev.is_owner CASCADE;
