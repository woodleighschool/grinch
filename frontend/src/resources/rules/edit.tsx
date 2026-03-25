import { RULE_POLICY_CHOICES } from "@/resources/rules/choices";
import { RuleDetailsFields, RuleTargetsFields } from "@/resources/rules/fields";
import { ShowActions } from "@/resources/shared/actions";
import type { ReactElement } from "react";
import {
  ArrayField,
  BooleanField,
  DataTable,
  Edit,
  ReferenceField,
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
        <ArrayField source="machines">
          <DataTable bulkActionButtons={false} rowClick={false}>
            <DataTable.Col source="machine_id" label="Machine">
              <ReferenceField source="machine_id" reference="machines" link="show">
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
        </ArrayField>
      </TabbedForm.Tab>
    </TabbedForm>
  </Edit>
);
