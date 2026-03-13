import { RuleFields } from "@/resources/rules/fields";
import { RuleTargetsTab } from "@/resources/rules/targetsTab";
import type { ReactElement } from "react";
import { Edit, ListButton, TabbedForm, TopToolbar } from "react-admin";

const RuleEditActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const RuleEdit = (): ReactElement => (
  <Edit mutationMode="optimistic" redirect="edit" actions={<RuleEditActions />}>
    <TabbedForm>
      <TabbedForm.Tab label="Overview">
        <RuleFields />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Targets">
        <RuleTargetsTab />
      </TabbedForm.Tab>
    </TabbedForm>
  </Edit>
);
