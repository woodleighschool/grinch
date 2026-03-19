import { RULE_POLICY_CHOICES } from "@/resources/rules/choices";
import { RuleDetailsFields, RuleTargetsFields } from "@/resources/rules/fields";
import { ShowActions } from "@/resources/shared/actions";
import type { ReactElement } from "react";
import {
  BooleanField,
  DataTable,
  Edit,
  Pagination,
  ReferenceField,
  ReferenceManyField,
  SelectField,
  TabbedForm,
  TextField,
} from "react-admin";

export const RuleEdit = (): ReactElement => (
  <Edit mutationMode="pessimistic" redirect="edit" actions={<ShowActions />}>
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
            <DataTable.Col source="machine_id" label="Machine">
              <ReferenceField source="machine_id" reference="machines" label="Machine">
                <TextField source="hostname" />
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
