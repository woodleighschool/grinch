// Package policies persists policy data and related records.
package policies

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/domain/policies"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/store/constraints"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// Repo provides persistence operations for policies.
type Repo struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo backed by the given pool.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: sqlc.New(pool), pool: pool}
}

// Get returns a policy by ID, including targets and attachments.
func (r *Repo) Get(ctx context.Context, id uuid.UUID) (policies.Policy, error) {
	row, err := r.q.GetPolicyByID(ctx, id)
	if err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}

	policy := toDomainPolicy(row)

	targets, err := r.q.ListPolicyTargetsByPolicyID(ctx, id)
	if err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}
	policy.Targets = toDomainTargets(targets)

	atts, err := r.q.ListPolicyRuleAttachmentsByPolicyID(ctx, id)
	if err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}
	policy.Attachments = toDomainAttachments(atts)

	return policy, nil
}

// Create inserts a policy and its targets and attachments.
func (r *Repo) Create(ctx context.Context, policy policies.Policy) (policies.Policy, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.q.WithTx(tx)

	row, err := qtx.CreatePolicy(ctx, toCreateParams(policy))
	if err != nil {
		return policies.Policy{}, errx.FromStore(err, constraints.PolicyFields())
	}

	created := toDomainPolicy(row)

	if err = saveTargets(ctx, qtx, created.ID, policy.Targets); err != nil {
		return policies.Policy{}, err
	}

	if err = saveAttachments(ctx, qtx, created.ID, policy.Attachments); err != nil {
		return policies.Policy{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}

	created.Targets = policy.Targets
	created.Attachments = policy.Attachments
	return created, nil
}

// Update replaces a policy and its targets and attachments.
func (r *Repo) Update(ctx context.Context, policy policies.Policy) (policies.Policy, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.q.WithTx(tx)

	row, err := qtx.UpdatePolicyByID(ctx, toUpdateParams(policy))
	if err != nil {
		return policies.Policy{}, errx.FromStore(err, constraints.PolicyFields())
	}

	updated := toDomainPolicy(row)

	if err = qtx.DeletePolicyTargetsByPolicyID(ctx, updated.ID); err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}
	if err = saveTargets(ctx, qtx, updated.ID, policy.Targets); err != nil {
		return policies.Policy{}, err
	}

	if err = qtx.DeletePolicyRuleAttachmentsByPolicyID(ctx, updated.ID); err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}
	if err = saveAttachments(ctx, qtx, updated.ID, policy.Attachments); err != nil {
		return policies.Policy{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return policies.Policy{}, errx.FromStore(err, nil)
	}

	updated.Targets = policy.Targets
	updated.Attachments = policy.Attachments
	return updated, nil
}

// Delete removes a policy by ID.
func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	return errx.FromStore(r.q.DeletePolicyByID(ctx, id), nil)
}

// List returns policies matching the listing query.
func (r *Repo) List(ctx context.Context, query listing.Query) ([]policies.ListItem, listing.Page, error) {
	items, total, err := listPolicies(ctx, r.pool, query)
	if err != nil {
		return nil, listing.Page{}, errx.FromStore(err, nil)
	}
	return items, listing.Page{Total: total}, nil
}

// ListEnabled returns enabled policies ordered by priority.
func (r *Repo) ListEnabled(ctx context.Context) ([]policies.Policy, error) {
	rows, err := r.q.ListEnabledPolicies(ctx)
	if err != nil {
		return nil, errx.FromStore(err, nil)
	}
	out := make([]policies.Policy, len(rows))
	for i, row := range rows {
		out[i] = toDomainPolicy(row)
	}
	return out, nil
}

// ListPolicyTargetsByPolicyIDs returns targets for the given policy IDs.
func (r *Repo) ListPolicyTargetsByPolicyIDs(ctx context.Context, policyIDs []uuid.UUID) ([]policies.Target, error) {
	if len(policyIDs) == 0 {
		return nil, nil
	}
	rows, err := r.q.ListPolicyTargetsByPolicyIDs(ctx, policyIDs)
	if err != nil {
		return nil, errx.FromStore(err, nil)
	}
	return toDomainTargets(rows), nil
}

// ListPolicyRuleAttachmentsByPolicyID returns attachments for a policy.
func (r *Repo) ListPolicyRuleAttachmentsByPolicyID(
	ctx context.Context,
	policyID uuid.UUID,
) ([]policies.Attachment, error) {
	rows, err := r.q.ListPolicyRuleAttachmentsByPolicyID(ctx, policyID)
	if err != nil {
		return nil, errx.FromStore(err, nil)
	}
	return toDomainAttachments(rows), nil
}

// ListPolicyRuleAttachmentsForSyncByPolicyID returns a page of attachments for sync.
func (r *Repo) ListPolicyRuleAttachmentsForSyncByPolicyID(
	ctx context.Context,
	policyID uuid.UUID,
	limit, offset int,
) ([]policies.Attachment, error) {
	l, o := pgconv.LimitOffset(limit, offset)
	rows, err := r.q.ListPolicyRuleAttachmentsForSyncByPolicyID(
		ctx,
		sqlc.ListPolicyRuleAttachmentsForSyncByPolicyIDParams{
			PolicyID: policyID,
			Limit:    l,
			Offset:   o,
		},
	)
	if err != nil {
		return nil, errx.FromStore(err, nil)
	}
	return toDomainAttachments(rows), nil
}

// UpdatePolicyRulesVersionByRuleID bumps rules_version for policies that reference the given rule.
func (r *Repo) UpdatePolicyRulesVersionByRuleID(ctx context.Context, ruleID uuid.UUID) error {
	return errx.FromStore(r.q.UpdatePolicyRulesVersionByRuleID(ctx, ruleID), nil)
}

func saveTargets(ctx context.Context, q *sqlc.Queries, policyID uuid.UUID, targets []policies.Target) error {
	for _, t := range targets {
		var userID *uuid.UUID
		var groupID *uuid.UUID
		var machineID *uuid.UUID
		switch t.Kind {
		case policies.TargetUser:
			userID = t.RefID
		case policies.TargetGroup:
			groupID = t.RefID
		case policies.TargetMachine:
			machineID = t.RefID
		case policies.TargetAll:
		}
		if err := q.CreatePolicyTarget(ctx, sqlc.CreatePolicyTargetParams{
			PolicyID:  policyID,
			Kind:      string(t.Kind),
			UserID:    userID,
			GroupID:   groupID,
			MachineID: machineID,
		}); err != nil {
			return errx.FromStore(err, nil)
		}
	}
	return nil
}

func saveAttachments(ctx context.Context, q *sqlc.Queries, policyID uuid.UUID, atts []policies.Attachment) error {
	for _, a := range atts {
		if err := q.CreatePolicyRuleAttachment(ctx, sqlc.CreatePolicyRuleAttachmentParams{
			PolicyID: policyID,
			RuleID:   a.RuleID,
			Action:   int32(a.Action),
			CelExpr:  pgconv.TextOrNull(a.CELExpr),
		}); err != nil {
			return errx.FromStore(err, nil)
		}
	}
	return nil
}
