// Package machines provides persistence for machine records.
package machines

import (
	"context"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	coremachines "github.com/woodleighschool/grinch/internal/core/machines"
	"github.com/woodleighschool/grinch/internal/core/policies"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// Repo persists machine records.
type Repo struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo backed by PostgreSQL.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: sqlc.New(pool), pool: pool}
}

// Get returns the machine with the given ID.
func (r *Repo) Get(ctx context.Context, id uuid.UUID) (coremachines.Machine, error) {
	row, err := r.q.GetMachineByID(ctx, id)
	if err != nil {
		return coremachines.Machine{}, coreerrors.FromStore(err, nil)
	}
	return mapMachine(row), nil
}

// Upsert creates or updates a machine and returns the stored record.
func (r *Repo) Upsert(ctx context.Context, mc coremachines.Machine) (coremachines.Machine, error) {
	row, err := r.q.UpsertMachineByID(ctx, toUpsertParams(mc))
	if err != nil {
		return coremachines.Machine{}, coreerrors.FromStore(err, nil)
	}
	return mapMachine(row), nil
}

// List returns machines matching the query.
func (r *Repo) List(ctx context.Context, query listing.Query) ([]coremachines.MachineListItem, listing.Page, error) {
	items, total, err := listMachines(ctx, r.pool, query)
	if err != nil {
		return nil, listing.Page{}, coreerrors.FromStore(err, nil)
	}
	return items, listing.Page{Total: total}, nil
}

// Delete removes a machine by ID.
func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	return coreerrors.FromStore(r.q.DeleteMachineByID(ctx, id), nil)
}

// UpdatePolicyState updates policy assignment metadata for a machine.
func (r *Repo) UpdatePolicyState(
	ctx context.Context,
	id uuid.UUID,
	policyID *uuid.UUID,
	status policies.Status,
) error {
	err := r.q.UpdateMachinePolicyStateByID(ctx, sqlc.UpdateMachinePolicyStateByIDParams{
		ID:           id,
		PolicyID:     policyID,
		PolicyStatus: int16(status),
	})
	return coreerrors.FromStore(err, nil)
}

func mapMachine(row sqlc.Machine) coremachines.Machine {
	return coremachines.Machine{
		ID:                     row.ID,
		SerialNumber:           row.SerialNumber,
		Hostname:               row.Hostname,
		Model:                  row.ModelIdentifier,
		OSVersion:              row.OsVersion,
		OSBuild:                row.OsBuild,
		SantaVersion:           row.SantaVersion,
		PrimaryUser:            pgconv.TextVal(row.PrimaryUser),
		PrimaryUserGroups:      pgconv.TextArray(row.PrimaryUserGroups),
		PushToken:              pgconv.TextVal(row.PushNotificationToken),
		SIPStatus:              row.SipStatus,
		ClientMode:             syncv1.ClientMode(row.ClientMode),
		RequestCleanSync:       row.RequestCleanSync,
		PushNotificationSync:   row.PushNotificationSync,
		BinaryRuleCount:        row.BinaryRuleCount,
		CertificateRuleCount:   row.CertificateRuleCount,
		CompilerRuleCount:      row.CompilerRuleCount,
		TransitiveRuleCount:    row.TransitiveRuleCount,
		TeamIDRuleCount:        row.TeamidRuleCount,
		SigningIDRuleCount:     row.SigningidRuleCount,
		CDHashRuleCount:        row.CdhashRuleCount,
		RulesHash:              pgconv.TextVal(row.RulesHash),
		UserID:                 row.UserID,
		LastSeen:               row.LastSeen.Time,
		PolicyID:               row.PolicyID,
		AppliedPolicyID:        row.AppliedPolicyID,
		AppliedSettingsVersion: pgconv.Int32Val(row.AppliedSettingsVersion),
		AppliedRulesVersion:    pgconv.Int32Val(row.AppliedRulesVersion),
		PolicyStatus:           policies.Status(row.PolicyStatus),
	}
}

func toUpsertParams(mc coremachines.Machine) sqlc.UpsertMachineByIDParams {
	return sqlc.UpsertMachineByIDParams{
		ID:                     mc.ID,
		SerialNumber:           mc.SerialNumber,
		Hostname:               mc.Hostname,
		ModelIdentifier:        mc.Model,
		OsVersion:              mc.OSVersion,
		OsBuild:                mc.OSBuild,
		SantaVersion:           mc.SantaVersion,
		PrimaryUser:            pgconv.TextOrNull(mc.PrimaryUser),
		PrimaryUserGroups:      pgconv.TextArray(mc.PrimaryUserGroups),
		PushNotificationToken:  pgconv.TextOrNull(mc.PushToken),
		SipStatus:              mc.SIPStatus,
		ClientMode:             int32(mc.ClientMode),
		RequestCleanSync:       mc.RequestCleanSync,
		PushNotificationSync:   mc.PushNotificationSync,
		BinaryRuleCount:        mc.BinaryRuleCount,
		CertificateRuleCount:   mc.CertificateRuleCount,
		CompilerRuleCount:      mc.CompilerRuleCount,
		TransitiveRuleCount:    mc.TransitiveRuleCount,
		TeamidRuleCount:        mc.TeamIDRuleCount,
		SigningidRuleCount:     mc.SigningIDRuleCount,
		CdhashRuleCount:        mc.CDHashRuleCount,
		RulesHash:              pgconv.TextOrNull(mc.RulesHash),
		UserID:                 mc.UserID,
		LastSeen:               pgtype.Timestamptz{Time: mc.LastSeen, Valid: true},
		PolicyID:               mc.PolicyID,
		AppliedPolicyID:        mc.AppliedPolicyID,
		AppliedSettingsVersion: pgconv.Int32OrNull(mc.AppliedSettingsVersion),
		AppliedRulesVersion:    pgconv.Int32OrNull(mc.AppliedRulesVersion),
		PolicyStatus:           int16(mc.PolicyStatus),
	}
}
