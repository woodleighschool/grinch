import { z } from "zod";

// Helper function to validate identifiers based on rule type
export function getIdentifierValidation(ruleType: string) {
  switch (ruleType) {
    case "BINARY":
    case "CERTIFICATE":
      return z.string().regex(/^[a-fA-F0-9]{64}$/, "Must be a valid 64-character SHA-256 hash");
    case "SIGNINGID":
      return z.string().regex(/^[A-Z0-9]{10}:[a-zA-Z0-9.-]+$/, "Must be in format: TEAMID:bundle.identifier");
    case "TEAMID":
      return z.string().regex(/^[A-Z0-9]{10}$/, "Must be a 10-character Apple Developer Team ID");
    case "CDHASH":
      return z.string().regex(/^[a-fA-F0-9]{40}$/, "Must be a 40-character CDHash");
    default:
      return z.string().min(1, "Identifier is required");
  }
}

// Application form validation schema
export const applicationFormSchema = z
  .object({
    name: z.string().min(1, "Application name is required").max(100, "Name too long"),
    rule_type: z.enum(["BINARY", "CERTIFICATE", "SIGNINGID", "TEAMID", "CDHASH"]),
    identifier: z.string().min(1, "Identifier is required"),
    description: z.string().optional(),
  })
  .refine(
    (data) => {
      const identifierValidation = getIdentifierValidation(data.rule_type);
      const result = identifierValidation.safeParse(data.identifier);
      return result.success;
    },
    {
      message: "Invalid identifier format for the selected rule type",
      path: ["identifier"],
    },
  );

// User search form validation
export const userSearchSchema = z.object({
  searchTerm: z.string().max(100, "Search term too long"),
});

// Device search form validation
export const deviceSearchSchema = z.object({
  searchTerm: z.string().max(100, "Search term too long"),
});

// Rule assignment validation
export const ruleAssignmentSchema = z.object({
  targetType: z.enum(["user", "group"], {
    message: "Please select either user or group",
  }),
  targetId: z.string().min(1, "Please select a target"),
  action: z.enum(["allow", "block"], {
    message: "Please select an action",
  }),
});

// Export types for use in components
export type ApplicationFormData = z.infer<typeof applicationFormSchema>;
export type UserSearchData = z.infer<typeof userSearchSchema>;
export type DeviceSearchData = z.infer<typeof deviceSearchSchema>;
export type RuleAssignmentData = z.infer<typeof ruleAssignmentSchema>;
