import type { ReactElement } from "react";
import { FormDataConsumer, SelectInput, TextInput, required } from "react-admin";
import { RULE_TYPE, enumDescription } from "@/api/constants";
import type { Rule } from "@/api/types";

export const RuleFields = (): ReactElement => (
  <>
    <TextInput source="name" label="Name" validate={[required()]} helperText="Unique name for this rule." />
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
    <TextInput source="identifier" label="Identifier" validate={[required()]} />
    <TextInput source="custom_msg" label="Custom Message" multiline minRows={2} />
    <TextInput source="custom_url" label="Custom URL" />
    <TextInput source="notification_app_name" label="Notification App Name" />
  </>
);
