import { RuleFields } from "@/resources/rules/fields";
import type { ReactElement } from "react";
import { Create, ListButton, TabbedForm, TopToolbar } from "react-admin";

const defaultRuleValues = {
  rule_type: "binary",
} as const;

const RuleCreateActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const RuleCreate = (): ReactElement => (
  <Create redirect="edit" actions={<RuleCreateActions />}>
    <TabbedForm defaultValues={defaultRuleValues}>
      <TabbedForm.Tab label="Overview">
        <RuleFields />
      </TabbedForm.Tab>
    </TabbedForm>
  </Create>
);
