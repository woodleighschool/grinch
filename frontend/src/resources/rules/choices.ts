import type { components } from "@/api/openapi";

type RulePolicy = components["schemas"]["RulePolicy"];
type RuleTargetSubjectKind = components["schemas"]["RuleTargetSubjectKind"];
type RuleType = components["schemas"]["RuleType"];

export const RULE_ENABLED_CHOICES = [
  { id: true, name: "Enabled" },
  { id: false, name: "Disabled" },
] satisfies { id: boolean; name: string }[];

export const RULE_TYPE_CHOICES = [
  { id: "binary", name: "Binary" },
  { id: "certificate", name: "Certificate" },
  { id: "team_id", name: "Team ID" },
  { id: "signing_id", name: "Signing ID" },
  { id: "cd_hash", name: "CD Hash" },
] satisfies { id: RuleType; name: string }[];

export const RULE_POLICY_CHOICES = [
  { id: "allowlist", name: "Allowlist" },
  { id: "blocklist", name: "Blocklist" },
  { id: "silent_blocklist", name: "Silent Blocklist" },
  { id: "cel", name: "CEL" },
] satisfies { id: RulePolicy; name: string }[];

export const RULE_TARGET_SUBJECT_KIND_CHOICES = [
  { id: "group", name: "Group" },
  { id: "all_devices", name: "All Devices" },
  { id: "all_users", name: "All Users" },
] satisfies { id: RuleTargetSubjectKind; name: string }[];
