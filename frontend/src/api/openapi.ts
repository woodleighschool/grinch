export * from "./openapi-generated";

import type {
  Executable,
  ExecutableListResponse,
  ExecutionDecision,
  ExecutionEvent,
  ExecutionEventListResponse,
  FileAccessDecision,
  FileAccessEvent,
  FileAccessEventListResponse,
  Group,
  GroupListResponse,
  Machine,
  MachineClientMode,
  MachineListResponse,
  MachineRuleListResponse,
  MachineRuleSyncStatus,
  MemberKind,
  Membership,
  MembershipCreateRequest,
  MembershipListResponse,
  Rule,
  RuleListResponse,
  RuleMachineListResponse,
  RulePolicy,
  RuleTargetSubjectKind,
  RuleType,
  Source,
  User,
  UserListResponse,
} from "./openapi-generated";

export interface components {
  schemas: {
    Executable: Executable;
    ExecutableListResponse: ExecutableListResponse;
    ExecutionDecision: ExecutionDecision;
    ExecutionEvent: ExecutionEvent;
    ExecutionEventListResponse: ExecutionEventListResponse;
    FileAccessDecision: FileAccessDecision;
    FileAccessEvent: FileAccessEvent;
    FileAccessEventListResponse: FileAccessEventListResponse;
    Group: Group;
    GroupListResponse: GroupListResponse;
    Machine: Machine;
    MachineClientMode: MachineClientMode;
    MachineListResponse: MachineListResponse;
    MachineRuleListResponse: MachineRuleListResponse;
    MachineRuleSyncStatus: MachineRuleSyncStatus;
    MemberKind: MemberKind;
    Membership: Membership;
    MembershipCreateRequest: MembershipCreateRequest;
    MembershipListResponse: MembershipListResponse;
    Rule: Rule;
    RuleListResponse: RuleListResponse;
    RuleMachineListResponse: RuleMachineListResponse;
    RulePolicy: RulePolicy;
    RuleTargetSubjectKind: RuleTargetSubjectKind;
    RuleType: RuleType;
    Source: Source;
    User: User;
    UserListResponse: UserListResponse;
  };
}
