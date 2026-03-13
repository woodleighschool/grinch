import type { FileAccessDecision } from "@/api/types";

export const FILE_ACCESS_DECISION_CHOICES = [
  { id: "unknown", name: "Unknown" },
  { id: "denied", name: "Denied" },
  { id: "denied_invalid_signature", name: "Denied (Invalid Signature)" },
  { id: "audit_only", name: "Audit Only" },
] satisfies { id: FileAccessDecision; name: string }[];

export const FILE_ACCESS_DECISION_DESCRIPTIONS: Record<FileAccessDecision, string> = {
  unknown: "Santa reported a file access event but did not attach a clearer decision state.",
  denied: "Santa denied the file access because the rule matched and access was blocked.",
  denied_invalid_signature: "Santa denied the file access because the accessing process had an invalid signature.",
  audit_only: "Santa recorded the file access event without blocking it.",
};

export const fileAccessDecisionName = (decision: FileAccessDecision): string =>
  FILE_ACCESS_DECISION_CHOICES.find((c): boolean => c.id === decision)?.name ?? decision;
