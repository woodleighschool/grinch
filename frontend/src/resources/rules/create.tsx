import { RuleDetailsFields, RuleTargetsFields } from "@/resources/rules/fields";
import type { ReactElement } from "react";
import { Create, ListButton, TabbedForm, TopToolbar } from "react-admin";

const defaultRuleValues = {
  enabled: true,
  rule_type: "binary",
  targets: {
    include: [],
    exclude: [],
  },
} as const;

const RuleCreateActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const RuleCreate = (): ReactElement => (
  <Create mutationMode="pessimistic" redirect="edit" actions={<RuleCreateActions />}>
    <TabbedForm defaultValues={defaultRuleValues}>
      <TabbedForm.Tab label="Details">
        <RuleDetailsFields />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Targets">
        <RuleTargetsFields />
      </TabbedForm.Tab>
    </TabbedForm>
  </Create>
);
