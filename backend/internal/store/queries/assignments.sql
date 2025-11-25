-- name: ListRuleAssignments :many
SELECT *
FROM rule_assignments
WHERE rule_id = $1;

-- name: InsertRuleAssignment :exec
INSERT INTO rule_assignments (rule_id, scope_id, target_type, action, user_id, group_id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING;

-- name: DeleteAssignmentsByRule :exec
DELETE FROM rule_assignments WHERE rule_id = $1;

-- name: ListAllAssignments :many
SELECT *
FROM rule_assignments;

-- name: ListUserAssignments :many
SELECT
    rs.id AS scope_id,
    rs.rule_id,
    rs.target_type,
    rs.action,
    (CASE WHEN rs.target_type = 'user' THEN rs.target_id ELSE gm.user_id END)::uuid AS user_id,
    (CASE WHEN rs.target_type = 'group' THEN rs.target_id END)::uuid AS group_id,
    r.name AS rule_name,
    r.type AS rule_type,
    r.target AS rule_target,
    rs.created_at AS scope_created_at,
    g.display_name AS group_name
FROM rule_scopes rs
JOIN rules r ON r.id = rs.rule_id
LEFT JOIN group_members gm ON rs.target_type = 'group' AND rs.target_id = gm.group_id
LEFT JOIN groups g ON rs.target_type = 'group' AND g.id = rs.target_id
WHERE (rs.target_type = 'user' AND rs.target_id = $1)
   OR (rs.target_type = 'group' AND gm.user_id = $1)
ORDER BY r.name, rs.created_at;

-- name: ListApplicationAssignmentStats :many
WITH scope_counts AS (
    SELECT
        rule_id,
        COUNT(*) FILTER (WHERE action = 'allow') AS allow_scopes,
        COUNT(*) FILTER (WHERE action = 'block') AS block_scopes,
        COUNT(*) FILTER (WHERE action = 'cel') AS cel_scopes,
        COUNT(*) AS total_scopes
    FROM rule_scopes
    GROUP BY rule_id
),
user_counts AS (
    SELECT
        rule_id,
        COUNT(DISTINCT user_id) FILTER (WHERE action = 'allow') AS allow_users,
        COUNT(DISTINCT user_id) FILTER (WHERE action = 'block') AS block_users,
        COUNT(DISTINCT user_id) FILTER (WHERE action = 'cel') AS cel_users,
        COUNT(DISTINCT user_id) AS total_users
    FROM (
        SELECT
            rs.rule_id,
            rs.action,
            CASE
                WHEN rs.target_type = 'user' THEN rs.target_id
                WHEN rs.target_type = 'group' THEN gm.user_id
            END AS user_id
        FROM rule_scopes rs
        LEFT JOIN group_members gm ON rs.target_type = 'group' AND rs.target_id = gm.group_id
    ) expanded
    WHERE user_id IS NOT NULL
    GROUP BY rule_id
)
SELECT
    sc.rule_id,
    sc.allow_scopes,
    sc.block_scopes,
    sc.cel_scopes,
    sc.total_scopes,
    COALESCE(uc.allow_users, 0) AS allow_users,
    COALESCE(uc.block_users, 0) AS block_users,
    COALESCE(uc.cel_users, 0) AS cel_users,
    COALESCE(uc.total_users, 0) AS total_users
FROM scope_counts sc
LEFT JOIN user_counts uc USING (rule_id);
