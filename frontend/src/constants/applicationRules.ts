export const applicationRuleTypes = ["BINARY", "CERTIFICATE", "SIGNINGID", "TEAMID", "CDHASH"] as const;
export type ApplicationRuleType = (typeof applicationRuleTypes)[number];

export type ApplicationRuleTypeMetadata = {
  label: string;
  placeholder: string;
  example: string;
  description: string;
  referenceGroup?: "signingChain";
};

export const applicationRuleTypeMetadata: Record<ApplicationRuleType, ApplicationRuleTypeMetadata> = {
  BINARY: {
    label: "SHA-256",
    placeholder: "f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef",
    example: "f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef",
    description: "Matches one exact copy of a program using its full file hash, so nothing else counts as that app.",
  },
  CERTIFICATE: {
    label: "SHA-256",
    placeholder: "1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64",
    example: "1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64",
    referenceGroup: "signingChain",
    description: "Covers every app signed with this certificate, so trusting or blocking it affects all software that uses that signer.",
  },
  SIGNINGID: {
    label: "Signing ID",
    placeholder: "ZMCG7MLDV9:com.northpolesec.santa",
    example: "ZMCG7MLDV9:com.northpolesec.santa",
    description: "Groups every version of the same app when its bundle ID and team stay the same, letting you treat that app together across updates.",
  },
  TEAMID: {
    label: "Team ID",
    placeholder: "ZMCG7MLDV9",
    example: "ZMCG7MLDV9",
    description: "Applies to every app from one developer team, useful when you want to trust or block that entire publisher.",
  },
  CDHASH: {
    label: "CDHash",
    placeholder: "a9fdcbc0427a0a585f91bbc7342c261c8ead1942",
    example: "a9fdcbc0427a0a585f91bbc7342c261c8ead1942",
    description: "Matches the internal hash stored in the code signature so it points to one specific build of the app.",
  },
};

export type RuleTypeEntry = { type: ApplicationRuleType; meta: ApplicationRuleTypeMetadata };

export const signingChainReference = applicationRuleTypeMetadata.CERTIFICATE;
export const primaryRuleTypeEntries: RuleTypeEntry[] = applicationRuleTypes
  .filter((t) => applicationRuleTypeMetadata[t].referenceGroup !== "signingChain")
  .map((t) => ({ type: t, meta: applicationRuleTypeMetadata[t] }));
