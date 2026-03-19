package domain

import "github.com/google/uuid"

type ListOptions struct {
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

type GroupMembershipListOptions struct {
	ListOptions

	GroupID   *uuid.UUID
	UserID    *uuid.UUID
	MachineID *uuid.UUID
}

type MachineListOptions struct {
	ListOptions

	UserID *uuid.UUID
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
}

type FileAccessEventListOptions struct {
	ListOptions

	MachineID    *uuid.UUID
	ExecutableID *uuid.UUID
}

type RuleListOptions struct {
	ListOptions
}

type RuleTargetListOptions struct {
	ListOptions

	RuleID      *uuid.UUID
	SubjectKind *RuleTargetSubjectKind
	SubjectID   *uuid.UUID
	Assignment  *RuleTargetAssignment
	Policy      *RulePolicy
}
