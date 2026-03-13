package santa

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

type signingChainRecord struct {
	CommonName         string    `json:"common_name"`
	Organization       string    `json:"organization"`
	OrganizationalUnit string    `json:"organizational_unit"`
	SHA256             string    `json:"sha256"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
}

type fileAccessProcessRecord struct {
	Position     int32     `json:"position"`
	Pid          int32     `json:"pid"`
	FilePath     string    `json:"file_path"`
	ExecutableID uuid.UUID `json:"executable_id"`
}

func (store *Store) UpsertMachine(ctx context.Context, machine appsanta.MachineUpsert) error {
	_, err := store.store.Queries().UpsertMachine(ctx, db.UpsertMachineParams{
		MachineID:            machine.MachineID,
		SerialNumber:         machine.SerialNumber,
		Hostname:             machine.Hostname,
		ModelIdentifier:      machine.ModelIdentifier,
		OsVersion:            machine.OSVersion,
		OsBuild:              machine.OSBuild,
		SantaVersion:         machine.SantaVersion,
		PrimaryUser:          machine.PrimaryUser,
		PrimaryUserGroupsRaw: machine.PrimaryUserGroupsRaw,
		LastSeenAt:           machine.LastSeenAt,
	})
	return err
}

func (store *Store) IngestEvents(
	ctx context.Context,
	machineID uuid.UUID,
	events []*syncv1.Event,
	fileAccessEvents []*syncv1.FileAccessEvent,
	allowlist map[domain.EventDecision]struct{},
) (int, error) {
	var ingested int

	runErr := store.store.RunInTx(ctx, func(queries *db.Queries) error {
		executionIngested, executionErr := ingestExecutionEvents(ctx, queries, machineID, events, allowlist)
		if executionErr != nil {
			return executionErr
		}

		fileAccessIngested, fileAccessErr := ingestFileAccessEvents(ctx, queries, machineID, fileAccessEvents)
		if fileAccessErr != nil {
			return fileAccessErr
		}

		ingested = executionIngested + fileAccessIngested

		return nil
	})
	if runErr != nil {
		return 0, runErr
	}

	return ingested, nil
}

func (store *Store) DeleteEventsBefore(ctx context.Context, createdAt time.Time) (int64, error) {
	deletedExecution, executionErr := store.store.Queries().DeleteExecutionEventsBefore(ctx, createdAt)
	if executionErr != nil {
		return 0, executionErr
	}

	deletedFileAccess, fileAccessErr := store.store.Queries().DeleteFileAccessEventsBefore(ctx, createdAt)
	if fileAccessErr != nil {
		return 0, fileAccessErr
	}

	return deletedExecution + deletedFileAccess, nil
}

func ingestExecutionEvents(
	ctx context.Context,
	queries *db.Queries,
	machineID uuid.UUID,
	events []*syncv1.Event,
	allowlist map[domain.EventDecision]struct{},
) (int, error) {
	var ingested int

	for _, event := range events {
		if event == nil {
			continue
		}

		decision, decisionErr := mapDecision(event.GetDecision())
		if decisionErr != nil {
			return 0, decisionErr
		}

		if !shouldIngestDecision(allowlist, decision) {
			continue
		}

		executableID, executableErr := getOrCreateEventExecutable(ctx, queries, event)
		if executableErr != nil {
			return 0, executableErr
		}

		if createErr := createExecutionEvent(ctx, queries, machineID, executableID, event, decision); createErr != nil {
			return 0, createErr
		}

		ingested++
	}

	return ingested, nil
}

func ingestFileAccessEvents(
	ctx context.Context,
	queries *db.Queries,
	machineID uuid.UUID,
	events []*syncv1.FileAccessEvent,
) (int, error) {
	var ingested int

	for _, event := range events {
		if event == nil {
			continue
		}

		if createErr := createFileAccessEvent(ctx, queries, machineID, event); createErr != nil {
			return 0, createErr
		}

		ingested++
	}

	return ingested, nil
}

func getOrCreateEventExecutable(
	ctx context.Context,
	queries *db.Queries,
	event *syncv1.Event,
) (uuid.UUID, error) {
	entitlementsJSON, entitlementsErr := marshalEntitlements(event.GetEntitlementInfo())
	if entitlementsErr != nil {
		return uuid.Nil, entitlementsErr
	}

	signingChainJSON, signingChainErr := marshalSigningChain(event.GetSigningChain())
	if signingChainErr != nil {
		return uuid.Nil, signingChainErr
	}

	id, idErr := newUUID()
	if idErr != nil {
		return uuid.Nil, idErr
	}

	row, err := queries.GetOrCreateEventExecutable(ctx, db.GetOrCreateEventExecutableParams{
		ID:             id,
		FileSha256:     event.GetFileSha256(),
		FileName:       event.GetFileName(),
		FileBundleID:   event.GetFileBundleId(),
		FileBundlePath: event.GetFileBundlePath(),
		SigningID:      event.GetSigningId(),
		TeamID:         event.GetTeamId(),
		Cdhash:         event.GetCdhash(),
		Entitlements:   entitlementsJSON,
		SigningChain:   signingChainJSON,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("get or create event executable: %w", err)
	}

	return row.ID, nil
}

func createExecutionEvent(
	ctx context.Context,
	queries *db.Queries,
	machineID uuid.UUID,
	executableID uuid.UUID,
	event *syncv1.Event,
	decision domain.EventDecision,
) error {
	eventID, idErr := newUUID()
	if idErr != nil {
		return idErr
	}

	_, err := queries.CreateExecutionEvent(ctx, db.CreateExecutionEventParams{
		ID:              eventID,
		MachineID:       machineID,
		ExecutableID:    executableID,
		Decision:        string(decision),
		FilePath:        event.GetFilePath(),
		ExecutingUser:   event.GetExecutingUser(),
		LoggedInUsers:   normalizeStringSlice(event.GetLoggedInUsers()),
		CurrentSessions: normalizeStringSlice(event.GetCurrentSessions()),
		OccurredAt:      executionTime(event.GetExecutionTime()),
	})
	if err != nil {
		return fmt.Errorf("create execution event: %w", err)
	}

	return nil
}

func createFileAccessEvent(
	ctx context.Context,
	queries *db.Queries,
	machineID uuid.UUID,
	event *syncv1.FileAccessEvent,
) error {
	decision, decisionErr := mapFileAccessDecision(event.GetDecision())
	if decisionErr != nil {
		return decisionErr
	}

	processChain := make([]fileAccessProcessRecord, 0, len(event.GetProcessChain()))
	var primaryExecutableID *uuid.UUID
	for index, process := range event.GetProcessChain() {
		if process == nil {
			continue
		}

		executableID, executableErr := getOrCreateProcessExecutable(ctx, queries, process)
		if executableErr != nil {
			return executableErr
		}

		if primaryExecutableID == nil {
			primaryExecutableID = &executableID
		}

		processChain = append(processChain, fileAccessProcessRecord{
			Position:     int32(index),
			Pid:          process.GetPid(),
			FilePath:     process.GetFilePath(),
			ExecutableID: executableID,
		})
	}

	processChainJSON, processChainErr := json.Marshal(processChain)
	if processChainErr != nil {
		return fmt.Errorf("marshal file access process chain: %w", processChainErr)
	}

	eventID, idErr := newUUID()
	if idErr != nil {
		return idErr
	}

	_, err := queries.CreateFileAccessEvent(ctx, db.CreateFileAccessEventParams{
		ID:           eventID,
		MachineID:    machineID,
		ExecutableID: nullableUUID(primaryExecutableID),
		RuleVersion:  event.GetRuleVersion(),
		RuleName:     event.GetRuleName(),
		Target:       event.GetTarget(),
		Decision:     string(decision),
		ProcessChain: processChainJSON,
		OccurredAt:   executionTime(event.GetAccessTime()),
	})
	if err != nil {
		return fmt.Errorf("create file access event: %w", err)
	}

	return nil
}

func getOrCreateProcessExecutable(
	ctx context.Context,
	queries *db.Queries,
	process *syncv1.Process,
) (uuid.UUID, error) {
	signingChainJSON, signingChainErr := marshalSigningChain(process.GetSigningChain())
	if signingChainErr != nil {
		return uuid.Nil, signingChainErr
	}

	id, idErr := newUUID()
	if idErr != nil {
		return uuid.Nil, idErr
	}

	row, err := queries.GetOrCreateProcessExecutable(ctx, db.GetOrCreateProcessExecutableParams{
		ID:           id,
		FileSha256:   process.GetFileSha256(),
		FilePath:     process.GetFilePath(),
		SigningID:    process.GetSigningId(),
		TeamID:       process.GetTeamId(),
		Cdhash:       process.GetCdhash(),
		SigningChain: signingChainJSON,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("get or create process executable: %w", err)
	}

	return row.ID, nil
}

func nullableUUID(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}

	return pgtype.UUID{Bytes: *value, Valid: true}
}

func marshalEntitlements(info *syncv1.EntitlementInfo) ([]byte, error) {
	if info == nil {
		return []byte("{}"), nil
	}

	entitlements := make(map[string]any, len(info.GetEntitlements()))
	for _, entitlement := range info.GetEntitlements() {
		if entitlement == nil {
			continue
		}

		key := entitlement.GetKey()
		if key == "" {
			continue
		}

		rawValue := entitlement.GetValue()
		if rawValue == "" {
			entitlements[key] = nil
			continue
		}

		var decodedValue any
		if err := json.Unmarshal([]byte(rawValue), &decodedValue); err != nil {
			entitlements[key] = rawValue
			continue
		}

		entitlements[key] = decodedValue
	}

	encoded, err := json.Marshal(entitlements)
	if err != nil {
		return nil, fmt.Errorf("marshal entitlements: %w", err)
	}

	return encoded, nil
}

func marshalSigningChain(certificates []*syncv1.Certificate) ([]byte, error) {
	records := make([]signingChainRecord, 0, len(certificates))
	for _, certificate := range certificates {
		if certificate == nil {
			continue
		}

		records = append(records, signingChainRecord{
			CommonName:         certificate.GetCn(),
			Organization:       certificate.GetOrg(),
			OrganizationalUnit: certificate.GetOu(),
			SHA256:             certificate.GetSha256(),
			ValidFrom:          time.Unix(int64(certificate.GetValidFrom()), 0).UTC(),
			ValidUntil:         time.Unix(int64(certificate.GetValidUntil()), 0).UTC(),
		})
	}

	encoded, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("marshal signing chain: %w", err)
	}

	return encoded, nil
}
