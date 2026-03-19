import type { components } from "@/api/openapi";

export type Source = components["schemas"]["Source"];
export type MemberKind = components["schemas"]["MemberKind"];
export type GroupMembershipKind = components["schemas"]["GroupMembershipKind"];
export type EventDecision = components["schemas"]["EventDecision"];
export type FileAccessDecision = components["schemas"]["FileAccessDecision"];
export type RuleType = components["schemas"]["RuleType"];
export type RulePolicy = components["schemas"]["RulePolicy"];
export type RuleTargetSubjectKind = components["schemas"]["RuleTargetSubjectKind"];
export type ExecutableSource = components["schemas"]["ExecutableSource"];

export type User = components["schemas"]["User"];
export type UserListResponse = components["schemas"]["UserListResponse"];

export type Group = components["schemas"]["Group"];
export type GroupListResponse = components["schemas"]["GroupListResponse"];
export type GroupCreateRequest = components["schemas"]["GroupCreateRequest"];

export type GroupMembership = components["schemas"]["GroupMembership"];
export type GroupMembershipListItem = components["schemas"]["GroupMembershipListItem"];
export type GroupMembershipListResponse = components["schemas"]["GroupMembershipListResponse"];
export type GroupMembershipCreateRequest = components["schemas"]["GroupMembershipCreateRequest"];

export type Machine = components["schemas"]["Machine"];
export type MachineSummary = components["schemas"]["MachineSummary"];
export type MachineListResponse = components["schemas"]["MachineListResponse"];
export type MachineRule = components["schemas"]["MachineRule"];
export type MachineRuleListResponse = components["schemas"]["MachineRuleListResponse"];
export type MachineRuleSyncStatus = components["schemas"]["MachineRuleSyncStatus"];
export type MachineClientMode = components["schemas"]["MachineClientMode"];

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
export type RuleMachine = components["schemas"]["RuleMachine"];
export type RuleMachineListResponse = components["schemas"]["RuleMachineListResponse"];
export type RuleCreateRequest = components["schemas"]["RuleCreateRequest"];
export type RuleUpdateRequest = components["schemas"]["RuleUpdateRequest"];
export type RuleTargets = components["schemas"]["RuleTargets"];
export type IncludeRuleTarget = components["schemas"]["IncludeRuleTarget"];
export type ExcludedGroup = components["schemas"]["ExcludedGroup"];
