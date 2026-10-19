BEGIN;
SELECT plan(8);

-- 1. Verifica se schema geev existe
SELECT has_schema('geev', 'Schema geev deve existir');

-- 2. Verifica se as tabelas principais existem
SELECT has_table('geev', 'cards', 'Tabela cards deve existir');
SELECT has_table('geev', 'approvals', 'Tabela approvals deve existir');
SELECT has_table('geev', 'card_events', 'Tabela card_events deve existir');
SELECT has_table('geev', 'audit_log', 'Tabela audit_log deve existir');

-- 3. Verifica permissões essenciais da role da API
SELECT ok(has_table_privilege('geev_api', 'geev.cards', 'SELECT'), 'geev_api deve ter permissão SELECT');

-- 4. Setup de dados para validação de FK e imutabilidade
INSERT INTO geev.clients (id, owner_id, name)
VALUES ('00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', 'Cliente Teste');

INSERT INTO geev.cards (id, owner_id, client_id, type, title, price_cents)
VALUES ('00000000-0000-0000-0000-000000000020', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000010', 'orcamento', 'Card Teste', 100000);

PREPARE insert_approval AS
INSERT INTO geev.approvals (id, owner_id, card_id, gate, decision, decided_by, note)
VALUES ('00000000-0000-0000-0000-000000000030', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000020', 'proposta_aprovar', 'aprovado', '00000000-0000-0000-0000-000000000001', 'Nota teste');

SELECT lives_ok('insert_approval', 'Inserção em approvals deve funcionar com chave estrangeira e owner válidos');

PREPARE update_approval AS
UPDATE geev.approvals SET note = 'Tentativa de alteração' WHERE id = '00000000-0000-0000-0000-000000000030';

SELECT throws_ok('update_approval', '23505', NULL, 'Tentativa de UPDATE em approvals deve falhar por trigger de imutabilidade');

SELECT * FROM finish();
ROLLBACK;
