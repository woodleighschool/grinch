package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/woodleighschool/grinch/backend/internal/models"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) UpsertUserByExternal(ctx context.Context, externalID, principal, displayName, email string) (*models.User, error) {
	return s.UpsertCloudUserByExternal(ctx, externalID, principal, displayName, email)
}

func (s *Store) UpsertCloudUserByExternal(ctx context.Context, externalID, principal, displayName, email string) (*models.User, error) {
	// First, handle potential conflicts with local users that have the same principal name
	const checkConflictQ = `
		UPDATE users 
		SET external_id = $1,
		    display_name = $2,
		    email = $3,
		    user_type = 'cloud',
		    synced_at = NOW(),
		    updated_at = NOW()
		WHERE principal_name = $4 AND user_type = 'local' AND is_protected_local = false
		RETURNING id, external_id, principal_name, display_name, email, user_type, synced_at, created_at, updated_at;
	`

	row := s.pool.QueryRow(ctx, checkConflictQ, externalID, displayName, email, principal)
	var (
		user        models.User
		userTypeStr string
		dbDisplay   sql.NullString
		dbEmail     sql.NullString
	)

	// Try to convert existing local user first
	if err := row.Scan(
		&user.ID,
		&user.ExternalID,
		&user.PrincipalName,
		&dbDisplay,
		&dbEmail,
		&userTypeStr,
		&user.SyncedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err == nil {
		// Successfully converted local user to cloud user
		user.UserType = models.UserType(userTypeStr)
		setUserStrings(&user, dbDisplay, dbEmail)
		return &user, nil
	}

	// No local user to convert, proceed with normal upsert
	const q = `
		INSERT INTO users (external_id, principal_name, display_name, email, user_type, synced_at, updated_at)
		VALUES ($1, $2, $3, $4, 'cloud', NOW(), NOW())
		ON CONFLICT (external_id) DO UPDATE
			SET principal_name = EXCLUDED.principal_name,
			    display_name = EXCLUDED.display_name,
			    email = EXCLUDED.email,
			    user_type = 'cloud',
			    synced_at = NOW(),
			    updated_at = NOW()
		RETURNING id,
		          external_id,
		          principal_name,
		          display_name,
		          email,
		          user_type,
		          synced_at,
		          created_at,
		          updated_at;
	`

	row = s.pool.QueryRow(ctx, q, externalID, principal, displayName, email)
	if err := row.Scan(
		&user.ID,
		&user.ExternalID,
		&user.PrincipalName,
		&dbDisplay,
		&dbEmail,
		&userTypeStr,
		&user.SyncedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	user.UserType = models.UserType(userTypeStr)
	setUserStrings(&user, dbDisplay, dbEmail)
	return &user, nil
}

func (s *Store) UpsertLocalUser(ctx context.Context, principal, displayName, machineID string) (*models.User, error) {
	const q = `
		INSERT INTO users (principal_name, display_name, user_type, created_at, updated_at)
		VALUES ($1, NULLIF($2, ''), 'local', NOW(), NOW())
		ON CONFLICT (principal_name) DO UPDATE
			SET display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), users.display_name),
			    external_id = NULL,
			    user_type = 'local',
			    updated_at = NOW()
		WHERE users.user_type = 'local'
		RETURNING id,
		          external_id,
		          principal_name,
		          display_name,
		          email,
		          user_type,
		          synced_at,
		          created_at,
		          updated_at;
	`

	row := s.pool.QueryRow(ctx, q, principal, displayName)
	var (
		user        models.User
		userTypeStr string
		dbDisplay   sql.NullString
		dbEmail     sql.NullString
	)
	if err := row.Scan(
		&user.ID,
		&user.ExternalID,
		&user.PrincipalName,
		&dbDisplay,
		&dbEmail,
		&userTypeStr,
		&user.SyncedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, lookupErr := s.UserByPrincipalName(ctx, principal)
			if lookupErr != nil {
				return nil, lookupErr
			}
			if existing == nil {
				return nil, errors.New("failed to upsert local user")
			}
			return existing, nil
		}
		return nil, err
	}
	user.UserType = models.UserType(userTypeStr)
	setUserStrings(&user, dbDisplay, dbEmail)

	// Insert local user metadata
	const metaQ = `
		INSERT INTO local_user_metadata (user_id, santa_agent_machine_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE
			SET santa_agent_machine_id = COALESCE(EXCLUDED.santa_agent_machine_id, local_user_metadata.santa_agent_machine_id),
			    updated_at = NOW();
	`
	if _, err := s.pool.Exec(ctx, metaQ, user.ID, machineID); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const q = `
		SELECT id,
		       external_id,
		       principal_name,
		       display_name,
		       email,
		       user_type,
		       is_protected_local,
		       synced_at,
		       created_at,
		       updated_at
		FROM users
		WHERE id = $1;
	`
	row := s.pool.QueryRow(ctx, q, id)
	var (
		user        models.User
		userTypeStr string
		dbDisplay   sql.NullString
		dbEmail     sql.NullString
	)
	if err := row.Scan(
		&user.ID,
		&user.ExternalID,
		&user.PrincipalName,
		&dbDisplay,
		&dbEmail,
		&userTypeStr,
		&user.IsProtectedLocal,
		&user.SyncedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	user.UserType = models.UserType(userTypeStr)
	setUserStrings(&user, dbDisplay, dbEmail)
	return &user, nil
}

func (s *Store) UserByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	const q = `
		SELECT id,
		       external_id,
		       principal_name,
		       display_name,
		       email,
		       user_type,
		       synced_at,
		       created_at,
		       updated_at
		FROM users
		WHERE external_id = $1;
	`
	row := s.pool.QueryRow(ctx, q, externalID)
	var (
		user        models.User
		userTypeStr string
		dbDisplay   sql.NullString
		dbEmail     sql.NullString
	)
	if err := row.Scan(
		&user.ID,
		&user.ExternalID,
		&user.PrincipalName,
		&dbDisplay,
		&dbEmail,
		&userTypeStr,
		&user.SyncedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	user.UserType = models.UserType(userTypeStr)
	setUserStrings(&user, dbDisplay, dbEmail)
	return &user, nil
}

func (s *Store) UserByPrincipalName(ctx context.Context, principal string) (*models.User, error) {
	const q = `
		SELECT id,
		       external_id,
		       principal_name,
		       display_name,
		       email,
		       user_type,
		       synced_at,
		       created_at,
		       updated_at
		FROM users
		WHERE principal_name = $1;
	`
	row := s.pool.QueryRow(ctx, q, principal)
	var (
		user        models.User
		userTypeStr string
		dbDisplay   sql.NullString
		dbEmail     sql.NullString
	)
	if err := row.Scan(
		&user.ID,
		&user.ExternalID,
		&user.PrincipalName,
		&dbDisplay,
		&dbEmail,
		&userTypeStr,
		&user.SyncedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	user.UserType = models.UserType(userTypeStr)
	setUserStrings(&user, dbDisplay, dbEmail)
	return &user, nil
}

func (s *Store) ListLocalUsers(ctx context.Context) ([]*models.User, error) {
	const q = `
		SELECT id,
		       external_id,
		       principal_name,
		       display_name,
		       email,
		       user_type,
		       synced_at,
		       created_at,
		       updated_at
		FROM users
		WHERE user_type = 'local'
		ORDER BY principal_name;
	`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var (
			user        models.User
			userTypeStr string
			dbDisplay   sql.NullString
			dbEmail     sql.NullString
		)
		if err := rows.Scan(
			&user.ID,
			&user.ExternalID,
			&user.PrincipalName,
			&dbDisplay,
			&dbEmail,
			&userTypeStr,
			&user.SyncedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.UserType = models.UserType(userTypeStr)
		setUserStrings(&user, dbDisplay, dbEmail)
		users = append(users, &user)
	}
	return users, rows.Err()
}

func (s *Store) ConvertLocalToCloudUser(ctx context.Context, userID uuid.UUID, externalID, displayName, email string) error {
	const q = `
		UPDATE users 
		SET external_id = $2,
		    display_name = $3,
		    email = $4,
		    user_type = 'cloud',
		    synced_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1 AND user_type = 'local' AND is_protected_local = false;
	`

	result, err := s.pool.Exec(ctx, q, userID, externalID, displayName, email)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found, not a local user, or is a protected local user")
	}

	// Update local user metadata to mark conversion check
	const metaQ = `
		UPDATE local_user_metadata 
		SET last_converted_check = NOW(),
		    updated_at = NOW()
		WHERE user_id = $1;
	`
	_, _ = s.pool.Exec(ctx, metaQ, userID)

	return nil
}

func (s *Store) GetCloudUsersNotSyncedSince(ctx context.Context, since time.Time) ([]*models.User, error) {
	const q = `
		SELECT id,
		       external_id,
		       principal_name,
		       display_name,
		       email,
		       user_type,
		       synced_at,
		       created_at,
		       updated_at
		FROM users
		WHERE user_type = 'cloud'
		  AND (synced_at IS NULL OR synced_at < $1)
		ORDER BY principal_name;
	`
	rows, err := s.pool.Query(ctx, q, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var (
			user        models.User
			userTypeStr string
			dbDisplay   sql.NullString
			dbEmail     sql.NullString
		)
		if err := rows.Scan(
			&user.ID,
			&user.ExternalID,
			&user.PrincipalName,
			&dbDisplay,
			&dbEmail,
			&userTypeStr,
			&user.SyncedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.UserType = models.UserType(userTypeStr)
		setUserStrings(&user, dbDisplay, dbEmail)
		users = append(users, &user)
	}
	return users, rows.Err()
}

func (s *Store) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM users WHERE id = $1;`
	result, err := s.pool.Exec(ctx, q, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (s *Store) UpsertGroup(ctx context.Context, externalID, displayName, description string) (*models.Group, error) {
	const q = `
		INSERT INTO groups (external_id, display_name, description, synced_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (external_id) DO UPDATE
			SET display_name = EXCLUDED.display_name,
			    description = EXCLUDED.description,
			    synced_at = NOW(),
			    updated_at = NOW()
		RETURNING id, external_id, display_name, description, synced_at, created_at, updated_at;
	`
	row := s.pool.QueryRow(ctx, q, externalID, displayName, description)
	var (
		group         models.Group
		dbDescription sql.NullString
	)
	if err := row.Scan(
		&group.ID,
		&group.ExternalID,
		&group.DisplayName,
		&dbDescription,
		&group.SyncedAt,
		&group.CreatedAt,
		&group.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if dbDescription.Valid {
		group.Description = dbDescription.String
	}
	return &group, nil
}

func (s *Store) ListGroups(ctx context.Context) ([]models.Group, error) {
	const q = `
		SELECT id, external_id, display_name, description, synced_at, created_at, updated_at
		FROM groups ORDER BY display_name ASC;
	`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var (
			group         models.Group
			dbDescription sql.NullString
		)
		if err := rows.Scan(&group.ID, &group.ExternalID, &group.DisplayName, &dbDescription, &group.SyncedAt, &group.CreatedAt, &group.UpdatedAt); err != nil {
			return nil, err
		}
		if dbDescription.Valid {
			group.Description = dbDescription.String
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (s *Store) ListUsers(ctx context.Context, limit int) ([]models.User, error) {
	baseQuery := `
		SELECT id,
		       external_id,
		       principal_name,
		       display_name,
		       email,
		       user_type,
		       synced_at,
		       created_at,
		       updated_at
		FROM users
		ORDER BY display_name NULLS LAST, principal_name
	`
	var (
		rows pgx.Rows
		err  error
	)

	if limit > 0 {
		rows, err = s.pool.Query(ctx, baseQuery+" LIMIT $1", limit)
	} else {
		rows, err = s.pool.Query(ctx, baseQuery)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var (
			user        models.User
			userTypeStr string
			dbDisplay   sql.NullString
			dbEmail     sql.NullString
		)
		if err := rows.Scan(
			&user.ID,
			&user.ExternalID,
			&user.PrincipalName,
			&dbDisplay,
			&dbEmail,
			&userTypeStr,
			&user.SyncedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.UserType = models.UserType(userTypeStr)
		setUserStrings(&user, dbDisplay, dbEmail)
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Store) GroupsForUser(ctx context.Context, userID uuid.UUID) ([]models.Group, error) {
	const q = `
		SELECT g.id,
		       g.external_id,
		       g.display_name,
		       g.description,
		       g.synced_at,
		       g.created_at,
		       g.updated_at
		FROM group_memberships gm
		JOIN groups g ON gm.group_id = g.id
		WHERE gm.user_id = $1
		ORDER BY g.display_name;
	`
	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var (
			group        models.Group
			dbDesc       sql.NullString
			dbExternalID string
		)
		if err := rows.Scan(
			&group.ID,
			&dbExternalID,
			&group.DisplayName,
			&dbDesc,
			&group.SyncedAt,
			&group.CreatedAt,
			&group.UpdatedAt,
		); err != nil {
			return nil, err
		}
		group.ExternalID = dbExternalID
		if dbDesc.Valid {
			group.Description = dbDesc.String
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (s *Store) ListGroupMemberships(ctx context.Context) ([]models.GroupMembership, error) {
	const q = `
		SELECT group_id, user_id
		FROM group_memberships
		ORDER BY group_id, user_id;
	`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberships := make([]models.GroupMembership, 0)
	for rows.Next() {
		var membership models.GroupMembership
		if err := rows.Scan(&membership.GroupID, &membership.UserID); err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}
	return memberships, rows.Err()
}

func (s *Store) HostsForUser(ctx context.Context, userID uuid.UUID) ([]models.Host, error) {
	const q = `
		SELECT
			h.id,
			h.hostname,
			h.serial_number,
			h.machine_id,
			h.primary_user_id,
			h.last_seen,
			h.created_at,
			h.updated_at,
			h.os_version,
			h.os_build,
			h.model_identifier,
			h.santa_version,
			h.client_mode,
			u.principal_name,
			u.display_name
		FROM hosts h
		LEFT JOIN users u ON u.id = h.primary_user_id
		WHERE h.primary_user_id = $1
		ORDER BY h.hostname NULLS LAST;
	`
	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []models.Host
	for rows.Next() {
		var (
			host        models.Host
			clientMode  sql.NullString
			principal   sql.NullString
			displayName sql.NullString
			lastSeen    sql.NullTime
		)
		if err := rows.Scan(
			&host.ID,
			&host.Hostname,
			&host.SerialNumber,
			&host.MachineID,
			&host.PrimaryUserID,
			&lastSeen,
			&host.CreatedAt,
			&host.UpdatedAt,
			&host.OSVersion,
			&host.OSBuild,
			&host.ModelIdentifier,
			&host.SantaVersion,
			&clientMode,
			&principal,
			&displayName,
		); err != nil {
			return nil, err
		}
		if clientMode.Valid {
			host.ClientMode = models.ClientMode(clientMode.String)
		}
		if principal.Valid {
			host.PrimaryUserPrincipal = principal.String
		}
		if displayName.Valid {
			host.PrimaryUserDisplayName = displayName.String
		}
		if lastSeen.Valid {
			ls := lastSeen.Time
			host.LastSeen = &ls
		}
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}

func (s *Store) RecentUserEvents(ctx context.Context, userID uuid.UUID, limit int) ([]models.UserEvent, error) {
	if limit <= 0 {
		limit = 10
	}
	const q = `
		SELECT
			be.id,
			be.host_id,
			h.hostname,
			be.application_id,
			be.process_path,
			be.blocked_reason,
			be.decision,
			be.occurred_at
		FROM blocked_events be
		LEFT JOIN hosts h ON h.id = be.host_id
		WHERE be.user_id = $1
		ORDER BY be.occurred_at DESC
		LIMIT $2;
	`
	rows, err := s.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.UserEvent, 0)
	for rows.Next() {
		var (
			event       models.UserEvent
			hostname    sql.NullString
			blockReason sql.NullString
			decision    sql.NullString
		)
		if err := rows.Scan(
			&event.ID,
			&event.HostID,
			&hostname,
			&event.ApplicationID,
			&event.ProcessPath,
			&blockReason,
			&decision,
			&event.OccurredAt,
		); err != nil {
			return nil, err
		}
		if hostname.Valid {
			event.Hostname = hostname.String
		}
		if blockReason.Valid {
			event.BlockedReason = blockReason.String
		}
		if decision.Valid {
			event.Decision = decision.String
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) UserPoliciesForUser(ctx context.Context, userID uuid.UUID, groupIDs []uuid.UUID) ([]models.UserPolicy, error) {
	const userQ = `
		SELECT
			s.id,
			s.application_id,
			a.name,
			a.rule_type,
			a.identifier,
			s.action,
			s.created_at,
			u.display_name,
			u.principal_name
		FROM application_scopes s
		JOIN applications a ON a.id = s.application_id
		JOIN users u ON u.id = s.target_id
		WHERE s.target_type = 'user' AND s.target_id = $1
		ORDER BY a.name, s.created_at DESC;
	`
	rows, err := s.pool.Query(ctx, userQ, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make([]models.UserPolicy, 0)
	for rows.Next() {
		var (
			policy      models.UserPolicy
			displayName sql.NullString
			principal   string
		)
		if err := rows.Scan(
			&policy.ScopeID,
			&policy.ApplicationID,
			&policy.ApplicationName,
			&policy.RuleType,
			&policy.Identifier,
			&policy.Action,
			&policy.CreatedAt,
			&displayName,
			&principal,
		); err != nil {
			return nil, err
		}
		policy.TargetType = "user"
		policy.TargetID = userID
		if displayName.Valid {
			policy.TargetName = displayName.String
		} else {
			policy.TargetName = principal
		}
		policy.ViaGroup = false
		policies = append(policies, policy)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(groupIDs) == 0 {
		return policies, nil
	}

	const groupQ = `
		SELECT
			s.id,
			s.application_id,
			a.name,
			a.rule_type,
			a.identifier,
			s.action,
			s.created_at,
			s.target_id,
			g.display_name
		FROM application_scopes s
		JOIN applications a ON a.id = s.application_id
		JOIN groups g ON g.id = s.target_id
		WHERE s.target_type = 'group' AND s.target_id = ANY($1::uuid[])
		ORDER BY a.name, s.created_at DESC;
	`
	groupRows, err := s.pool.Query(ctx, groupQ, groupIDs)
	if err != nil {
		return nil, err
	}
	defer groupRows.Close()

	for groupRows.Next() {
		var (
			policy   models.UserPolicy
			targetID uuid.UUID
			name     sql.NullString
		)
		if err := groupRows.Scan(
			&policy.ScopeID,
			&policy.ApplicationID,
			&policy.ApplicationName,
			&policy.RuleType,
			&policy.Identifier,
			&policy.Action,
			&policy.CreatedAt,
			&targetID,
			&name,
		); err != nil {
			return nil, err
		}
		policy.TargetType = "group"
		policy.TargetID = targetID
		if name.Valid {
			policy.TargetName = name.String
		}
		policy.ViaGroup = true
		policies = append(policies, policy)
	}
	return policies, groupRows.Err()
}

func (s *Store) ReplaceGroupMemberships(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx) // Ignore rollback errors; transaction may be committed
	}()

	if _, err := tx.Exec(ctx, `DELETE FROM group_memberships WHERE group_id = $1`, groupID); err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for _, userID := range userIDs {
		batch.Queue(`INSERT INTO group_memberships (group_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, groupID, userID)
	}

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) ListApplications(ctx context.Context) ([]models.Application, error) {
	const q = `
		SELECT id, name, rule_type, identifier, description, enabled, created_at, updated_at
		FROM applications ORDER BY created_at DESC;
	`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var (
			app           models.Application
			dbDescription sql.NullString
		)
		if err := rows.Scan(
			&app.ID,
			&app.Name,
			&app.RuleType,
			&app.Identifier,
			&dbDescription,
			&app.Enabled,
			&app.CreatedAt,
			&app.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if dbDescription.Valid {
			app.Description = dbDescription.String
		}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

func (s *Store) GetApplicationByIdentifier(ctx context.Context, identifier string) (*models.Application, error) {
	const q = `
		SELECT id, name, rule_type, identifier, description, enabled, created_at, updated_at
		FROM applications WHERE identifier = $1;
	`
	var (
		app           models.Application
		dbDescription sql.NullString
	)
	row := s.pool.QueryRow(ctx, q, identifier)
	if err := row.Scan(
		&app.ID,
		&app.Name,
		&app.RuleType,
		&app.Identifier,
		&dbDescription,
		&app.Enabled,
		&app.CreatedAt,
		&app.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if dbDescription.Valid {
		app.Description = dbDescription.String
	}
	return &app, nil
}

func (s *Store) CreateApplication(ctx context.Context, app models.Application) (*models.Application, error) {
	const q = `
		INSERT INTO applications (name, rule_type, identifier, description, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at;
	`
	// Default enabled to true if not set
	if !app.Enabled {
		app.Enabled = true
	}
	row := s.pool.QueryRow(ctx, q, app.Name, app.RuleType, app.Identifier, app.Description, app.Enabled)
	if err := row.Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt); err != nil {
		return nil, err
	}
	return &app, nil
}

func (s *Store) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM applications WHERE id = $1`
	commandTag, err := s.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) UpdateApplication(ctx context.Context, id uuid.UUID, enabled bool) (*models.Application, error) {
	const q = `
		UPDATE applications 
		SET enabled = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, rule_type, identifier, description, enabled, created_at, updated_at;
	`
	var (
		app           models.Application
		dbDescription sql.NullString
	)
	row := s.pool.QueryRow(ctx, q, id, enabled)
	if err := row.Scan(
		&app.ID,
		&app.Name,
		&app.RuleType,
		&app.Identifier,
		&dbDescription,
		&app.Enabled,
		&app.CreatedAt,
		&app.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if dbDescription.Valid {
		app.Description = dbDescription.String
	}
	return &app, nil
}

func (s *Store) AddApplicationScope(ctx context.Context, scope models.ApplicationScope) (*models.ApplicationScope, error) {
	const q = `
		INSERT INTO application_scopes (application_id, target_type, target_id, action)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;
	`
	row := s.pool.QueryRow(ctx, q, scope.ApplicationID, scope.TargetType, scope.TargetID, scope.Action)
	if err := row.Scan(&scope.ID, &scope.CreatedAt); err != nil {
		return nil, err
	}
	return &scope, nil
}

func (s *Store) DeleteApplicationScope(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM application_scopes WHERE id = $1`
	commandTag, err := s.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListApplicationScopes(ctx context.Context, appID uuid.UUID) ([]models.ApplicationScope, error) {
	const q = `
		SELECT id, application_id, target_type, target_id, action, created_at
		FROM application_scopes
		WHERE application_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := s.pool.Query(ctx, q, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scopes []models.ApplicationScope
	for rows.Next() {
		var scope models.ApplicationScope
		if err := rows.Scan(
			&scope.ID,
			&scope.ApplicationID,
			&scope.TargetType,
			&scope.TargetID,
			&scope.Action,
			&scope.CreatedAt,
		); err != nil {
			return nil, err
		}
		scopes = append(scopes, scope)
	}
	return scopes, rows.Err()
}

func (s *Store) InsertBlockedEvent(ctx context.Context, event models.BlockedEvent) (*models.BlockedEvent, error) {
	const q = `
		INSERT INTO blocked_events (
			host_id, user_id, application_id, process_path, process_hash, signer, blocked_reason, event_payload, occurred_at, ingested_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, NOW()), NOW())
		RETURNING id, occurred_at, ingested_at;
	`
	var occurredAt time.Time
	if !event.OccurredAt.IsZero() {
		occurredAt = event.OccurredAt
	}
	row := s.pool.QueryRow(
		ctx,
		q,
		event.HostID,
		event.UserID,
		event.ApplicationID,
		event.ProcessPath,
		event.ProcessHash,
		event.Signer,
		event.BlockedReason,
		event.EventPayload,
		sqlNullTime(occurredAt),
	)
	if err := row.Scan(&event.ID, &event.OccurredAt, &event.IngestedAt); err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *Store) UpsertHost(ctx context.Context, host models.Host) (*models.Host, error) {
	const q = `
		INSERT INTO hosts (machine_id, hostname, serial_number, primary_user_id, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), NOW())
		ON CONFLICT (machine_id) DO UPDATE
			SET hostname = EXCLUDED.hostname,
			    serial_number = EXCLUDED.serial_number,
			    primary_user_id = EXCLUDED.primary_user_id,
			    last_seen = NOW(),
			    updated_at = NOW()
		RETURNING id, last_seen, created_at, updated_at;
	`
	row := s.pool.QueryRow(ctx, q, host.MachineID, host.Hostname, host.SerialNumber, host.PrimaryUserID)
	if err := row.Scan(&host.ID, &host.LastSeen, &host.CreatedAt, &host.UpdatedAt); err != nil {
		return nil, err
	}
	return &host, nil
}

func (s *Store) HostByMachineID(ctx context.Context, machineID string) (*models.Host, error) {
	const q = `
		SELECT id, hostname, serial_number, machine_id, primary_user_id, last_seen,
		       created_at, updated_at, os_version, os_build, model_identifier,
		       santa_version, client_mode
		FROM hosts WHERE machine_id = $1;
	`
	row := s.pool.QueryRow(ctx, q, machineID)
	var host models.Host
	var clientMode sql.NullString
	if err := row.Scan(
		&host.ID,
		&host.Hostname,
		&host.SerialNumber,
		&host.MachineID,
		&host.PrimaryUserID,
		&host.LastSeen,
		&host.CreatedAt,
		&host.UpdatedAt,
		&host.OSVersion,
		&host.OSBuild,
		&host.ModelIdentifier,
		&host.SantaVersion,
		&clientMode,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if clientMode.Valid {
		host.ClientMode = models.ClientMode(clientMode.String)
	}
	return &host, nil
}

func (s *Store) ApplicationScopesForTargets(ctx context.Context, targetIDs []uuid.UUID) ([]models.ApplicationScope, error) {
	if len(targetIDs) == 0 {
		return []models.ApplicationScope{}, nil
	}
	const q = `
		SELECT id, application_id, target_type, target_id, action, created_at
		FROM application_scopes
		WHERE (target_type = 'group' AND target_id = ANY($1))
		   OR (target_type = 'user' AND target_id = ANY($1))
	`
	rows, err := s.pool.Query(ctx, q, targetIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scopes []models.ApplicationScope
	for rows.Next() {
		var scope models.ApplicationScope
		if err := rows.Scan(&scope.ID, &scope.ApplicationID, &scope.TargetType, &scope.TargetID, &scope.Action, &scope.CreatedAt); err != nil {
			return nil, err
		}
		scopes = append(scopes, scope)
	}
	return scopes, rows.Err()
}

func (s *Store) ApplicationsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]models.Application, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]models.Application{}, nil
	}
	const q = `
		SELECT id, name, rule_type, identifier, description, enabled, created_at, updated_at
		FROM applications
		WHERE id = ANY($1);
	`
	rows, err := s.pool.Query(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]models.Application, len(ids))
	for rows.Next() {
		var (
			app           models.Application
			dbDescription sql.NullString
		)
		if err := rows.Scan(&app.ID, &app.Name, &app.RuleType, &app.Identifier, &dbDescription, &app.Enabled, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, err
		}
		if dbDescription.Valid {
			app.Description = dbDescription.String
		}
		result[app.ID] = app
	}
	return result, rows.Err()
}

func (s *Store) RecentBlockedEvents(ctx context.Context, limit int) ([]models.BlockedEvent, error) {
	const q = `
		SELECT id, host_id, user_id, application_id, process_path, process_hash, signer, blocked_reason, event_payload, occurred_at, ingested_at
		FROM blocked_events
		ORDER BY occurred_at DESC
		LIMIT $1;
	`
	rows, err := s.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.BlockedEvent
	for rows.Next() {
		var event models.BlockedEvent
		if err := rows.Scan(
			&event.ID,
			&event.HostID,
			&event.UserID,
			&event.ApplicationID,
			&event.ProcessPath,
			&event.ProcessHash,
			&event.Signer,
			&event.BlockedReason,
			&event.EventPayload,
			&event.OccurredAt,
			&event.IngestedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) UserGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	const q = `
		SELECT group_id FROM group_memberships WHERE user_id = $1;
	`
	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) GroupsByExternalIDs(ctx context.Context, externalIDs []string) (map[string]uuid.UUID, error) {
	if len(externalIDs) == 0 {
		return map[string]uuid.UUID{}, nil
	}
	const q = `
		SELECT external_id, id FROM groups WHERE external_id = ANY($1);
	`
	rows, err := s.pool.Query(ctx, q, externalIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]uuid.UUID, len(externalIDs))
	for rows.Next() {
		var external string
		var id uuid.UUID
		if err := rows.Scan(&external, &id); err != nil {
			return nil, err
		}
		result[external] = id
	}
	return result, rows.Err()
}

func (s *Store) ListHosts(ctx context.Context) ([]models.Host, error) {
	const q = `
		SELECT
			h.id,
			h.hostname,
			h.serial_number,
			h.machine_id,
			h.primary_user_id,
			h.last_seen,
			h.created_at,
			h.updated_at,
			h.os_version,
			h.os_build,
			h.model_identifier,
			h.santa_version,
			h.client_mode,
			u.principal_name,
			u.display_name
		FROM hosts h
		LEFT JOIN users u ON u.id = h.primary_user_id
		ORDER BY h.hostname NULLS LAST;
	`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hosts := make([]models.Host, 0)
	for rows.Next() {
		var host models.Host
		var clientMode, principal, displayName sql.NullString
		if err := rows.Scan(
			&host.ID,
			&host.Hostname,
			&host.SerialNumber,
			&host.MachineID,
			&host.PrimaryUserID,
			&host.LastSeen,
			&host.CreatedAt,
			&host.UpdatedAt,
			&host.OSVersion,
			&host.OSBuild,
			&host.ModelIdentifier,
			&host.SantaVersion,
			&clientMode,
			&principal,
			&displayName,
		); err != nil {
			return nil, err
		}
		if clientMode.Valid {
			host.ClientMode = models.ClientMode(clientMode.String)
		}
		if principal.Valid {
			host.PrimaryUserPrincipal = principal.String
		}
		if displayName.Valid {
			host.PrimaryUserDisplayName = displayName.String
		}
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}

func sqlNullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func setUserStrings(user *models.User, display, email sql.NullString) {
	if display.Valid {
		user.DisplayName = display.String
	} else {
		user.DisplayName = ""
	}
	if email.Valid {
		user.Email = email.String
	} else {
		user.Email = ""
	}
}

// Role management methods
func (s *Store) EnsureUserHasRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	// Get role group ID
	roleGroupID, err := s.getRoleGroupID(ctx, roleName)
	if err != nil {
		return fmt.Errorf("get role group: %w", err)
	}

	const q = `
		INSERT INTO user_role_assignments (user_id, role_group_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_group_id) DO NOTHING;
	`
	_, err = s.pool.Exec(ctx, q, userID, roleGroupID)
	return err
}

func (s *Store) getRoleGroupID(ctx context.Context, roleName string) (uuid.UUID, error) {
	const q = `SELECT id FROM role_groups WHERE name = $1;`
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, q, roleName).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("role group %s not found", roleName)
	}
	return id, err
}

func (s *Store) UserHasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	roleGroupID, err := s.getRoleGroupID(ctx, roleName)
	if err != nil {
		return false, err
	}

	const q = `
		SELECT EXISTS(
			SELECT 1 FROM user_role_assignments 
			WHERE user_id = $1 AND role_group_id = $2
		);
	`
	var exists bool
	err = s.pool.QueryRow(ctx, q, userID, roleGroupID).Scan(&exists)
	return exists, err
}

func (s *Store) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	const q = `
		SELECT rg.name
		FROM user_role_assignments ura
		JOIN role_groups rg ON ura.role_group_id = rg.id
		WHERE ura.user_id = $1
		ORDER BY rg.name;
	`
	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (s *Store) RemoveUserRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	roleGroupID, err := s.getRoleGroupID(ctx, roleName)
	if err != nil {
		return err
	}

	const q = `
		DELETE FROM user_role_assignments 
		WHERE user_id = $1 AND role_group_id = $2;
	`
	_, err = s.pool.Exec(ctx, q, userID, roleGroupID)
	return err
}

// EnsureInitialAdminUser creates or updates an initial admin user based on provided configuration
func (s *Store) EnsureInitialAdminUser(ctx context.Context, password string) error {
	if password == "" {
		return fmt.Errorf("password is required for initial admin user")
	}

	const adminPrincipal = "admin"
	const adminDisplayName = "Master Claus"

	// Hash the password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create or update the user as a protected local user with password
	const upsertUserQuery = `
		INSERT INTO users (principal_name, display_name, email, user_type, password_hash, is_protected_local, created_at, updated_at)
		VALUES ($1, $2, NULL, 'local', $3, true, NOW(), NOW())
		ON CONFLICT (principal_name) DO UPDATE
			SET display_name = EXCLUDED.display_name,
			    email = NULL,
			    password_hash = EXCLUDED.password_hash,
			    user_type = 'local',
			    is_protected_local = true,
			    updated_at = NOW()
		WHERE users.is_protected_local = true  -- Only update if it's already a protected local user
		RETURNING id;
	`

	var userID uuid.UUID
	err = s.pool.QueryRow(ctx, upsertUserQuery, adminPrincipal, adminDisplayName, string(passwordHash)).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to create/update initial admin user: %w", err)
	}

	// Ensure the user has admin role
	if err := s.EnsureUserHasRole(ctx, userID, "admin"); err != nil {
		return fmt.Errorf("failed to assign admin role to initial user: %w", err)
	}

	return nil
}

// AuthenticateLocalUser validates a local user's credentials
func (s *Store) AuthenticateLocalUser(ctx context.Context, principal, password string) (*models.User, error) {
	const q = `
		SELECT id,
		       external_id,
		       principal_name,
		       display_name,
		       email,
		       user_type,
		       password_hash,
		       is_protected_local,
		       synced_at,
		       created_at,
		       updated_at
		FROM users
		WHERE principal_name = $1 AND user_type = 'local' AND password_hash IS NOT NULL;
	`
	row := s.pool.QueryRow(ctx, q, principal)

	var (
		user         models.User
		userTypeStr  string
		passwordHash string
		dbDisplay    sql.NullString
		dbEmail      sql.NullString
	)

	if err := row.Scan(
		&user.ID,
		&user.ExternalID,
		&user.PrincipalName,
		&dbDisplay,
		&dbEmail,
		&userTypeStr,
		&passwordHash,
		&user.IsProtectedLocal,
		&user.SyncedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, nil // Invalid password
	}

	user.UserType = models.UserType(userTypeStr)
	setUserStrings(&user, dbDisplay, dbEmail)
	return &user, nil
}
