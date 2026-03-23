package santa

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
)

var eventDecisionMap = map[syncv1.Decision]domain.ExecutionDecision{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
	syncv1.Decision_DECISION_UNKNOWN:  domain.ExecutionDecisionUnknown,
	syncv1.Decision_ALLOW_UNKNOWN:     domain.ExecutionDecisionAllowUnknown,
	syncv1.Decision_ALLOW_BINARY:      domain.ExecutionDecisionAllowBinary,
	syncv1.Decision_ALLOW_CERTIFICATE: domain.ExecutionDecisionAllowCertificate,
	syncv1.Decision_ALLOW_SCOPE:       domain.ExecutionDecisionAllowScope,
	syncv1.Decision_ALLOW_TEAMID:      domain.ExecutionDecisionAllowTeamID,
	syncv1.Decision_ALLOW_SIGNINGID:   domain.ExecutionDecisionAllowSigningID,
	syncv1.Decision_ALLOW_CDHASH:      domain.ExecutionDecisionAllowCDHash,
	syncv1.Decision_BLOCK_UNKNOWN:     domain.ExecutionDecisionBlockUnknown,
	syncv1.Decision_BLOCK_BINARY:      domain.ExecutionDecisionBlockBinary,
	syncv1.Decision_BLOCK_CERTIFICATE: domain.ExecutionDecisionBlockCertificate,
	syncv1.Decision_BLOCK_SCOPE:       domain.ExecutionDecisionBlockScope,
	syncv1.Decision_BLOCK_TEAMID:      domain.ExecutionDecisionBlockTeamID,
	syncv1.Decision_BLOCK_SIGNINGID:   domain.ExecutionDecisionBlockSigningID,
	syncv1.Decision_BLOCK_CDHASH:      domain.ExecutionDecisionBlockCDHash,
	syncv1.Decision_BUNDLE_BINARY:     domain.ExecutionDecisionBundleBinary,
}

var fileAccessDecisionMap = map[syncv1.FileAccessDecision]domain.FileAccessDecision{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
	syncv1.FileAccessDecision_FILE_ACCESS_DECISION_UNKNOWN:                  domain.FileAccessDecisionUnknown,
	syncv1.FileAccessDecision_FILE_ACCESS_DECISION_DENIED:                   domain.FileAccessDecisionDenied,
	syncv1.FileAccessDecision_FILE_ACCESS_DECISION_DENIED_INVALID_SIGNATURE: domain.FileAccessDecisionDeniedInvalidSignature,
	syncv1.FileAccessDecision_FILE_ACCESS_DECISION_AUDIT_ONLY:               domain.FileAccessDecisionAuditOnly,
}

func (s *Service) HandleEventUpload(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.EventUploadRequest,
) (*syncv1.EventUploadResponse, error) {
	executionEvents, err := mapExecutionEvents(req.GetEvents(), s.eventAllowlist)
	if err != nil {
		return nil, err
	}

	fileAccessEvents, err := mapFileAccessEvents(req.GetFileAccessEvents())
	if err != nil {
		return nil, err
	}

	if err = s.dataStore.IngestEvents(ctx, machineID, executionEvents, fileAccessEvents); err != nil {
		return nil, err
	}

	return syncv1.EventUploadResponse_builder{}.Build(), nil
}

func mapExecutionEvents(
	events []*syncv1.Event,
	allowlist map[domain.ExecutionDecision]struct{},
) ([]model.ExecutionEventWrite, error) {
	writes := make([]model.ExecutionEventWrite, 0, len(events))

	for _, event := range events {
		if event == nil {
			continue
		}

		decision, err := mapDecision(event.GetDecision())
		if err != nil {
			return nil, err
		}
		if !isAllowedDecision(allowlist, decision) {
			continue
		}

		entitlements, err := marshalEntitlements(event.GetEntitlementInfo())
		if err != nil {
			return nil, err
		}

		signingChain, err := marshalSigningChain(event.GetSigningChain())
		if err != nil {
			return nil, err
		}

		writes = append(writes, model.ExecutionEventWrite{
			Executable: model.ExecutableWrite{
				FileSHA256:     event.GetFileSha256(),
				FileName:       event.GetFileName(),
				FileBundleID:   event.GetFileBundleId(),
				FileBundlePath: event.GetFileBundlePath(),
				SigningID:      event.GetSigningId(),
				TeamID:         event.GetTeamId(),
				CDHash:         event.GetCdhash(),
				Entitlements:   entitlements,
				SigningChain:   signingChain,
			},
			FilePath:        event.GetFilePath(),
			ExecutingUser:   event.GetExecutingUser(),
			LoggedInUsers:   normalizeStrings(event.GetLoggedInUsers()),
			CurrentSessions: normalizeStrings(event.GetCurrentSessions()),
			Decision:        decision,
			OccurredAt:      protoTime(event.GetExecutionTime()),
		})
	}

	return writes, nil
}

func mapFileAccessEvents(events []*syncv1.FileAccessEvent) ([]model.FileAccessEventWrite, error) {
	writes := make([]model.FileAccessEventWrite, 0, len(events))

	for _, event := range events {
		if event == nil {
			continue
		}

		decision, err := mapFileAccessDecision(event.GetDecision())
		if err != nil {
			return nil, err
		}

		processes, err := mapProcessChain(event.GetProcessChain())
		if err != nil {
			return nil, err
		}

		writes = append(writes, model.FileAccessEventWrite{
			RuleVersion: event.GetRuleVersion(),
			RuleName:    event.GetRuleName(),
			Target:      event.GetTarget(),
			Decision:    decision,
			Processes:   processes,
			OccurredAt:  protoTime(event.GetAccessTime()),
		})
	}

	return writes, nil
}

func mapProcessChain(processes []*syncv1.Process) ([]model.ProcessWrite, error) {
	writes := make([]model.ProcessWrite, 0, len(processes))

	for _, process := range processes {
		if process == nil {
			continue
		}

		signingChain, err := marshalSigningChain(process.GetSigningChain())
		if err != nil {
			return nil, err
		}

		writes = append(writes, model.ProcessWrite{
			Pid:          process.GetPid(),
			FilePath:     process.GetFilePath(),
			FileSHA256:   process.GetFileSha256(),
			SigningID:    process.GetSigningId(),
			TeamID:       process.GetTeamId(),
			CDHash:       process.GetCdhash(),
			SigningChain: signingChain,
		})
	}

	return writes, nil
}

func mapDecision(value syncv1.Decision) (domain.ExecutionDecision, error) {
	decision, ok := eventDecisionMap[value]
	if !ok {
		return "", fmt.Errorf("unsupported decision %q", value)
	}

	return decision, nil
}

func mapFileAccessDecision(value syncv1.FileAccessDecision) (domain.FileAccessDecision, error) {
	decision, ok := fileAccessDecisionMap[value]
	if !ok {
		return "", fmt.Errorf("unsupported file access decision %q", value)
	}

	return decision, nil
}

func isAllowedDecision(allowlist map[domain.ExecutionDecision]struct{}, decision domain.ExecutionDecision) bool {
	if len(allowlist) == 0 {
		return true
	}

	_, ok := allowlist[decision]
	return ok
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

		var value any
		if err := json.Unmarshal([]byte(rawValue), &value); err != nil {
			entitlements[key] = rawValue
			continue
		}

		entitlements[key] = value
	}

	data, err := json.Marshal(entitlements)
	if err != nil {
		return nil, fmt.Errorf("marshal entitlements: %w", err)
	}

	return data, nil
}

func marshalSigningChain(certificates []*syncv1.Certificate) ([]byte, error) {
	entries := make([]domain.SigningChainEntry, 0, len(certificates))
	for _, certificate := range certificates {
		if certificate == nil {
			continue
		}

		entries = append(entries, domain.SigningChainEntry{
			CommonName:         certificate.GetCn(),
			Organization:       certificate.GetOrg(),
			OrganizationalUnit: certificate.GetOu(),
			SHA256:             certificate.GetSha256(),
			ValidFrom:          time.Unix(int64(certificate.GetValidFrom()), 0).UTC(),
			ValidUntil:         time.Unix(int64(certificate.GetValidUntil()), 0).UTC(),
		})
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("marshal signing chain: %w", err)
	}

	return data, nil
}

func normalizeStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func protoTime(seconds float64) *time.Time {
	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return nil
	}

	whole, frac := math.Modf(seconds)
	t := time.Unix(int64(whole), int64(frac*float64(time.Second))).UTC()
	return &t
}
