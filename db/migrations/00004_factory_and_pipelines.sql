-- +goose Up
-- Migration 00004: Pipeline de Fábrica de Software, Especificações Executáveis, Playbooks, Módulos, Previews e Métricas

-- 1. Pipelines de Execução por Cartão
CREATE TABLE IF NOT EXISTS geev.card_pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    stage TEXT NOT NULL DEFAULT 'requisitos',
    active_role TEXT NOT NULL DEFAULT 'requisitos',
    spend_total_cents BIGINT NOT NULL DEFAULT 0,
    spend_limit_cents BIGINT NOT NULL DEFAULT 10000,
    status TEXT NOT NULL DEFAULT 'idle',
    current_task TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT uq_card_pipeline UNIQUE (card_id)
);
CREATE TRIGGER trg_card_pipelines_updated_at
BEFORE UPDATE ON geev.card_pipelines
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 2. Especificação Executável (User Stories e BDD)
CREATE TABLE IF NOT EXISTS geev.specifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    user_stories JSONB NOT NULL DEFAULT '[]'::jsonb,
    tests_code TEXT NOT NULL DEFAULT '',
    tests_passed BOOLEAN NOT NULL DEFAULT false,
    client_approved BOOLEAN NOT NULL DEFAULT false,
    client_approved_at TIMESTAMPTZ,
    payment_signal_cents BIGINT NOT NULL DEFAULT 0,
    payment_signal_paid BOOLEAN NOT NULL DEFAULT false,
    scope_token_hash TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT uq_card_specification UNIQUE (card_id)
);
CREATE TRIGGER trg_specifications_updated_at
BEFORE UPDATE ON geev.specifications
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 3. Biblioteca de Playbooks
CREATE TABLE IF NOT EXISTS geev.playbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    product_type TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    stack JSONB NOT NULL DEFAULT '{}'::jsonb,
    checklist JSONB NOT NULL DEFAULT '[]'::jsonb,
    prompts JSONB NOT NULL DEFAULT '{}'::jsonb,
    autonomy_level INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_playbooks_updated_at
BEFORE UPDATE ON geev.playbooks
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 4. Biblioteca de Módulos Reutilizáveis
CREATE TABLE IF NOT EXISTS geev.modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    name TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL DEFAULT 'general',
    description TEXT NOT NULL DEFAULT '',
    code_templates JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_reviewed BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_modules_updated_at
BEFORE UPDATE ON geev.modules
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 5. Prévia ao Vivo e Reversão (Previews)
CREATE TABLE IF NOT EXISTS geev.previews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    preview_url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    rollback_plan JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER trg_previews_updated_at
BEFORE UPDATE ON geev.previews
FOR EACH ROW EXECUTE FUNCTION geev.set_updated_at();

-- 6. Memória Viva Isolada por Projeto
CREATE TABLE IF NOT EXISTS geev.project_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES geev.clients(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'nota',
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX IF NOT EXISTS idx_project_memory_card ON geev.project_memory(card_id);
CREATE INDEX IF NOT EXISTS idx_project_memory_client ON geev.project_memory(client_id);

-- 7. Placar da Fábrica (Métricas)
CREATE TABLE IF NOT EXISTS geev.factory_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL DEFAULT auth.uid(),
    card_id UUID NOT NULL REFERENCES geev.cards(id) ON DELETE CASCADE,
    lead_time_seconds BIGINT NOT NULL DEFAULT 0,
    ai_cost_cents BIGINT NOT NULL DEFAULT 0,
    infra_cost_cents BIGINT NOT NULL DEFAULT 0,
    first_time_approval BOOLEAN NOT NULL DEFAULT false,
    post_delivery_defects INT NOT NULL DEFAULT 0,
    owner_decision_time_seconds BIGINT NOT NULL DEFAULT 0,
    margin_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- Habilita e Força RLS em todas as tabelas novas (Princípio Inegociável)
ALTER TABLE geev.card_pipelines ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.card_pipelines FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.specifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.specifications FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.playbooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.playbooks FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.modules ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.modules FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.previews ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.previews FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.project_memory ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.project_memory FORCE ROW LEVEL SECURITY;

ALTER TABLE geev.factory_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE geev.factory_metrics FORCE ROW LEVEL SECURITY;

-- Policies para authenticated com is_owner()
CREATE POLICY p_owner_card_pipelines ON geev.card_pipelines FOR ALL TO authenticated USING (geev.is_owner() AND owner_id = auth.uid()) WITH CHECK (geev.is_owner() AND owner_id = auth.uid());
CREATE POLICY p_owner_specifications ON geev.specifications FOR ALL TO authenticated USING (geev.is_owner() AND owner_id = auth.uid()) WITH CHECK (geev.is_owner() AND owner_id = auth.uid());
CREATE POLICY p_owner_playbooks ON geev.playbooks FOR ALL TO authenticated USING (geev.is_owner() AND owner_id = auth.uid()) WITH CHECK (geev.is_owner() AND owner_id = auth.uid());
CREATE POLICY p_owner_modules ON geev.modules FOR ALL TO authenticated USING (geev.is_owner() AND owner_id = auth.uid()) WITH CHECK (geev.is_owner() AND owner_id = auth.uid());
CREATE POLICY p_owner_previews ON geev.previews FOR ALL TO authenticated USING (geev.is_owner() AND owner_id = auth.uid()) WITH CHECK (geev.is_owner() AND owner_id = auth.uid());
CREATE POLICY p_owner_project_memory ON geev.project_memory FOR ALL TO authenticated USING (geev.is_owner() AND owner_id = auth.uid()) WITH CHECK (geev.is_owner() AND owner_id = auth.uid());
CREATE POLICY p_owner_factory_metrics ON geev.factory_metrics FOR ALL TO authenticated USING (geev.is_owner() AND owner_id = auth.uid()) WITH CHECK (geev.is_owner() AND owner_id = auth.uid());

-- +goose Down
DROP TABLE IF EXISTS geev.factory_metrics CASCADE;
DROP TABLE IF EXISTS geev.project_memory CASCADE;
DROP TABLE IF EXISTS geev.previews CASCADE;
DROP TABLE IF EXISTS geev.modules CASCADE;
DROP TABLE IF EXISTS geev.playbooks CASCADE;
DROP TABLE IF EXISTS geev.specifications CASCADE;
DROP TABLE IF EXISTS geev.card_pipelines CASCADE;
