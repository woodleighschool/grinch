import { RULE_TYPE, enumDescription } from "@/api/constants";
import type { Rule } from "@/api/types";
import type { ReactElement } from "react";
import { FormDataConsumer, SelectInput, TextInput, required } from "react-admin";

const identifierPlaceholder = (t?: number): string => {
  switch (t) {
    case RULE_TYPE.BINARY: {
      return "fc6679da622c3ff38933220b8e73c7322ecdc94b4570c50ecab0da311b292682";
    }
    case RULE_TYPE.CERTIFICATE: {
      return "7ae80b9ab38af0c63a9a81765f434d9a7cd8f720eb6037ef303de39d779bc258";
    }
    case RULE_TYPE.TEAM_ID: {
      return "EQHXZ8M8AV";
    }
    case RULE_TYPE.SIGNING_ID: {
      return "UBF8T346G9:com.microsoft.VSCode";
    }
    case RULE_TYPE.CDHASH: {
      return "dbe8c39801f93e05fc7bc53a02af5b4d3cfc670a";
    }
    default: {
      return "";
    }
  }
};

export const RuleFields = (): ReactElement => (
  <>
    <TextInput
      source="name"
      label="Name"
      validate={[required()]}
      helperText="Unique rule name."
      placeholder="Google Chrome"
    />

    <TextInput source="description" label="Description" multiline minRows={2} />

    <FormDataConsumer<Partial<Rule>>>
      {({ formData }): ReactElement => (
        <SelectInput
          source="rule_type"
          label="Rule Type"
          choices={RULE_TYPE.choices("BINARY", "CERTIFICATE", "TEAM_ID", "SIGNING_ID", "CDHASH")}
          validate={[required()]}
          helperText={enumDescription(RULE_TYPE, formData.rule_type)}
        />
      )}
    </FormDataConsumer>

    <FormDataConsumer<Partial<Rule>>>
      {({ formData }): ReactElement => (
        <TextInput
          source="identifier"
          label="Identifier"
          validate={[required()]}
          helperText="Identifier for the selected rule type."
          placeholder={identifierPlaceholder(formData.rule_type)}
        />
      )}
    </FormDataConsumer>

    <TextInput
      source="custom_msg"
      label="Block Message"
      multiline
      minRows={2}
      helperText="Shown when this rule blocks execution."
      placeholder="This app is not approved. Contact IT."
    />
    <TextInput
      source="custom_url"
      label="Block Help URL"
      helperText="Help link shown when this rule blocks execution."
      placeholder="https://helpdesk.example.com/software?app=%bundle_or_file_identifier%"
    />
    <TextInput
      source="notification_app_name"
      label="Allow Notification App Name"
      helperText="App name shown in allow notifications."
      placeholder="Google Chrome"
    />
  </>
);
