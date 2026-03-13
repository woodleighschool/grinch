package santa

import (
	"fmt"
	"math"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

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

func shouldIngestDecision(allowlist map[domain.EventDecision]struct{}, decision domain.EventDecision) bool {
	if len(allowlist) == 0 {
		return true
	}

	_, ok := allowlist[decision]
	return ok
}

func newUUID() (uuid.UUID, error) {
	created, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("create uuid: %w", err)
	}
	return created, nil
}

func executionTime(seconds float64) *time.Time {
	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 1) {
		return nil
	}

	wholeSeconds := int64(seconds)
	nanos := int64((seconds - float64(wholeSeconds)) * float64(time.Second))
	resolved := time.Unix(wholeSeconds, nanos).UTC()
	return &resolved
}

func normalizeStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}

	return values
}
