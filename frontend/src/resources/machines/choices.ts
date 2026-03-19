import type { components } from "@/api/openapi";

type MachineClientMode = components["schemas"]["MachineClientMode"];
type MachineRuleSyncStatus = components["schemas"]["MachineRuleSyncStatus"];

export const RULE_SYNC_STATUS_CHOICES = [
  { id: "synced", name: "Synced" },
  { id: "pending", name: "Pending" },
  { id: "issue", name: "Issue" },
] satisfies { id: MachineRuleSyncStatus; name: string }[];

export const CLIENT_MODE_CHOICES = [
  { id: "unknown", name: "Unknown" },
  { id: "monitor", name: "Monitor" },
  { id: "lockdown", name: "Lockdown" },
  { id: "standalone", name: "Standalone" },
] satisfies { id: MachineClientMode; name: string }[];
