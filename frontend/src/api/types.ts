import type { components } from "@/api/openapi";

export type Source = components["schemas"]["Source"];
export type MemberKind = components["schemas"]["MemberKind"];
export type GroupMembershipKind = components["schemas"]["GroupMembershipKind"];
export type EventDecision = components["schemas"]["EventDecision"];
export type FileAccessDecision = components["schemas"]["FileAccessDecision"];
export type RuleType = components["schemas"]["RuleType"];
export type RulePolicy = components["schemas"]["RulePolicy"];
export type RuleTargetAssignment = components["schemas"]["RuleTargetAssignment"];
export type RuleTargetSubjectKind = components["schemas"]["RuleTargetSubjectKind"];

export type User = components["schemas"]["User"];
export type UserListResponse = components["schemas"]["UserListResponse"];

export type Group = components["schemas"]["Group"];
export type GroupListResponse = components["schemas"]["GroupListResponse"];
export type GroupCreateRequest = components["schemas"]["GroupCreateRequest"];
export type GroupPatchRequest = components["schemas"]["GroupPatchRequest"];

export type GroupMembership = components["schemas"]["GroupMembership"];
export type GroupMembershipListResponse = components["schemas"]["GroupMembershipListResponse"];
export type GroupMembershipCreateRequest = components["schemas"]["GroupMembershipCreateRequest"];

export type Machine = components["schemas"]["Machine"];
export type MachineSummary = components["schemas"]["MachineSummary"];
export type MachineListResponse = components["schemas"]["MachineListResponse"];

export type Executable = components["schemas"]["Executable"];
export type ExecutableSummary = components["schemas"]["ExecutableSummary"];
export type ExecutableListResponse = components["schemas"]["ExecutableListResponse"];

export type ExecutionEvent = components["schemas"]["ExecutionEvent"];
export type ExecutionEventSummary = components["schemas"]["ExecutionEventSummary"];
export type ExecutionEventListResponse = components["schemas"]["ExecutionEventListResponse"];

export type FileAccessEvent = components["schemas"]["FileAccessEvent"];
export type FileAccessEventSummary = components["schemas"]["FileAccessEventSummary"];
export type FileAccessEventListResponse = components["schemas"]["FileAccessEventListResponse"];

export type Rule = components["schemas"]["Rule"];
export type RuleSummary = components["schemas"]["RuleSummary"];
export type RuleListResponse = components["schemas"]["RuleListResponse"];
export type RuleCreateRequest = components["schemas"]["RuleCreateRequest"];
export type RulePatchRequest = components["schemas"]["RulePatchRequest"];

export type RuleTarget = components["schemas"]["RuleTarget"];
export type RuleTargetSummary = components["schemas"]["RuleTargetSummary"];
export type RuleTargetListResponse = components["schemas"]["RuleTargetListResponse"];
export type RuleTargetCreateRequest = components["schemas"]["RuleTargetCreateRequest"];
export type RuleTargetPatchRequest = components["schemas"]["RuleTargetPatchRequest"];
