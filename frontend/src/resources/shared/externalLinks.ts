import type { components } from "@/api/openapi";

type RulePolicy = components["schemas"]["RulePolicy"];

export const SANTA_RULE_POLICY_DOCS: Record<RulePolicy, string> = {
  allowlist: "https://northpole.dev/features/binary-authorization/#allowlist",
  blocklist: "https://northpole.dev/features/binary-authorization/#blocklist",
  silent_blocklist: "https://northpole.dev/features/binary-authorization/#silent-blocklist",
  cel: "https://northpole.dev/features/binary-authorization/#cel",
};

export const SANTA_CEL_COOKBOOK_URL = "https://northpole.dev/cookbook/cel/";
export const SANTA_CEL_PLAYGROUND_URL = "https://northpole.dev/cookbook/cel-playground/";
