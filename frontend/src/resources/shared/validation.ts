import type { RuleType } from "@/api/types";
import type { Validator } from "react-admin";
import { z } from "zod";

type FormValues = Record<string, unknown>;

const nonEmptyTrimmedString = z.string().trim().min(1);

const identifierSchemas: Record<RuleType, z.ZodType<string>> = {
  binary: z.string().regex(/^[0-9a-fA-F]{64}$/),
  certificate: z.string().regex(/^[0-9a-fA-F]{64}$/),
  cd_hash: z.string().regex(/^[0-9a-fA-F]{40}$/),
  signing_id: z.string().regex(/^(?:[A-Z0-9]{10}|platform):[a-zA-Z0-9.-]+$/),
  team_id: z.string().regex(/^[A-Z0-9]{10}$/),
};

export const trimmedRequired =
  (label: string): Validator =>
  (value: unknown): string | undefined => {
    const parsedValue = typeof value === "string" ? value : "";
    return nonEmptyTrimmedString.safeParse(parsedValue).success ? undefined : `${label} is required`;
  };

export const ruleIdentifierValidator: Validator = (value: unknown, allValues?: FormValues): string | undefined => {
  const ruleType = allValues?.rule_type;
  if (typeof ruleType !== "string" || !(ruleType in identifierSchemas)) {
    return undefined;
  }

  const schema = identifierSchemas[ruleType as RuleType];
  const parsedValue = typeof value === "string" ? value.trim() : "";
  if (schema.safeParse(parsedValue).success) {
    return undefined;
  }

  return "Identifier format is invalid.";
};
