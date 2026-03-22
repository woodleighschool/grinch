package domain

import "github.com/google/uuid"

type ListOptions struct {
	IDs    []uuid.UUID
	Limit  int32
	Offset int32
	Search string
	Sort   string
	Order  string
}

type UserListOptions struct {
	ListOptions
}

type GroupListOptions struct {
	ListOptions
}

type MembershipListOptions struct {
	ListOptions

	GroupID   *uuid.UUID
	UserID    *uuid.UUID
	MachineID *uuid.UUID
}

type MachineListOptions struct {
	ListOptions

	UserID           *uuid.UUID
	RuleSyncStatuses []MachineRuleSyncStatus
	ClientModes      []MachineClientMode
}

type MachineRuleListOptions struct {
	ListOptions

	MachineID *uuid.UUID
}

type RuleMachineListOptions struct {
	ListOptions

	RuleID *uuid.UUID
}

type ExecutableListOptions struct {
	ListOptions
}

type ExecutionEventListOptions struct {
	ListOptions

	MachineID    *uuid.UUID
	UserID       *uuid.UUID
	ExecutableID *uuid.UUID
	Decisions    []EventDecision
}

type FileAccessEventListOptions struct {
	ListOptions

	MachineID *uuid.UUID
	Decisions []FileAccessDecision
}

type RuleListOptions struct {
	ListOptions

	Enabled   []bool
	RuleTypes []RuleType
}

type RuleTargetListOptions struct {
	ListOptions

	RuleID      *uuid.UUID
	SubjectKind *RuleTargetSubjectKind
	SubjectID   *uuid.UUID
	Assignment  *RuleTargetAssignment
	Policy      *RulePolicy
}
