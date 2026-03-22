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

// HandleEventUpload maps proto events to domain write types and ingests them.
func (service *Service) HandleEventUpload(
	ctx context.Context,
	machineID uuid.UUID,
	request *syncv1.EventUploadRequest,
) (*syncv1.EventUploadResponse, error) {
	executionEvents, err := mapExecutionEvents(request.GetEvents(), service.eventAllowlist)
	if err != nil {
		return nil, err
	}

	fileAccessEvents, err := mapFileAccessEvents(request.GetFileAccessEvents())
	if err != nil {
		return nil, err
	}

	err = service.dataStore.IngestEvents(ctx, machineID, executionEvents, fileAccessEvents)
	if err != nil {
		return nil, err
	}

	return syncv1.EventUploadResponse_builder{}.Build(), nil
}

func mapExecutionEvents(
	events []*syncv1.Event,
	allowlist map[domain.EventDecision]struct{},
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
		if !inAllowlist(allowlist, decision) {
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

func mapDecision(value syncv1.Decision) (domain.EventDecision, error) {
	switch value {
	case syncv1.Decision_DECISION_UNKNOWN:
		return domain.EventDecisionUnknown, nil
	case syncv1.Decision_ALLOW_UNKNOWN:
		return domain.EventDecisionAllowUnknown, nil
	case syncv1.Decision_ALLOW_BINARY:
		return domain.EventDecisionAllowBinary, nil
	case syncv1.Decision_ALLOW_CERTIFICATE:
		return domain.EventDecisionAllowCertificate, nil
	case syncv1.Decision_ALLOW_SCOPE:
		return domain.EventDecisionAllowScope, nil
	case syncv1.Decision_ALLOW_TEAMID:
		return domain.EventDecisionAllowTeamID, nil
	case syncv1.Decision_ALLOW_SIGNINGID:
		return domain.EventDecisionAllowSigningID, nil
	case syncv1.Decision_ALLOW_CDHASH:
		return domain.EventDecisionAllowCDHash, nil
	case syncv1.Decision_BLOCK_UNKNOWN:
		return domain.EventDecisionBlockUnknown, nil
	case syncv1.Decision_BLOCK_BINARY:
		return domain.EventDecisionBlockBinary, nil
	case syncv1.Decision_BLOCK_CERTIFICATE:
		return domain.EventDecisionBlockCertificate, nil
	case syncv1.Decision_BLOCK_SCOPE:
		return domain.EventDecisionBlockScope, nil
	case syncv1.Decision_BLOCK_TEAMID:
		return domain.EventDecisionBlockTeamID, nil
	case syncv1.Decision_BLOCK_SIGNINGID:
		return domain.EventDecisionBlockSigningID, nil
	case syncv1.Decision_BLOCK_CDHASH:
		return domain.EventDecisionBlockCDHash, nil
	case syncv1.Decision_BUNDLE_BINARY:
		return domain.EventDecisionBundleBinary, nil
	default:
		return "", fmt.Errorf("unsupported decision %q", value)
	}
}

func mapFileAccessDecision(value syncv1.FileAccessDecision) (domain.FileAccessDecision, error) {
	switch value {
	case syncv1.FileAccessDecision_FILE_ACCESS_DECISION_UNKNOWN:
		return domain.FileAccessDecisionUnknown, nil
	case syncv1.FileAccessDecision_FILE_ACCESS_DECISION_DENIED:
		return domain.FileAccessDecisionDenied, nil
	case syncv1.FileAccessDecision_FILE_ACCESS_DECISION_DENIED_INVALID_SIGNATURE:
		return domain.FileAccessDecisionDeniedInvalidSignature, nil
	case syncv1.FileAccessDecision_FILE_ACCESS_DECISION_AUDIT_ONLY:
		return domain.FileAccessDecisionAuditOnly, nil
	default:
		return "", fmt.Errorf("unsupported file access decision %q", value)
	}
}

func inAllowlist(allowlist map[domain.EventDecision]struct{}, decision domain.EventDecision) bool {
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
	records := make([]domain.SigningChainEntry, 0, len(certificates))
	for _, certificate := range certificates {
		if certificate == nil {
			continue
		}
		records = append(records, domain.SigningChainEntry{
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

func protoTime(seconds float64) *time.Time {
	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 1) {
		return nil
	}
	wholeSeconds := int64(seconds)
	nanos := int64((seconds - float64(wholeSeconds)) * float64(time.Second))
	resolved := time.Unix(wholeSeconds, nanos).UTC()
	return &resolved
}

func normalizeStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
