-- Executar após o trecho Up da migration 00012, na mesma transação revertida.
SET LOCAL ROLE authenticated;
SELECT set_config('request.jwt.claims', '{"sub":"00000000-0000-4000-8000-000000000111","aal":"aal2"}', true);
INSERT INTO geev.clients (id, name) VALUES ('00000000-0000-4000-8000-000000000112', 'Teste reversivel');
INSERT INTO geev.cards (id, type, client_id, title) VALUES ('00000000-0000-4000-8000-000000000113', 'orcamento', '00000000-0000-4000-8000-000000000112', 'Teste reversivel');
INSERT INTO geev.factory_jobs (id, card_id, idempotency_key) VALUES ('00000000-0000-4000-8000-000000000114', '00000000-0000-4000-8000-000000000113', '00000000-0000-4000-8000-000000000115');
INSERT INTO geev.factory_workspace_files (card_id, path, content) VALUES ('00000000-0000-4000-8000-000000000113', 'index.html', '<html>teste</html>');
DO $$ BEGIN
    BEGIN
        INSERT INTO geev.factory_jobs (card_id, idempotency_key) VALUES ('00000000-0000-4000-8000-000000000113', '00000000-0000-4000-8000-000000000116');
        RAISE EXCEPTION 'segunda tarefa ativa foi aceita';
    EXCEPTION WHEN unique_violation THEN NULL; END;
    BEGIN
        INSERT INTO geev.factory_workspace_files (card_id, path, content) VALUES ('00000000-0000-4000-8000-000000000113', '../segredo', 'teste');
        RAISE EXCEPTION 'path traversal foi aceito';
    EXCEPTION WHEN check_violation THEN NULL; END;
END $$;
SELECT set_config('request.jwt.claims', '{"sub":"00000000-0000-4000-8000-000000000222","aal":"aal2"}', true);
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM geev.factory_jobs) OR EXISTS (SELECT 1 FROM geev.factory_workspace_files) THEN
        RAISE EXCEPTION 'outro dono visualizou registros';
    END IF;
    BEGIN
        INSERT INTO geev.factory_workspace_files (card_id, path, content) VALUES ('00000000-0000-4000-8000-000000000113', 'x', 'teste');
        RAISE EXCEPTION 'associação entre donos foi aceita';
    EXCEPTION WHEN foreign_key_violation THEN NULL; END;
END $$;
SELECT set_config('request.jwt.claims', '{"sub":"00000000-0000-4000-8000-000000000111","aal":"aal2"}', true);
UPDATE geev.factory_jobs SET status = 'running', lease_token = '00000000-0000-4000-8000-000000000117', lease_expires_at = clock_timestamp() - interval '1 second', attempts = attempts + 1 WHERE id = '00000000-0000-4000-8000-000000000114';
UPDATE geev.factory_jobs SET status = 'interrupted', lease_token = NULL, lease_expires_at = NULL WHERE status = 'running' AND lease_expires_at <= clock_timestamp();
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM geev.factory_jobs WHERE id = '00000000-0000-4000-8000-000000000114' AND status = 'interrupted' AND attempts = 1) THEN
        RAISE EXCEPTION 'interrupção não registrada';
    END IF;
END $$;
