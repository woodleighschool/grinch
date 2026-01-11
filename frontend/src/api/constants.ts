export type EnumEntry<ID extends number> = Readonly<{
  id: ID;
  name: string;
  description: string;
}>;

type EnumId<Entries extends readonly EnumEntry<number>[]> = Entries[number]["id"];

type EnumSpec<
  Values extends Record<string, number>,
  Entries extends readonly EnumEntry<Values[keyof Values]>[],
> = Readonly<
  Values & {
    entries: Entries;
    nameById: Record<EnumId<Entries>, string>;
    descriptionById: Record<EnumId<Entries>, string>;
    choices: (...keys: readonly (keyof Values)[]) => { id: EnumId<Entries>; name: string }[];
  }
>;

export type EnumValue<T> = {
  [K in keyof T]: T[K] extends number ? T[K] : never;
}[keyof T];

interface EnumChoice<Id extends number> {
  id: Id;
  name: string;
}

const toChoice = <Id extends number>(entry: EnumChoice<Id>): EnumChoice<Id> => ({
  id: entry.id,
  name: entry.name,
});

export const enumName = (
  enumSpec: { nameById: Record<number, string> },
  id: number | undefined,
): string | undefined => {
  if (id == undefined) {
    return undefined;
  }
  return enumSpec.nameById[id];
};

export const enumDescription = (
  enumSpec: { descriptionById: Record<number, string> },
  id: number | undefined,
): string | undefined => {
  if (id == undefined) {
    return undefined;
  }
  return enumSpec.descriptionById[id];
};

export const defineEnum = <
  const Values extends Record<string, number>,
  const Entries extends readonly EnumEntry<Values[keyof Values]>[],
>(
  values: Values,
  entries: Entries,
): EnumSpec<Values, Entries> => {
  type Id = Entries[number]["id"];

  const nameById = {} as Record<Id, string>;
  const descriptionById = {} as Record<Id, string>;

  for (const entry of entries) {
    nameById[entry.id as Id] = entry.name;
    descriptionById[entry.id as Id] = entry.description;
  }

  const choices = (...keys: readonly (keyof Values)[]): EnumChoice<Id>[] =>
    keys.map((key): EnumChoice<Id> => {
      const id = values[key] as unknown as Id;
      return { id, name: nameById[id] };
    });

  return { ...values, entries, nameById, descriptionById, choices };
};

// CLIENT_MODE

export const CLIENT_MODE = defineEnum(
  {
    UNSPECIFIED: 0,
    MONITOR: 1,
    LOCKDOWN: 2,
    STANDALONE: 3,
  } as const,
  [
    { id: 0, name: "Unspecified/Unknown", description: "No specific mode is indicated." },
    { id: 1, name: "Monitor", description: "Apps that don't match a rule are still allowed to run." },
    { id: 2, name: "Lockdown", description: "Apps that don't match a rule are blocked from running." },
    {
      id: 3,
      name: "Standalone",
      description: "Apps that don't match a rule are blocked, but the user can approve the run by authenticating.",
    },
  ] as const,
);
export type ClientMode = EnumValue<typeof CLIENT_MODE>;
export const CLIENT_MODE_CHOICES = CLIENT_MODE.entries.map((entry): EnumChoice<typeof entry.id> => toChoice(entry));

// SYNC_TYPE

export const SYNC_TYPE = defineEnum(
  {
    UNSPECIFIED: 0,
    NORMAL_PROGRESSIVE: 1,
    CLEAN: 2,
    CLEAN_ALL: 3,
    CLEAN_STANDALONE: 4,
    CLEAN_RULES_ONLY: 5,
    CLEAN_FILE_ACCESS_RULES_ONLY: 6,
  } as const,
  [
    { id: 0, name: "Unspecified", description: 'Treated as the default "Normal" behavior.' },
    {
      id: 1,
      name: "Normal (Progressive)",
      description:
        'New rules are applied on top of existing rules; matching rules are replaced; rules marked "Remove" delete the matching rule.',
    },
    {
      id: 2,
      name: "Clean",
      description:
        "Deletes previously received (non-transitive) execution rules and all file access rules before applying new ones.",
    },
    {
      id: 3,
      name: "Clean All",
      description: "Deletes all previously received rules and all file access rules before applying new ones.",
    },
    {
      id: 4,
      name: "Clean Standalone",
      description: "Deletes rules created while in Standalone mode before applying new ones.",
    },
    {
      id: 5,
      name: "Clean Rules Only",
      description:
        "Deletes previously received (non-transitive) execution rules before applying new ones; file access rules are left alone.",
    },
    {
      id: 6,
      name: "Clean File Access Rules Only",
      description: "Deletes all existing file access rules before applying new ones; execution rules are left alone.",
    },
  ] as const,
);
export type SyncType = EnumValue<typeof SYNC_TYPE>;

// FILE_ACCESS_ACTION

export const FILE_ACCESS_ACTION = defineEnum(
  {
    UNSPECIFIED: 0,
    NO_OVERRIDE: 1,
    AUDIT_ONLY: 2,
    DISABLED: 3,
  } as const,
  [
    { id: 0, name: "Unspecified", description: "Does not change any file access settings on the computer." },
    {
      id: 1,
      name: "No Override (Apply Policy as Written)",
      description: "Uses the file access policy exactly as configured.",
    },
    { id: 2, name: "Audit Only", description: "Actions that would be denied are logged, but still allowed." },
    { id: 3, name: "Disabled", description: "No file access action is taken." },
  ] as const,
);
export type FileAccessAction = EnumValue<typeof FILE_ACCESS_ACTION>;
export const FILE_ACCESS_ACTION_CHOICES = FILE_ACCESS_ACTION.entries.map(
  (entry): EnumChoice<typeof entry.id> => toChoice(entry),
);

// DECISION

export const DECISION = defineEnum(
  {
    UNKNOWN_UNSPECIFIED: 0,
    ALLOWED_MONITOR: 1,
    ALLOWED_BINARY: 2,
    ALLOWED_CERTIFICATE: 3,
    ALLOWED_SCOPE_PATH_SCRIPT: 4,
    ALLOWED_TEAM_ID: 5,
    ALLOWED_SIGNING_ID: 6,
    ALLOWED_CDHASH: 7,
    BLOCKED_LOCKDOWN: 8,
    BLOCKED_BINARY: 9,
    BLOCKED_CERTIFICATE: 10,
    BLOCKED_SCOPE_PATH: 11,
    BLOCKED_TEAM_ID: 12,
    BLOCKED_SIGNING_ID: 13,
    BLOCKED_CDHASH: 14,
    BUNDLE_INVENTORY_ITEM: 15,
  } as const,
  [
    { id: 0, name: "Unknown/Unspecified", description: "No decision is indicated." },
    {
      id: 1,
      name: "Allowed (Monitor Fallback)",
      description: "Allowed because nothing matched while in Monitor mode.",
    },
    { id: 2, name: "Allowed (Binary)", description: "Allowed by a rule for this exact binary." },
    { id: 3, name: "Allowed (Certificate)", description: "Allowed by a matching signing certificate." },
    {
      id: 4,
      name: "Allowed (Scope/Path/Script Allow)",
      description: "Allowed by an approved path/scope or script exception.",
    },
    { id: 5, name: "Allowed (Team ID)", description: "Allowed by a matching publisher team ID." },
    { id: 6, name: "Allowed (Signing ID)", description: "Allowed by a matching signing identifier." },
    { id: 7, name: "Allowed (CDHash)", description: "Allowed by a matching code directory hash." },
    {
      id: 8,
      name: "Blocked (Lockdown Fallback)",
      description: "Blocked because nothing matched while in Lockdown mode.",
    },
    { id: 9, name: "Blocked (Binary)", description: "Blocked by a rule for this exact binary." },
    { id: 10, name: "Blocked (Certificate)", description: "Blocked by a matching certificate rule." },
    {
      id: 11,
      name: "Blocked (Scope/Path Protection)",
      description: "Blocked by a protected path/scope requirement.",
    },
    { id: 12, name: "Blocked (Team ID)", description: "Blocked by a publisher team ID rule." },
    { id: 13, name: "Blocked (Signing ID)", description: "Blocked by a signing identifier rule." },
    { id: 14, name: "Blocked (CDHash)", description: "Blocked by a code directory hash rule." },
    {
      id: 15,
      name: "Bundle Inventory Item (Not an Execution)",
      description: "Metadata entry for bundle contents, not an execution decision.",
    },
  ] as const,
);
export type Decision = EnumValue<typeof DECISION>;
export const DECISION_CHOICES = DECISION.entries.map((entry): EnumChoice<typeof entry.id> => toChoice(entry));

// SIGNING_STATUS

export const SIGNING_STATUS = defineEnum(
  {
    UNSPECIFIED: 0,
    UNSIGNED: 1,
    INVALID_SIGNATURE: 2,
    AD_HOC_SIGNED: 3,
    SIGNED_DEVELOPMENT: 4,
    SIGNED_PRODUCTION: 5,
  } as const,
  [
    { id: 0, name: "Unspecified", description: "No signing status is indicated." },
    { id: 1, name: "Unsigned", description: "The app/binary wasn't signed." },
    { id: 2, name: "Invalid Signature", description: "Signature validation failed or the signature was invalid." },
    { id: 3, name: "Ad-Hoc Signed", description: "The app/binary was ad-hoc signed." },
    { id: 4, name: "Signed (Development)", description: "Valid signature using a development certificate." },
    { id: 5, name: "Signed (Production)", description: "Valid signature using a production certificate." },
  ] as const,
);
export type SigningStatus = EnumValue<typeof SIGNING_STATUS>;
export const SIGNING_STATUS_CHOICES = SIGNING_STATUS.entries.map(
  (entry): EnumChoice<typeof entry.id> => toChoice(entry),
);

// FILE_ACCESS_DECISION

export const FILE_ACCESS_DECISION = defineEnum(
  {
    UNKNOWN_UNSPECIFIED: 0,
    DENIED_POLICY: 1,
    DENIED_INVALID_SIGNATURE: 2,
    ALLOWED_AUDITED: 3,
  } as const,
  [
    { id: 0, name: "Unknown/Unspecified", description: "No decision is indicated." },
    { id: 1, name: "Denied (Policy)", description: "The access was blocked by policy." },
    {
      id: 2,
      name: "Denied (Invalid Signature)",
      description: "The access was blocked because the process signature was invalid.",
    },
    { id: 3, name: "Allowed but Audited", description: "The access was allowed, but recorded for auditing." },
  ] as const,
);
export type FileAccessDecision = EnumValue<typeof FILE_ACCESS_DECISION>;

// POLICY_STATUS

export const POLICY_STATUS = defineEnum(
  {
    UNASSIGNED: 0,
    PENDING: 1,
    UP_TO_DATE: 2,
  } as const,
  [
    { id: 0, name: "Unassigned", description: "No policy is assigned." },
    { id: 1, name: "Pending", description: "The policy is awaiting a client sync." },
    { id: 2, name: "Up to date", description: "The client has applied the assigned policy." },
  ] as const,
);
export type PolicyStatus = EnumValue<typeof POLICY_STATUS>;
export const POLICY_STATUS_CHOICES = POLICY_STATUS.entries.map((entry): EnumChoice<typeof entry.id> => toChoice(entry));

// POLICY

export const POLICY = defineEnum(
  {
    UNKNOWN_IGNORE: 0,
    ALLOW: 1,
    ALLOW_COMPILER: 2,
    BLOCK: 3,
    BLOCK_SILENTLY: 4,
    REMOVE_EXISTING_RULE: 5,
    EVALUATE_EXPRESSION: 6,
  } as const,
  [
    { id: 0, name: "Unknown/Ignore", description: "Should not be set; ignored." },
    { id: 1, name: "Allow", description: "This target should be allowed." },
    {
      id: 2,
      name: "Allow (Compiler)",
      description:
        "Allowed; and if transitive allowlisting is enabled on the computer, files created by this process become locally allowed too.",
    },
    { id: 3, name: "Block", description: "This target should be blocked." },
    {
      id: 4,
      name: "Block (Silently)",
      description:
        "Block it and do not show GUI notifications (intended to be used sparingly because silent blocks are confusing).",
    },
    {
      id: 5,
      name: "Remove Existing Rule",
      description:
        "Removes the matching existing rule so the computer falls back to lower-precedence rules or the current ClientMode behavior.",
    },
    {
      id: 6,
      name: "Evaluate Expression",
      description: "The rule's outcome is decided by evaluating the attached CEL expression.",
    },
  ] as const,
);
export type Policy = EnumValue<typeof POLICY>;

// RULE_TYPE

export const RULE_TYPE = defineEnum(
  {
    UNKNOWN_IGNORE: 0,
    BINARY: 1,
    CERTIFICATE: 2,
    TEAM_ID: 3,
    SIGNING_ID: 4,
    CDHASH: 5,
  } as const,
  [
    { id: 0, name: "Unknown/Ignore", description: "Should not be set; ignored." },
    { id: 1, name: "Binary", description: "Matches a file by its SHA-256 hash." },
    { id: 2, name: "Certificate", description: "Matches the leaf signing certificate SHA-256." },
    {
      id: 3,
      name: "Team ID",
      description: "Matches the publisher's 10-character Apple Team ID.",
    },
    {
      id: 4,
      name: "Signing ID",
      description: "Matches the signing identifier combined with the team/platform prefix.",
    },
    { id: 5, name: "CDHash", description: "Matches the code directory hash." },
  ] as const,
);
export type RuleType = EnumValue<typeof RULE_TYPE>;
export const RULE_TYPE_CHOICES = RULE_TYPE.entries.map((entry): EnumChoice<typeof entry.id> => toChoice(entry));
