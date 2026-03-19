package pgutil

import (
	"context"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

func GetGroup(ctx context.Context, queries *db.Queries, id uuid.UUID) (domain.Group, error) {
	row, err := queries.GetGroup(ctx, id)
	if err != nil {
		return domain.Group{}, err
	}

	source, err := ToSource(row.Source)
	if err != nil {
		return domain.Group{}, err
	}

	return domain.Group{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Source:      source,
		MemberCount: row.MemberCount,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func GetUser(ctx context.Context, queries *db.Queries, id uuid.UUID) (domain.User, error) {
	row, err := queries.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	source, err := ToSource(row.Source)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:          row.ID,
		UPN:         row.Upn,
		DisplayName: row.DisplayName,
		Source:      source,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func GetMachine(ctx context.Context, queries *db.Queries, id uuid.UUID) (domain.Machine, error) {
	row, err := queries.GetMachine(ctx, id)
	if err != nil {
		return domain.Machine{}, err
	}

	return domain.Machine{
		ID:                   row.MachineID,
		SerialNumber:         row.SerialNumber,
		Hostname:             row.Hostname,
		ModelIdentifier:      row.ModelIdentifier,
		OSVersion:            row.OsVersion,
		OSBuild:              row.OsBuild,
		SantaVersion:         row.SantaVersion,
		PrimaryUser:          row.PrimaryUser,
		PrimaryUserID:        row.PrimaryUserID,
		RuleSyncStatus:       domain.DeriveMachineRuleSyncStatus(row.PendingPreflightAt, row.LastRuleSyncAttemptAt),
		ClientMode:           domain.ParseMachineClientMode(row.ClientMode),
		BinaryRuleCount:      row.BinaryRuleCount,
		CertificateRuleCount: row.CertificateRuleCount,
		CompilerRuleCount:    row.CompilerRuleCount,
		TransitiveRuleCount:  row.TransitiveRuleCount,
		TeamIDRuleCount:      row.TeamidRuleCount,
		SigningIDRuleCount:   row.SigningidRuleCount,
		CDHashRuleCount:      row.CdhashRuleCount,
		LastSeenAt:           row.LastSeenAt,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}
