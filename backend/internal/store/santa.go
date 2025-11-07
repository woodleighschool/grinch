package store

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/backend/internal/models"
)

// UpsertSantaSyncState creates or updates sync state for a machine
func (s *Store) UpsertSantaSyncState(ctx context.Context, machineID string, state models.SyncState) (*models.SyncState, error) {
	const q = `
		INSERT INTO santa_sync_state (
			machine_id, last_sync_time, last_sync_type, rules_delivered, rules_processed,
			rule_count_hash, binary_rule_hash, certificate_rule_hash, teamid_rule_hash,
			signingid_rule_hash, cdhash_rule_hash, transitive_rule_hash, compiler_rule_hash,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
		ON CONFLICT (machine_id) DO UPDATE SET
			last_sync_time = EXCLUDED.last_sync_time,
			last_sync_type = EXCLUDED.last_sync_type,
			rules_delivered = EXCLUDED.rules_delivered,
			rules_processed = EXCLUDED.rules_processed,
			rule_count_hash = EXCLUDED.rule_count_hash,
			binary_rule_hash = EXCLUDED.binary_rule_hash,
			certificate_rule_hash = EXCLUDED.certificate_rule_hash,
			teamid_rule_hash = EXCLUDED.teamid_rule_hash,
			signingid_rule_hash = EXCLUDED.signingid_rule_hash,
			cdhash_rule_hash = EXCLUDED.cdhash_rule_hash,
			transitive_rule_hash = EXCLUDED.transitive_rule_hash,
			compiler_rule_hash = EXCLUDED.compiler_rule_hash,
			updated_at = NOW()
		RETURNING id, machine_id, last_sync_time, last_sync_type, rules_delivered, rules_processed,
			rule_count_hash, binary_rule_hash, certificate_rule_hash, teamid_rule_hash,
			signingid_rule_hash, cdhash_rule_hash, transitive_rule_hash, compiler_rule_hash,
			created_at, updated_at;
	`

	row := s.pool.QueryRow(ctx, q,
		machineID, state.LastSyncTime, string(state.LastSyncType), state.RulesDelivered, state.RulesProcessed,
		state.RuleCountHash, state.BinaryRuleHash, state.CertificateRuleHash, state.TeamIDRuleHash,
		state.SigningIDRuleHash, state.CDHashRuleHash, state.TransitiveRuleHash, state.CompilerRuleHash,
	)

	var result models.SyncState
	var syncType string
	if err := row.Scan(
		&result.ID, &result.MachineID, &result.LastSyncTime, &syncType, &result.RulesDelivered, &result.RulesProcessed,
		&result.RuleCountHash, &result.BinaryRuleHash, &result.CertificateRuleHash, &result.TeamIDRuleHash,
		&result.SigningIDRuleHash, &result.CDHashRuleHash, &result.TransitiveRuleHash, &result.CompilerRuleHash,
		&result.CreatedAt, &result.UpdatedAt,
	); err != nil {
		return nil, err
	}

	result.LastSyncType = models.SyncType(syncType)
	return &result, nil
}

// GetSantaSyncState retrieves sync state for a machine
func (s *Store) GetSantaSyncState(ctx context.Context, machineID string) (*models.SyncState, error) {
	const q = `
		SELECT id, machine_id, last_sync_time, last_sync_type, rules_delivered, rules_processed,
			rule_count_hash, binary_rule_hash, certificate_rule_hash, teamid_rule_hash,
			signingid_rule_hash, cdhash_rule_hash, transitive_rule_hash, compiler_rule_hash,
			created_at, updated_at
		FROM santa_sync_state
		WHERE machine_id = $1;
	`

	row := s.pool.QueryRow(ctx, q, machineID)
	var state models.SyncState
	var syncType string
	if err := row.Scan(
		&state.ID, &state.MachineID, &state.LastSyncTime, &syncType, &state.RulesDelivered, &state.RulesProcessed,
		&state.RuleCountHash, &state.BinaryRuleHash, &state.CertificateRuleHash, &state.TeamIDRuleHash,
		&state.SigningIDRuleHash, &state.CDHashRuleHash, &state.TransitiveRuleHash, &state.CompilerRuleHash,
		&state.CreatedAt, &state.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	state.LastSyncType = models.SyncType(syncType)
	return &state, nil
}

// CreateSantaRuleCursor creates a pagination cursor for rule download
func (s *Store) CreateSantaRuleCursor(ctx context.Context, machineID string, lastRuleID *uuid.UUID) (string, error) {
	// Create a unique cursor token
	h := sha256.New()
	if _, err := fmt.Fprintf(h, "%s:%d", machineID, time.Now().UnixNano()); err != nil {
		return "", fmt.Errorf("failed to write cursor data: %w", err)
	}
	cursorToken := fmt.Sprintf("%x", h.Sum(nil))[:32]

	const q = `
		INSERT INTO santa_rule_cursors (machine_id, cursor_token, last_rule_id, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING cursor_token;
	`

	expiresAt := time.Now().Add(30 * time.Minute) // Cursors expire in 30 minutes
	var token string
	if err := s.pool.QueryRow(ctx, q, machineID, cursorToken, lastRuleID, expiresAt).Scan(&token); err != nil {
		return "", err
	}

	return token, nil
}

// GetSantaRuleCursor retrieves cursor information
func (s *Store) GetSantaRuleCursor(ctx context.Context, cursorToken string) (machineID string, lastRuleID *uuid.UUID, err error) {
	const q = `
		SELECT machine_id, last_rule_id
		FROM santa_rule_cursors
		WHERE cursor_token = $1 AND expires_at > NOW();
	`

	row := s.pool.QueryRow(ctx, q, cursorToken)
	if err := row.Scan(&machineID, &lastRuleID); err != nil {
		if err == pgx.ErrNoRows {
			return "", nil, fmt.Errorf("invalid or expired cursor")
		}
		return "", nil, err
	}

	return machineID, lastRuleID, nil
}

// UpdateSantaRuleCursor updates the last rule ID for an existing cursor
func (s *Store) UpdateSantaRuleCursor(ctx context.Context, cursorToken string, lastRuleID uuid.UUID) error {
	const q = `
		UPDATE santa_rule_cursors
		SET last_rule_id = $2, expires_at = $3
		WHERE cursor_token = $1;
	`

	expiresAt := time.Now().Add(30 * time.Minute)
	_, err := s.pool.Exec(ctx, q, cursorToken, lastRuleID, expiresAt)
	return err
}

// DeleteSantaRuleCursor removes a cursor
func (s *Store) DeleteSantaRuleCursor(ctx context.Context, cursorToken string) error {
	const q = `DELETE FROM santa_rule_cursors WHERE cursor_token = $1;`
	_, err := s.pool.Exec(ctx, q, cursorToken)
	return err
}

// GetSantaRulesForMachine retrieves rules for a specific machine with pagination
func (s *Store) GetSantaRulesForMachine(ctx context.Context, machineID string, lastRuleID *uuid.UUID, limit int) ([]models.SantaRule, error) {
	// First get the user and groups for this machine
	host, err := s.HostByMachineID(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("get host: %w", err)
	}
	if host == nil {
		return []models.SantaRule{}, nil
	}

	targetIDs := make([]uuid.UUID, 0)
	if host.PrimaryUserID != nil {
		targetIDs = append(targetIDs, *host.PrimaryUserID)

		// Get user's groups
		groupIDs, err := s.UserGroups(ctx, *host.PrimaryUserID)
		if err != nil {
			return nil, fmt.Errorf("get user groups: %w", err)
		}
		targetIDs = append(targetIDs, groupIDs...)
	}

	if len(targetIDs) == 0 {
		return []models.SantaRule{}, nil
	}

	// Build query with pagination
	var q string
	var args []interface{}

	if lastRuleID == nil {
		q = `
			SELECT a.id, a.identifier, a.rule_type, s.action, a.description, a.created_at
			FROM applications a
			JOIN application_scopes s ON a.id = s.application_id
			WHERE a.enabled = TRUE
			  AND ((s.target_type = 'user' AND s.target_id = ANY($1))
			   OR (s.target_type = 'group' AND s.target_id = ANY($1)))
			ORDER BY a.id
			LIMIT $2;
		`
		args = []interface{}{targetIDs, limit}
	} else {
		q = `
			SELECT a.id, a.identifier, a.rule_type, s.action, a.description, a.created_at
			FROM applications a
			JOIN application_scopes s ON a.id = s.application_id
			WHERE a.enabled = TRUE
			  AND ((s.target_type = 'user' AND s.target_id = ANY($1))
			    OR (s.target_type = 'group' AND s.target_id = ANY($1)))
			  AND a.id > $2
			ORDER BY a.id
			LIMIT $3;
		`
		args = []interface{}{targetIDs, *lastRuleID, limit}
	}

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.SantaRule
	for rows.Next() {
		var appID uuid.UUID
		var identifier, ruleType, action, description string
		var createdAt time.Time

		if err := rows.Scan(&appID, &identifier, &ruleType, &action, &description, &createdAt); err != nil {
			return nil, err
		}

		policy := models.PolicyAllowlist
		if action == "block" {
			policy = models.PolicyBlocklist
		}

		rule := models.SantaRule{
			Identifier:   identifier,
			RuleType:     models.RuleType(ruleType),
			Policy:       policy,
			CustomMsg:    description,
			CreationTime: float64(createdAt.Unix()) + float64(createdAt.Nanosecond())/1e9,
		}

		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// InsertSantaEvent inserts a Santa event into blocked_events
func (s *Store) InsertSantaEvent(ctx context.Context, machineID string, event models.SantaEvent) error {
	// Get host by machine ID
	host, err := s.HostByMachineID(ctx, machineID)
	if err != nil {
		return fmt.Errorf("get host: %w", err)
	}
	if host == nil {
		return fmt.Errorf("unknown machine: %s", machineID)
	}

	// Try to find user by executing_user, create as local if not found
	var userID *uuid.UUID
	if event.ExecutingUser != "" {
		user, err := s.UserByUsername(ctx, event.ExecutingUser)
		if err == nil && user != nil {
			userID = &user.ID
		} else {
			// User not found, create as local user
			localUser, err := s.UpsertLocalUser(ctx, event.ExecutingUser, "", machineID)
			if err == nil {
				userID = &localUser.ID
			}
			// If local user creation fails, continue without user ID
		}
	}

	const q = `
		INSERT INTO blocked_events (
			host_id, user_id, process_path, process_hash, 
			file_sha256, file_name, executing_user, execution_time,
			logged_in_users, current_sessions, decision,
			file_bundle_id, file_bundle_path, file_bundle_name, file_bundle_version,
			file_bundle_hash, file_bundle_binary_count,
			pid, ppid, parent_name,
			quarantine_data_url, quarantine_referer_url, quarantine_timestamp, quarantine_agent_bundle_id,
			signing_chain, signing_id, team_id, cdhash,
			occurred_at, ingested_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, NOW()
		);
	`

	var executionTime *time.Time
	if event.ExecutionTime > 0 {
		t := time.Unix(int64(event.ExecutionTime), int64((event.ExecutionTime-float64(int64(event.ExecutionTime)))*1e9))
		executionTime = &t
	}

	var quarantineTimestamp *time.Time
	if event.QuarantineTimestamp > 0 {
		t := time.Unix(event.QuarantineTimestamp, 0)
		quarantineTimestamp = &t
	}

	_, err = s.pool.Exec(ctx, q,
		host.ID, userID, event.FilePath, event.FileSHA256,
		event.FileSHA256, event.FileName, event.ExecutingUser, executionTime,
		event.LoggedInUsers, event.CurrentSessions, string(event.Decision),
		event.FileBundleID, event.FileBundlePath, event.FileBundleName, event.FileBundleVersion,
		event.FileBundleHash, event.FileBundleBinaryCount,
		event.PID, event.PPID, event.ParentName,
		event.QuarantineDataURL, event.QuarantineRefererURL, quarantineTimestamp, event.QuarantineAgentBundleID,
		event.SigningChain, event.SigningID, event.TeamID, event.CDHash,
		executionTime,
	)

	return err
}

// RecordSantaRuleDelivery tracks rule delivery to a machine
func (s *Store) RecordSantaRuleDelivery(ctx context.Context, machineID string, rule models.SantaRule) error {
	// Create hash of rule for deduplication
	h := sha256.New()
	if _, err := fmt.Fprintf(h, "%s:%s:%s", rule.RuleType, rule.Identifier, rule.Policy); err != nil {
		return fmt.Errorf("failed to write rule data for hash: %w", err)
	}
	ruleHash := fmt.Sprintf("%x", h.Sum(nil))

	const q = `
		INSERT INTO santa_rule_deliveries (machine_id, rule_hash, rule_type, rule_identifier, policy)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (machine_id, rule_hash) DO NOTHING;
	`

	_, err := s.pool.Exec(ctx, q, machineID, ruleHash, string(rule.RuleType), rule.Identifier, string(rule.Policy))
	return err
}

// CleanupExpiredSantaCursors removes expired cursors
func (s *Store) CleanupExpiredSantaCursors(ctx context.Context) error {
	const q = `DELETE FROM santa_rule_cursors WHERE expires_at < NOW();`
	_, err := s.pool.Exec(ctx, q)
	return err
}

// UpsertHostWithMachineID updates host with Santa information using provided machine_id
func (s *Store) UpsertHostWithMachineID(ctx context.Context, machineID string, req models.PreflightRequest, primaryUserID *uuid.UUID) error {
	const q = `
		INSERT INTO hosts (
			machine_id, hostname, serial_number, primary_user_id,
			os_version, os_build, model_identifier, santa_version, client_mode,
			last_seen, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE(NULLIF($9, ''), 'MONITOR'), NOW(), NOW(), NOW())
		ON CONFLICT (machine_id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			serial_number = EXCLUDED.serial_number,
			primary_user_id = EXCLUDED.primary_user_id,
			os_version = EXCLUDED.os_version,
			os_build = EXCLUDED.os_build,
			model_identifier = EXCLUDED.model_identifier,
			santa_version = EXCLUDED.santa_version,
			client_mode = EXCLUDED.client_mode,
			last_seen = NOW(),
			updated_at = NOW();
	`

	_, err := s.pool.Exec(ctx, q,
		machineID, // Use the actual machine_id from URL parameter
		req.Hostname,
		req.SerialNum,
		primaryUserID,
		req.OSVersion,
		req.OSBuild,
		req.ModelIdentifier,
		req.SantaVersion,
		string(req.ClientMode),
	)

	return err
}
