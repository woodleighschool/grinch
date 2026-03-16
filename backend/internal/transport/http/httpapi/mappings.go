package httpapi

import (
	"fmt"

	"github.com/woodleighschool/grinch/internal/domain"
)

func mapMachine(machine domain.Machine) Machine {
	return Machine{
		Id:               machine.ID,
		SerialNumber:     machine.SerialNumber,
		Hostname:         machine.Hostname,
		ModelIdentifier:  machine.ModelIdentifier,
		OsVersion:        machine.OSVersion,
		OsBuild:          machine.OSBuild,
		SantaVersion:     machine.SantaVersion,
		PrimaryUser:      machine.PrimaryUser,
		PrimaryUserId:    machine.PrimaryUserID,
		RequestCleanSync: machine.RequestCleanSync,
		LastSeenAt:       machine.LastSeenAt,
		CreatedAt:        machine.CreatedAt,
		UpdatedAt:        machine.UpdatedAt,
	}
}

func mapMachineSummary(machine domain.MachineSummary) MachineSummary {
	return MachineSummary{
		Id:              machine.ID,
		SerialNumber:    machine.SerialNumber,
		Hostname:        machine.Hostname,
		ModelIdentifier: machine.ModelIdentifier,
		OsVersion:       machine.OSVersion,
		SantaVersion:    machine.SantaVersion,
		PrimaryUser:     machine.PrimaryUser,
		PrimaryUserId:   machine.PrimaryUserID,
		LastSeenAt:      machine.LastSeenAt,
		CreatedAt:       machine.CreatedAt,
		UpdatedAt:       machine.UpdatedAt,
	}
}

func mapGroupMembershipKind(kind domain.GroupMembershipKind) (GroupMembershipKind, error) {
	switch kind {
	case domain.GroupMembershipKindActual:
		return Actual, nil
	case domain.GroupMembershipKindEffective:
		return Effective, nil
	default:
		return "", fmt.Errorf("unsupported group membership kind %q", kind)
	}
}

func mapEventDecision(decision domain.EventDecision) EventDecision {
	switch decision {
	case domain.EventDecisionUnknown:
		return EventDecisionUnknown
	case domain.EventDecisionAllowUnknown:
		return EventDecisionAllowUnknown
	case domain.EventDecisionAllowBinary:
		return EventDecisionAllowBinary
	case domain.EventDecisionAllowCertificate:
		return EventDecisionAllowCertificate
	case domain.EventDecisionAllowScope:
		return EventDecisionAllowScope
	case domain.EventDecisionAllowTeamID:
		return EventDecisionAllowTeamId
	case domain.EventDecisionAllowSigningID:
		return EventDecisionAllowSigningId
	case domain.EventDecisionAllowCDHash:
		return EventDecisionAllowCdHash
	case domain.EventDecisionBlockUnknown:
		return EventDecisionBlockUnknown
	case domain.EventDecisionBlockBinary:
		return EventDecisionBlockBinary
	case domain.EventDecisionBlockCertificate:
		return EventDecisionBlockCertificate
	case domain.EventDecisionBlockScope:
		return EventDecisionBlockScope
	case domain.EventDecisionBlockTeamID:
		return EventDecisionBlockTeamId
	case domain.EventDecisionBlockSigningID:
		return EventDecisionBlockSigningId
	case domain.EventDecisionBlockCDHash:
		return EventDecisionBlockCdHash
	case domain.EventDecisionBundleBinary:
		return EventDecisionBundleBinary
	default:
		return EventDecisionUnknown
	}
}

func mapFileAccessDecision(decision domain.FileAccessDecision) FileAccessDecision {
	switch decision {
	case domain.FileAccessDecisionUnknown:
		return FileAccessDecisionUnknown
	case domain.FileAccessDecisionDenied:
		return FileAccessDecisionDenied
	case domain.FileAccessDecisionDeniedInvalidSignature:
		return FileAccessDecisionDeniedInvalidSignature
	case domain.FileAccessDecisionAuditOnly:
		return FileAccessDecisionAuditOnly
	default:
		return FileAccessDecisionUnknown
	}
}

func mapSigningChainEntry(entry domain.SigningChainEntry) SigningChainEntry {
	return SigningChainEntry{
		CommonName:         entry.CommonName,
		Organization:       entry.Organization,
		OrganizationalUnit: entry.OrganizationalUnit,
		Sha256:             entry.SHA256,
		ValidFrom:          entry.ValidFrom,
		ValidUntil:         entry.ValidUntil,
	}
}

func mapEntitlements(entitlements map[string]domain.Entitlement) map[string]any {
	result := make(map[string]any, len(entitlements))
	for key, entitlement := range entitlements {
		result[key] = entitlement.Value
	}
	return result
}

func mapExecutable(executable domain.Executable) Executable {
	signingChain := make([]SigningChainEntry, 0, len(executable.SigningChain))
	for _, entry := range executable.SigningChain {
		signingChain = append(signingChain, mapSigningChainEntry(entry))
	}

	return Executable{
		Id:             executable.ID,
		Source:         ExecutableSource(executable.Source),
		FileSha256:     executable.FileSHA256,
		FileName:       executable.FileName,
		FilePath:       executable.FilePath,
		FileBundleId:   executable.FileBundleID,
		FileBundlePath: executable.FileBundlePath,
		SigningId:      executable.SigningID,
		TeamId:         executable.TeamID,
		Cdhash:         executable.CDHash,
		Entitlements:   mapEntitlements(executable.Entitlements),
		SigningChain:   signingChain,
		CreatedAt:      executable.CreatedAt,
	}
}

func mapExecutableSummary(executable domain.ExecutableSummary) ExecutableSummary {
	return ExecutableSummary{
		Id:             executable.ID,
		Source:         ExecutableSource(executable.Source),
		FileSha256:     executable.FileSHA256,
		FileName:       executable.FileName,
		FilePath:       executable.FilePath,
		FileBundleId:   executable.FileBundleID,
		FileBundlePath: executable.FileBundlePath,
		SigningId:      executable.SigningID,
		TeamId:         executable.TeamID,
		Cdhash:         executable.CDHash,
		CreatedAt:      executable.CreatedAt,
	}
}

func mapExecutionEvent(event domain.ExecutionEvent) ExecutionEvent {
	return ExecutionEvent{
		Id:              event.ID,
		MachineId:       event.MachineID,
		ExecutableId:    event.ExecutableID,
		Decision:        mapEventDecision(event.Decision),
		FilePath:        event.FilePath,
		FileName:        event.FileName,
		FileSha256:      event.FileSHA256,
		FileBundleId:    event.FileBundleID,
		FileBundlePath:  event.FileBundlePath,
		SigningId:       event.SigningID,
		TeamId:          event.TeamID,
		Cdhash:          event.CDHash,
		ExecutingUser:   event.ExecutingUser,
		LoggedInUsers:   event.LoggedInUsers,
		CurrentSessions: event.CurrentSessions,
		SigningChain:    mapSigningChain(event.SigningChain),
		Entitlements:    mapEntitlements(event.Entitlements),
		OccurredAt:      event.OccurredAt,
		CreatedAt:       event.CreatedAt,
	}
}

func mapExecutionEventSummary(event domain.ExecutionEventSummary) ExecutionEventSummary {
	return ExecutionEventSummary{
		Id:           event.ID,
		MachineId:    event.MachineID,
		ExecutableId: event.ExecutableID,
		Decision:     mapEventDecision(event.Decision),
		FilePath:     event.FilePath,
		FileName:     event.FileName,
		SigningId:    event.SigningID,
		OccurredAt:   event.OccurredAt,
		CreatedAt:    event.CreatedAt,
	}
}

func mapSigningChain(entries []domain.SigningChainEntry) []SigningChainEntry {
	signingChain := make([]SigningChainEntry, 0, len(entries))
	for _, entry := range entries {
		signingChain = append(signingChain, mapSigningChainEntry(entry))
	}
	return signingChain
}

func mapFileAccessEventProcess(process domain.FileAccessEventProcess) FileAccessEventProcess {
	return FileAccessEventProcess{
		Pid:          process.Pid,
		FilePath:     process.FilePath,
		ExecutableId: process.ExecutableID,
		FileName:     process.FileName,
	}
}

func mapFileAccessEvent(event domain.FileAccessEvent) FileAccessEvent {
	processChain := make([]FileAccessEventProcess, 0, len(event.ProcessChain))
	for _, process := range event.ProcessChain {
		processChain = append(processChain, mapFileAccessEventProcess(process))
	}

	return FileAccessEvent{
		Id:           event.ID,
		MachineId:    event.MachineID,
		ExecutableId: event.ExecutableID,
		RuleVersion:  event.RuleVersion,
		RuleName:     event.RuleName,
		Target:       event.Target,
		Decision:     mapFileAccessDecision(event.Decision),
		FileName:     event.FileName,
		FileSha256:   event.FileSHA256,
		SigningId:    event.SigningID,
		TeamId:       event.TeamID,
		Cdhash:       event.CDHash,
		ProcessChain: processChain,
		OccurredAt:   event.OccurredAt,
		CreatedAt:    event.CreatedAt,
	}
}

func mapFileAccessEventSummary(event domain.FileAccessEventSummary) FileAccessEventSummary {
	return FileAccessEventSummary{
		Id:           event.ID,
		MachineId:    event.MachineID,
		ExecutableId: event.ExecutableID,
		Decision:     mapFileAccessDecision(event.Decision),
		RuleName:     event.RuleName,
		Target:       event.Target,
		FileName:     event.FileName,
		FileSha256:   event.FileSHA256,
		SigningId:    event.SigningID,
		TeamId:       event.TeamID,
		Cdhash:       event.CDHash,
		OccurredAt:   event.OccurredAt,
		CreatedAt:    event.CreatedAt,
	}
}

func mapSource(source domain.PrincipalSource) (Source, error) {
	switch source {
	case domain.PrincipalSourceLocal:
		return Local, nil
	case domain.PrincipalSourceEntra:
		return Entra, nil
	default:
		return "", fmt.Errorf("unsupported source %q", source)
	}
}

func mapUser(user domain.User) (User, error) {
	source, err := mapSource(user.Source)
	if err != nil {
		return User{}, err
	}

	return User{
		Id:          user.ID,
		Upn:         user.UPN,
		DisplayName: user.DisplayName,
		Source:      source,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

func mapGroup(group domain.Group) (Group, error) {
	source, err := mapSource(group.Source)
	if err != nil {
		return Group{}, err
	}

	return Group{
		Id:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Source:      source,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}, nil
}

func mapMemberKind(kind domain.MemberKind) (MemberKind, error) {
	switch kind {
	case domain.MemberKindUser:
		return MemberKindUser, nil
	case domain.MemberKindMachine:
		return MemberKindMachine, nil
	default:
		return "", fmt.Errorf("unsupported member kind %q", kind)
	}
}

func toDomainMemberKind(kind MemberKind) (domain.MemberKind, error) {
	switch kind {
	case MemberKindUser:
		return domain.MemberKindUser, nil
	case MemberKindMachine:
		return domain.MemberKindMachine, nil
	default:
		return "", fmt.Errorf("unsupported member kind %q", kind)
	}
}

func mapGroupMembershipGroup(group domain.GroupMembershipGroup) (GroupMembershipGroup, error) {
	source, err := mapSource(group.Source)
	if err != nil {
		return GroupMembershipGroup{}, err
	}

	return GroupMembershipGroup{
		Id:     group.ID,
		Name:   group.Name,
		Source: source,
	}, nil
}

func mapGroupMembershipMember(member domain.GroupMembershipMember) (GroupMembershipMember, error) {
	memberKind, err := mapMemberKind(member.Kind)
	if err != nil {
		return GroupMembershipMember{}, err
	}

	result := GroupMembershipMember{
		Kind: memberKind,
		Id:   member.ID,
	}
	if member.Name != "" {
		result.Name = toStringPointer(member.Name)
	}
	return result, nil
}

func mapGroupMembership(membership domain.GroupMembership) (GroupMembership, error) {
	group, err := mapGroupMembershipGroup(membership.Group)
	if err != nil {
		return GroupMembership{}, err
	}
	member, err := mapGroupMembershipMember(membership.Member)
	if err != nil {
		return GroupMembership{}, err
	}
	kind, err := mapGroupMembershipKind(membership.Kind)
	if err != nil {
		return GroupMembership{}, err
	}

	return GroupMembership{
		Id:        membership.ID,
		Kind:      kind,
		Group:     group,
		Member:    member,
		CreatedAt: membership.CreatedAt,
		UpdatedAt: membership.UpdatedAt,
	}, nil
}

func mapRuleType(ruleType domain.RuleType) (RuleType, error) {
	switch ruleType {
	case domain.RuleTypeBinary:
		return Binary, nil
	case domain.RuleTypeCertificate:
		return Certificate, nil
	case domain.RuleTypeTeamID:
		return TeamId, nil
	case domain.RuleTypeSigningID:
		return SigningId, nil
	case domain.RuleTypeCDHash:
		return CdHash, nil
	default:
		return "", fmt.Errorf("unsupported rule type %q", ruleType)
	}
}

func toDomainRuleType(ruleType RuleType) (domain.RuleType, error) {
	switch ruleType {
	case Binary:
		return domain.RuleTypeBinary, nil
	case Certificate:
		return domain.RuleTypeCertificate, nil
	case TeamId:
		return domain.RuleTypeTeamID, nil
	case SigningId:
		return domain.RuleTypeSigningID, nil
	case CdHash:
		return domain.RuleTypeCDHash, nil
	default:
		return "", fmt.Errorf("unsupported rule type %q", ruleType)
	}
}

func mapRulePolicy(policy domain.RulePolicy) (RulePolicy, error) {
	switch policy {
	case domain.RulePolicyAllowlist:
		return Allowlist, nil
	case domain.RulePolicyBlocklist:
		return Blocklist, nil
	case domain.RulePolicySilentBlocklist:
		return SilentBlocklist, nil
	case domain.RulePolicyCEL:
		return Cel, nil
	default:
		return "", fmt.Errorf("unsupported rule policy %q", policy)
	}
}

func toDomainRulePolicy(policy RulePolicy) (domain.RulePolicy, error) {
	switch policy {
	case Allowlist:
		return domain.RulePolicyAllowlist, nil
	case Blocklist:
		return domain.RulePolicyBlocklist, nil
	case SilentBlocklist:
		return domain.RulePolicySilentBlocklist, nil
	case Cel:
		return domain.RulePolicyCEL, nil
	default:
		return "", fmt.Errorf("unsupported rule policy %q", policy)
	}
}

func mapRule(rule domain.Rule) (Rule, error) {
	ruleType, err := mapRuleType(rule.RuleType)
	if err != nil {
		return Rule{}, err
	}

	return Rule{
		Id:            rule.ID,
		Name:          rule.Name,
		Description:   rule.Description,
		RuleType:      ruleType,
		Identifier:    rule.Identifier,
		CustomMessage: rule.CustomMessage,
		CustomUrl:     rule.CustomURL,
		Enabled:       rule.Enabled,
		CreatedAt:     rule.CreatedAt,
		UpdatedAt:     rule.UpdatedAt,
	}, nil
}

func mapRuleSummary(rule domain.RuleSummary) (RuleSummary, error) {
	ruleType, err := mapRuleType(rule.RuleType)
	if err != nil {
		return RuleSummary{}, err
	}

	return RuleSummary{
		Id:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		RuleType:    ruleType,
		Identifier:  rule.Identifier,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}, nil
}

func mapRuleTargetAssignment(assignment domain.RuleTargetAssignment) (RuleTargetAssignment, error) {
	switch assignment {
	case domain.RuleTargetAssignmentInclude:
		return Include, nil
	case domain.RuleTargetAssignmentExclude:
		return Exclude, nil
	default:
		return "", fmt.Errorf("unsupported rule target assignment %q", assignment)
	}
}

func toDomainRuleTargetAssignment(assignment RuleTargetAssignment) (domain.RuleTargetAssignment, error) {
	switch assignment {
	case Include:
		return domain.RuleTargetAssignmentInclude, nil
	case Exclude:
		return domain.RuleTargetAssignmentExclude, nil
	default:
		return "", fmt.Errorf("unsupported rule target assignment %q", assignment)
	}
}

func mapRuleTargetSubjectKind(subjectKind domain.RuleTargetSubjectKind) (RuleTargetSubjectKind, error) {
	switch subjectKind {
	case domain.RuleTargetSubjectKindGroup:
		return RuleTargetSubjectKindGroup, nil
	default:
		return "", fmt.Errorf("unsupported rule target subject kind %q", subjectKind)
	}
}

func toDomainRuleTargetSubjectKind(subjectKind RuleTargetSubjectKind) (domain.RuleTargetSubjectKind, error) {
	switch subjectKind {
	case RuleTargetSubjectKindGroup:
		return domain.RuleTargetSubjectKindGroup, nil
	default:
		return "", fmt.Errorf("unsupported rule target subject kind %q", subjectKind)
	}
}

func mapRuleTarget(target domain.RuleTarget) (RuleTarget, error) {
	assignment, err := mapRuleTargetAssignment(target.Assignment)
	if err != nil {
		return RuleTarget{}, err
	}
	subjectKind, err := mapRuleTargetSubjectKind(target.SubjectKind)
	if err != nil {
		return RuleTarget{}, err
	}

	result := RuleTarget{
		Id:          target.ID,
		RuleId:      target.RuleID,
		SubjectKind: subjectKind,
		SubjectId:   target.SubjectID,
		Assignment:  assignment,
		Priority:    target.Priority,
		CreatedAt:   target.CreatedAt,
		UpdatedAt:   target.UpdatedAt,
	}
	if target.Policy != nil {
		policy, policyErr := mapRulePolicy(*target.Policy)
		if policyErr != nil {
			return RuleTarget{}, policyErr
		}
		result.Policy = &policy
	}
	if target.CELExpression != "" {
		result.CelExpression = toStringPointer(target.CELExpression)
	}

	return result, nil
}

func mapRuleTargetSummary(target domain.RuleTargetSummary) (RuleTargetSummary, error) {
	assignment, err := mapRuleTargetAssignment(target.Assignment)
	if err != nil {
		return RuleTargetSummary{}, err
	}
	subjectKind, err := mapRuleTargetSubjectKind(target.SubjectKind)
	if err != nil {
		return RuleTargetSummary{}, err
	}

	result := RuleTargetSummary{
		Id:          target.ID,
		RuleId:      target.RuleID,
		SubjectKind: subjectKind,
		SubjectId:   target.SubjectID,
		Assignment:  assignment,
		Priority:    target.Priority,
		CreatedAt:   target.CreatedAt,
		UpdatedAt:   target.UpdatedAt,
	}
	if target.Policy != nil {
		policy, policyErr := mapRulePolicy(*target.Policy)
		if policyErr != nil {
			return RuleTargetSummary{}, policyErr
		}
		result.Policy = &policy
	}

	return result, nil
}
