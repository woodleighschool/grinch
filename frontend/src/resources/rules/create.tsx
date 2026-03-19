import { RuleDetailsFields, RuleTargetsFields } from "@/resources/rules/fields";
import { ShowActions } from "@/resources/shared/actions";
import type { ReactElement } from "react";
import { Create, TabbedForm } from "react-admin";

const defaultRuleValues = {
  enabled: true,
  rule_type: "binary",
  targets: {
    include: [],
    exclude: [],
  },
} as const;

export const RuleCreate = (): ReactElement => (
  <Create mutationMode="pessimistic" redirect="edit" actions={<ShowActions />}>
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
