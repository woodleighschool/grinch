package machines

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	dblisting "github.com/woodleighschool/grinch/internal/store/listing"
)

func listMachines(ctx context.Context, pool *pgxpool.Pool, query listing.Query) ([]machines.ListItem, int64, error) {
	cfg := dblisting.Config{
		Table: "machines",
		SelectCols: []string{
			"id", "serial_number", "hostname", "model_identifier", "os_version",
			"primary_user", "user_id", "last_seen", "policy_id",
			"applied_policy_id", "applied_settings_version", "applied_rules_version", "policy_status",
		},
		Columns: map[string]string{
			"id":            "id",
			"serial_number": "serial_number",
			"hostname":      "hostname",
			"model":         "model_identifier",
			"os_version":    "os_version",
			"primary_user":  "primary_user",
			"user_id":       "user_id",
			"last_seen":     "last_seen",
			"policy_id":     "policy_id",
			"policy_status": "policy_status",
		},
		SearchColumns: []string{"serial_number", "hostname", "primary_user"},
		DefaultSort:   listing.Sort{Field: "last_seen", Desc: true},
	}
	return dblisting.List(ctx, pool, cfg, query, scanMachineListItem)
}

func scanMachineListItem(rows pgx.Rows) (machines.ListItem, error) {
	var (
		item               machines.ListItem
		primaryUser        pgtype.Text
		userID             *uuid.UUID
		lastSeen           time.Time
		policyID           *uuid.UUID
		appliedPolicyID    *uuid.UUID
		appliedSettingsVer pgtype.Int4
		appliedRulesVer    pgtype.Int4
		policyStatus       int16
	)

	err := rows.Scan(
		&item.ID, &item.SerialNumber, &item.Hostname, &item.Model, &item.OSVersion,
		&primaryUser, &userID, &lastSeen, &policyID,
		&appliedPolicyID, &appliedSettingsVer, &appliedRulesVer, &policyStatus,
	)
	if err != nil {
		return item, err
	}

	item.PrimaryUser = pgconv.TextVal(primaryUser)
	item.UserID = userID
	item.LastSeen = lastSeen
	item.PolicyID = policyID
	item.AppliedPolicyID = appliedPolicyID
	item.AppliedSettingsVersion = pgconv.Int32Val(appliedSettingsVer)
	item.AppliedRulesVersion = pgconv.Int32Val(appliedRulesVer)
	item.PolicyStatus = machines.PolicyStatus(policyStatus)

	return item, nil
}
