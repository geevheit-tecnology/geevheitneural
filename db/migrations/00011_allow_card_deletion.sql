-- +goose Up
-- Permite que a exclusão autorizada de um cartão pelo dono limpe os registros de eventos e aprovações vinculados
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION geev.prevent_mutation()
RETURNS TRIGGER AS $$
BEGIN
    IF current_setting('geev.allow_card_delete', true) = 'on' AND TG_TABLE_NAME IN ('card_events', 'approvals') THEN
        RETURN OLD;
    END IF;

    RAISE EXCEPTION 'A tabela % é imutável. UPDATE e DELETE são proibidos.', TG_TABLE_NAME
        USING ERRCODE = '23505';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION geev.prevent_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'A tabela % é imutável. UPDATE e DELETE são proibidos.', TG_TABLE_NAME
        USING ERRCODE = '23505';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
