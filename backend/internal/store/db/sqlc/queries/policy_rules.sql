-- name: ListPolicyRuleAttachmentsByPolicyID :many
SELECT * FROM policy_rules WHERE policy_id = $1 ORDER BY rule_id;

-- name: ListPolicyRuleAttachmentsForSyncByPolicyID :many
SELECT * FROM policy_rules WHERE policy_id = $1 ORDER BY rule_id LIMIT $2 OFFSET $3;

-- name: CreatePolicyRuleAttachment :exec
INSERT INTO policy_rules (policy_id, rule_id, action, cel_expr)
VALUES ($1, $2, $3, $4);

-- name: DeletePolicyRuleAttachmentsByPolicyID :exec
DELETE FROM policy_rules WHERE policy_id = $1;
