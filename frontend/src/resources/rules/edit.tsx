import { RULE_POLICY_CHOICES, RuleDetailsFields, RuleTargetsFields } from "@/resources/rules/fields";
import type { ReactElement } from "react";
import {
  BooleanField,
  DataTable,
  Edit,
  ListButton,
  Pagination,
  ReferenceField,
  ReferenceManyField,
  SelectField,
  TabbedForm,
  TextField,
  TopToolbar,
} from "react-admin";

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
      <TabbedForm.Tab label="Machines">
        <ReferenceManyField reference="rule-machines" target="rule_id" pagination={<Pagination />}>
          <DataTable bulkActionButtons={false}>
            <DataTable.Col source="machine_id" label="Hostname">
              <ReferenceField source="machine_id" reference="machines" label="Hostname">
                <TextField source="hostname" />
              </ReferenceField>
            </DataTable.Col>
            <DataTable.Col source="machine_id" label="Serial Number">
              <ReferenceField source="machine_id" reference="machines" label="Serial Number">
                <TextField source="serial_number" />
              </ReferenceField>
            </DataTable.Col>
            <DataTable.Col source="policy" label="Policy">
              <SelectField source="policy" choices={RULE_POLICY_CHOICES} optionText="name" />
            </DataTable.Col>
            <DataTable.Col source="applied" label="Applied">
              <BooleanField source="applied" />
            </DataTable.Col>
          </DataTable>
        </ReferenceManyField>
      </TabbedForm.Tab>
    </TabbedForm>
  </Edit>
);
