import type { EventDecision } from "@/api/types";

export const EVENT_DECISION_CHOICES = [
  { id: "unknown", name: "Unknown" },
  { id: "allow_unknown", name: "Allow Unknown" },
  { id: "allow_binary", name: "Allow Binary" },
  { id: "allow_certificate", name: "Allow Certificate" },
  { id: "allow_scope", name: "Allow Scope" },
  { id: "allow_team_id", name: "Allow Team ID" },
  { id: "allow_signing_id", name: "Allow Signing ID" },
  { id: "allow_cd_hash", name: "Allow CD Hash" },
  { id: "block_unknown", name: "Block Unknown" },
  { id: "block_binary", name: "Block Binary" },
  { id: "block_certificate", name: "Block Certificate" },
  { id: "block_scope", name: "Block Scope" },
  { id: "block_team_id", name: "Block Team ID" },
  { id: "block_signing_id", name: "Block Signing ID" },
  { id: "block_cd_hash", name: "Block CD Hash" },
  { id: "bundle_binary", name: "Bundle Binary" },
] satisfies { id: EventDecision; name: string }[];

export const EVENT_DECISION_DESCRIPTIONS: Record<EventDecision, string> = {
  unknown: "No decision reported.",
  allow_unknown: "Allowed because no rule matched while in Monitor mode.",
  allow_binary: "Allowed by a rule for this exact binary.",
  allow_certificate: "Allowed by a matching signing certificate.",
  allow_scope: "Allowed by an approved path or script exception.",
  allow_team_id: "Allowed by a matching Team ID rule.",
  allow_signing_id: "Allowed by a matching Signing ID rule.",
  allow_cd_hash: "Allowed by a matching CDHash rule.",
  block_unknown: "Blocked because no rule matched while in Lockdown mode.",
  block_binary: "Blocked by a rule for this exact binary.",
  block_certificate: "Blocked by a matching certificate rule.",
  block_scope: "Blocked by a blocked path rule or Page Zero protection.",
  block_team_id: "Blocked by a matching Team ID rule.",
  block_signing_id: "Blocked by a matching Signing ID rule.",
  block_cd_hash: "Blocked by a matching CDHash rule.",
  bundle_binary: "Bundle contents metadata; not an execution decision.",
};

export const eventDecisionName = (decision: EventDecision): string =>
  EVENT_DECISION_CHOICES.find((c): boolean => c.id === decision)?.name ?? decision;
