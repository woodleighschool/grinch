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

func parseEnum[T ~string](value, kind string, valid ...T) (T, error) {
	for _, v := range valid {
		if T(value) == v {
			return v, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("unsupported %s %q", kind, value)
}

func ParsePrincipalSource(value string) (PrincipalSource, error) {
	return parseEnum(value, "source", PrincipalSourceLocal, PrincipalSourceEntra)
}

func ParseMemberKind(value string) (MemberKind, error) {
	return parseEnum(value, "member kind", MemberKindUser, MemberKindMachine)
}

func ParseRuleType(value string) (RuleType, error) {
	return parseEnum(value, "rule type",
		RuleTypeBinary, RuleTypeCertificate, RuleTypeTeamID, RuleTypeSigningID, RuleTypeCDHash,
	)
}

func ParseRulePolicy(value string) (RulePolicy, error) {
	return parseEnum(value, "rule policy",
		RulePolicyAllowlist, RulePolicyBlocklist, RulePolicySilentBlocklist, RulePolicyCEL,
	)
}

func ParseRuleTargetAssignment(value string) (RuleTargetAssignment, error) {
	return parseEnum(value, "rule target assignment",
		RuleTargetAssignmentInclude, RuleTargetAssignmentExclude,
	)
}

func ParseRuleTargetSubjectKind(value string) (RuleTargetSubjectKind, error) {
	return parseEnum(value, "rule target subject kind",
		RuleTargetSubjectKindGroup, RuleTargetSubjectKindAllDevices, RuleTargetSubjectKindAllUsers,
	)
}

func ParseMachineRuleSyncStatus(value string) (MachineRuleSyncStatus, error) {
	return parseEnum(value, "machine rule sync status",
		MachineRuleSyncStatusSynced, MachineRuleSyncStatusPending, MachineRuleSyncStatusIssue,
	)
}

func ParseMachineClientMode(value string) (MachineClientMode, error) {
	return parseEnum(value, "machine client mode",
		MachineClientModeUnknown, MachineClientModeMonitor, MachineClientModeLockdown, MachineClientModeStandalone,
	)
}

func ParseEventDecision(value string) (EventDecision, error) {
	return parseEnum(value, "event decision",
		EventDecisionUnknown, EventDecisionAllowUnknown, EventDecisionAllowBinary,
		EventDecisionAllowCertificate, EventDecisionAllowScope, EventDecisionAllowTeamID,
		EventDecisionAllowSigningID, EventDecisionAllowCDHash,
		EventDecisionBlockUnknown, EventDecisionBlockBinary, EventDecisionBlockCertificate,
		EventDecisionBlockScope, EventDecisionBlockTeamID, EventDecisionBlockSigningID,
		EventDecisionBlockCDHash, EventDecisionBundleBinary,
	)
}

func ParseFileAccessDecision(value string) (FileAccessDecision, error) {
	return parseEnum(value, "file access decision",
		FileAccessDecisionUnknown, FileAccessDecisionDenied,
		FileAccessDecisionDeniedInvalidSignature, FileAccessDecisionAuditOnly,
	)
}
