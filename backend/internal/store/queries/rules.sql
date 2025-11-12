-- name: CreateRule :one
INSERT INTO rules (id, name, type, target, scope, enabled, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateRule :one
UPDATE rules
SET name = $2,
    type = $3,
    target = $4,
    scope = $5,
    enabled = $6,
    metadata = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListRules :many
SELECT *
FROM rules
ORDER BY created_at DESC;

-- name: FilterRules :many
SELECT *
FROM rules
WHERE CASE
        WHEN sqlc.arg(search)::text = '' THEN TRUE
        ELSE (
            to_tsvector(
                'simple',
                coalesce(name, '') || ' ' ||
                coalesce(type, '') || ' ' ||
                coalesce(target, '') || ' ' ||
                coalesce(metadata ->> 'description', '')
            ) @@ websearch_to_tsquery('simple', sqlc.arg(search)::text)
            OR name ILIKE '%' || sqlc.arg(search)::text || '%'
            OR target ILIKE '%' || sqlc.arg(search)::text || '%'
            OR coalesce(metadata ->> 'description', '') ILIKE '%' || sqlc.arg(search)::text || '%'
        )
     END
  AND (sqlc.arg(rule_type)::text = '' OR LOWER(type) = LOWER(sqlc.arg(rule_type)::text))
  AND (sqlc.narg(enabled)::boolean IS NULL OR enabled = sqlc.narg(enabled)::boolean)
  AND (sqlc.arg(identifier)::text = '' OR target ILIKE '%' || sqlc.arg(identifier)::text || '%')
ORDER BY created_at DESC;

-- name: GetRule :one
SELECT *
FROM rules
WHERE id = $1;

-- name: GetRuleByTarget :one
SELECT *
FROM rules
WHERE LOWER(target) = LOWER($1)
LIMIT 1;

-- name: DeleteRule :exec
DELETE FROM rules WHERE id = $1;
