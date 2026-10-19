-- name: ListPlaybooks :many
SELECT id, owner_id, product_type, title, description, stack, checklist, prompts, autonomy_level, created_at, updated_at
FROM geev.playbooks
ORDER BY product_type ASC;

-- name: GetPlaybookByType :one
SELECT id, owner_id, product_type, title, description, stack, checklist, prompts, autonomy_level, created_at, updated_at
FROM geev.playbooks
WHERE product_type = $1
LIMIT 1;

-- name: UpsertPlaybook :one
INSERT INTO geev.playbooks (
    id, owner_id, product_type, title, description, stack, checklist, prompts, autonomy_level, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) ON CONFLICT (product_type) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    stack = EXCLUDED.stack,
    checklist = EXCLUDED.checklist,
    prompts = EXCLUDED.prompts,
    autonomy_level = EXCLUDED.autonomy_level,
    updated_at = EXCLUDED.updated_at
RETURNING id, owner_id, product_type, title, description, stack, checklist, prompts, autonomy_level, created_at, updated_at;

-- name: ListModules :many
SELECT id, owner_id, name, category, description, code_templates, is_reviewed, created_at, updated_at
FROM geev.modules
ORDER BY category ASC, name ASC;

-- name: UpsertModule :one
INSERT INTO geev.modules (
    id, owner_id, name, category, description, code_templates, is_reviewed, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) ON CONFLICT (name) DO UPDATE
SET category = EXCLUDED.category,
    description = EXCLUDED.description,
    code_templates = EXCLUDED.code_templates,
    is_reviewed = EXCLUDED.is_reviewed,
    updated_at = EXCLUDED.updated_at
RETURNING id, owner_id, name, category, description, code_templates, is_reviewed, created_at, updated_at;
