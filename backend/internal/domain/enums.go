package domain

import "fmt"

type ExecutionDecision string

const (
	ExecutionDecisionUnknown          ExecutionDecision = "unknown"
	ExecutionDecisionAllowUnknown     ExecutionDecision = "allow_unknown"
	ExecutionDecisionAllowBinary      ExecutionDecision = "allow_binary"
	ExecutionDecisionAllowCertificate ExecutionDecision = "allow_certificate"
	ExecutionDecisionAllowScope       ExecutionDecision = "allow_scope"
	ExecutionDecisionAllowTeamID      ExecutionDecision = "allow_team_id"
	ExecutionDecisionAllowSigningID   ExecutionDecision = "allow_signing_id"
	ExecutionDecisionAllowCDHash      ExecutionDecision = "allow_cd_hash"
	ExecutionDecisionBlockUnknown     ExecutionDecision = "block_unknown"
	ExecutionDecisionBlockBinary      ExecutionDecision = "block_binary"
	ExecutionDecisionBlockCertificate ExecutionDecision = "block_certificate"
	ExecutionDecisionBlockScope       ExecutionDecision = "block_scope"
	ExecutionDecisionBlockTeamID      ExecutionDecision = "block_team_id"
	ExecutionDecisionBlockSigningID   ExecutionDecision = "block_signing_id"
	ExecutionDecisionBlockCDHash      ExecutionDecision = "block_cd_hash"
	ExecutionDecisionBundleBinary     ExecutionDecision = "bundle_binary"
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

func ParseExecutionDecision(value string) (ExecutionDecision, error) {
	return parseEnum(value, "event decision",
		ExecutionDecisionUnknown, ExecutionDecisionAllowUnknown, ExecutionDecisionAllowBinary,
		ExecutionDecisionAllowCertificate, ExecutionDecisionAllowScope, ExecutionDecisionAllowTeamID,
		ExecutionDecisionAllowSigningID, ExecutionDecisionAllowCDHash,
		ExecutionDecisionBlockUnknown, ExecutionDecisionBlockBinary, ExecutionDecisionBlockCertificate,
		ExecutionDecisionBlockScope, ExecutionDecisionBlockTeamID, ExecutionDecisionBlockSigningID,
		ExecutionDecisionBlockCDHash, ExecutionDecisionBundleBinary,
	)
}

func ParseFileAccessDecision(value string) (FileAccessDecision, error) {
	return parseEnum(value, "file access decision",
		FileAccessDecisionUnknown, FileAccessDecisionDenied,
		FileAccessDecisionDeniedInvalidSignature, FileAccessDecisionAuditOnly,
	)
}
