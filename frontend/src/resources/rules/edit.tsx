import { RuleDetailsFields, RuleTargetsFields } from "@/resources/rules/fields";
import type { ReactElement } from "react";
import { Edit, ListButton, TabbedForm, TopToolbar } from "react-admin";

const RuleEditActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const RuleEdit = (): ReactElement => (
  <Edit mutationMode="pessimistic" redirect="edit" actions={<RuleEditActions />}>
    <TabbedForm>
      <TabbedForm.Tab label="Details">
        <RuleDetailsFields />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Targets">
        <RuleTargetsFields />
      </TabbedForm.Tab>
    </TabbedForm>
  </Edit>
);
