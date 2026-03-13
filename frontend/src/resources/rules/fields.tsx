import type { RuleType } from "@/api/types";
import { ruleIdentifierValidator, trimmedRequired } from "@/resources/shared/validation";
import { Typography } from "@mui/material";
import type { ReactElement } from "react";
import { FormDataConsumer, SelectInput, TextInput, required } from "react-admin";

export const RULE_TYPE_CHOICES = [
  { id: "binary", name: "Binary" },
  { id: "certificate", name: "Certificate" },
  { id: "team_id", name: "Team ID" },
  { id: "signing_id", name: "Signing ID" },
  { id: "cd_hash", name: "CD Hash" },
] as { id: RuleType; name: string }[];

const RULE_TYPE_DESCRIPTION: Record<RuleType, string> = {
  binary: "SHA-256 hash of the exact binary.",
  certificate: "SHA-256 hash of the signing certificate.",
  team_id: "10-character Apple Team ID.",
  signing_id: "Signing identifier with team or platform prefix.",
  cd_hash: "Code directory hash of the binary.",
};

const IDENTIFIER_PLACEHOLDER: Record<RuleType, string> = {
  binary: "fc6679da622c3ff38933220b8e73c7322ecdc94b4570c50ecab0da311b292682",
  certificate: "7ae80b9ab38af0c63a9a81765f434d9a7cd8f720eb6037ef303de39d779bc258",
  team_id: "EQHXZ8M8AV",
  signing_id: "UBF8T346G9:com.microsoft.VSCode",
  cd_hash: "dbe8c39801f93e05fc7bc53a02af5b4d3cfc670a",
};

export const RuleFields = (): ReactElement => (
  <>
    <TextInput
      source="name"
      label="Name"
      placeholder="Visual Studio Code"
      validate={[trimmedRequired("Name")]}
      fullWidth
    />
    <TextInput
      source="description"
      label="Description"
      placeholder="VSCode has been used to bypass terminal restrictions."
      multiline
      minRows={2}
      fullWidth
    />
    <Typography variant="body2" color="text">
      These fields are part of the payload sent to machines and are included in rule hash comparison.
    </Typography>
    <FormDataConsumer>
      {({ formData }): ReactElement => {
        const values = formData as { rule_type?: RuleType };
        return (
          <SelectInput
            source="rule_type"
            label="Rule Type"
            choices={RULE_TYPE_CHOICES}
            helperText={values.rule_type ? RULE_TYPE_DESCRIPTION[values.rule_type] : "Choose a rule type."}
            validate={[required()]}
            fullWidth
          />
        );
      }}
    </FormDataConsumer>
    <FormDataConsumer>
      {({ formData }): ReactElement => {
        const values = formData as { rule_type?: RuleType };
        return (
          <TextInput
            source="identifier"
            label="Identifier"
            helperText="Identifier for the selected rule type."
            placeholder={values.rule_type ? IDENTIFIER_PLACEHOLDER[values.rule_type] : ""}
            validate={[trimmedRequired("Identifier"), ruleIdentifierValidator]}
            fullWidth
          />
        );
      }}
    </FormDataConsumer>
    <TextInput
      source="custom_message"
      label="Block Message"
      multiline
      minRows={2}
      helperText="Shown when this rule blocks execution."
      placeholder="This app is not approved. Contact IT."
      fullWidth
    />
    <TextInput
      source="custom_url"
      label="Block Help URL"
      helperText="Help link shown when this rule blocks execution."
      placeholder="https://helpdesk.example.com/software?app=%bundle_or_file_identifier%"
      fullWidth
    />
  </>
);
