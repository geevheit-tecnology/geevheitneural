-- +goose Up
-- Permite leitura de catálogo de playbooks e módulos reutilizáveis por authenticated
DROP POLICY IF EXISTS p_owner_playbooks ON geev.playbooks;
DROP POLICY IF EXISTS p_owner_modules ON geev.modules;

CREATE POLICY p_catalog_read_playbooks ON geev.playbooks FOR SELECT TO authenticated, anon USING (true);
CREATE POLICY p_catalog_write_playbooks ON geev.playbooks FOR ALL TO authenticated USING (geev.is_owner());

CREATE POLICY p_catalog_read_modules ON geev.modules FOR SELECT TO authenticated, anon USING (true);
CREATE POLICY p_catalog_write_modules ON geev.modules FOR ALL TO authenticated USING (geev.is_owner());

-- +goose Down
DROP POLICY IF EXISTS p_catalog_read_playbooks ON geev.playbooks;
DROP POLICY IF EXISTS p_catalog_write_playbooks ON geev.playbooks;
DROP POLICY IF EXISTS p_catalog_read_modules ON geev.modules;
DROP POLICY IF EXISTS p_catalog_write_modules ON geev.modules;
