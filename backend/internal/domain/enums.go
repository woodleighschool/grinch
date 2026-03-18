package domain

import "fmt"

type EventDecision string

const (
	EventDecisionUnknown          EventDecision = "unknown"
	EventDecisionAllowUnknown     EventDecision = "allow_unknown"
	EventDecisionAllowBinary      EventDecision = "allow_binary"
	EventDecisionAllowCertificate EventDecision = "allow_certificate"
	EventDecisionAllowScope       EventDecision = "allow_scope"
	EventDecisionAllowTeamID      EventDecision = "allow_team_id"
	EventDecisionAllowSigningID   EventDecision = "allow_signing_id"
	EventDecisionAllowCDHash      EventDecision = "allow_cd_hash"
	EventDecisionBlockUnknown     EventDecision = "block_unknown"
	EventDecisionBlockBinary      EventDecision = "block_binary"
	EventDecisionBlockCertificate EventDecision = "block_certificate"
	EventDecisionBlockScope       EventDecision = "block_scope"
	EventDecisionBlockTeamID      EventDecision = "block_team_id"
	EventDecisionBlockSigningID   EventDecision = "block_signing_id"
	EventDecisionBlockCDHash      EventDecision = "block_cd_hash"
	EventDecisionBundleBinary     EventDecision = "bundle_binary"
)

type FileAccessDecision string

const (
	FileAccessDecisionUnknown                FileAccessDecision = "unknown"
	FileAccessDecisionDenied                 FileAccessDecision = "denied"
	FileAccessDecisionDeniedInvalidSignature FileAccessDecision = "denied_invalid_signature"
	FileAccessDecisionAuditOnly              FileAccessDecision = "audit_only"
)

func ParsePrincipalSource(value string) (PrincipalSource, error) {
	switch PrincipalSource(value) {
	case PrincipalSourceLocal:
		return PrincipalSourceLocal, nil
	case PrincipalSourceEntra:
		return PrincipalSourceEntra, nil
	default:
		return "", fmt.Errorf("unsupported source %q", value)
	}
}

func ParseMemberKind(value string) (MemberKind, error) {
	switch MemberKind(value) {
	case MemberKindUser:
		return MemberKindUser, nil
	case MemberKindMachine:
		return MemberKindMachine, nil
	default:
		return "", fmt.Errorf("unsupported member kind %q", value)
	}
}

func ParseRuleType(value string) (RuleType, error) {
	switch RuleType(value) {
	case RuleTypeBinary:
		return RuleTypeBinary, nil
	case RuleTypeCertificate:
		return RuleTypeCertificate, nil
	case RuleTypeTeamID:
		return RuleTypeTeamID, nil
	case RuleTypeSigningID:
		return RuleTypeSigningID, nil
	case RuleTypeCDHash:
		return RuleTypeCDHash, nil
	default:
		return "", fmt.Errorf("unsupported rule type %q", value)
	}
}

func ParseRulePolicy(value string) (RulePolicy, error) {
	switch RulePolicy(value) {
	case RulePolicyAllowlist:
		return RulePolicyAllowlist, nil
	case RulePolicyBlocklist:
		return RulePolicyBlocklist, nil
	case RulePolicySilentBlocklist:
		return RulePolicySilentBlocklist, nil
	case RulePolicyCEL:
		return RulePolicyCEL, nil
	default:
		return "", fmt.Errorf("unsupported rule policy %q", value)
	}
}

func ParseRuleTargetAssignment(value string) (RuleTargetAssignment, error) {
	switch RuleTargetAssignment(value) {
	case RuleTargetAssignmentInclude:
		return RuleTargetAssignmentInclude, nil
	case RuleTargetAssignmentExclude:
		return RuleTargetAssignmentExclude, nil
	default:
		return "", fmt.Errorf("unsupported rule target assignment %q", value)
	}
}

func ParseRuleTargetSubjectKind(value string) (RuleTargetSubjectKind, error) {
	switch RuleTargetSubjectKind(value) {
	case RuleTargetSubjectKindGroup:
		return RuleTargetSubjectKindGroup, nil
	case RuleTargetSubjectKindAllDevices:
		return RuleTargetSubjectKindAllDevices, nil
	case RuleTargetSubjectKindAllUsers:
		return RuleTargetSubjectKindAllUsers, nil
	default:
		return "", fmt.Errorf("unsupported rule target subject kind %q", value)
	}
}

func ParseEventDecision(value string) (EventDecision, error) {
	switch EventDecision(value) {
	case EventDecisionUnknown:
		return EventDecisionUnknown, nil
	case EventDecisionAllowUnknown:
		return EventDecisionAllowUnknown, nil
	case EventDecisionAllowBinary:
		return EventDecisionAllowBinary, nil
	case EventDecisionAllowCertificate:
		return EventDecisionAllowCertificate, nil
	case EventDecisionAllowScope:
		return EventDecisionAllowScope, nil
	case EventDecisionAllowTeamID:
		return EventDecisionAllowTeamID, nil
	case EventDecisionAllowSigningID:
		return EventDecisionAllowSigningID, nil
	case EventDecisionAllowCDHash:
		return EventDecisionAllowCDHash, nil
	case EventDecisionBlockUnknown:
		return EventDecisionBlockUnknown, nil
	case EventDecisionBlockBinary:
		return EventDecisionBlockBinary, nil
	case EventDecisionBlockCertificate:
		return EventDecisionBlockCertificate, nil
	case EventDecisionBlockScope:
		return EventDecisionBlockScope, nil
	case EventDecisionBlockTeamID:
		return EventDecisionBlockTeamID, nil
	case EventDecisionBlockSigningID:
		return EventDecisionBlockSigningID, nil
	case EventDecisionBlockCDHash:
		return EventDecisionBlockCDHash, nil
	case EventDecisionBundleBinary:
		return EventDecisionBundleBinary, nil
	default:
		return "", fmt.Errorf("unsupported event decision %q", value)
	}
}

func ParseFileAccessDecision(value string) (FileAccessDecision, error) {
	switch FileAccessDecision(value) {
	case FileAccessDecisionUnknown:
		return FileAccessDecisionUnknown, nil
	case FileAccessDecisionDenied:
		return FileAccessDecisionDenied, nil
	case FileAccessDecisionDeniedInvalidSignature:
		return FileAccessDecisionDeniedInvalidSignature, nil
	case FileAccessDecisionAuditOnly:
		return FileAccessDecisionAuditOnly, nil
	default:
		return "", fmt.Errorf("unsupported file access decision %q", value)
	}
}
