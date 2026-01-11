import type { ReactElement } from "react";
import { DataTable, DateField, Edit, ReferenceManyField, SelectField, TabbedForm } from "react-admin";

import {
  PolicyDetailsFields,
  PolicyRulesInput,
  PolicySettingsFields,
  PolicyTargetsInput,
} from "@/resources/policies/fields";
import { POLICY_STATUS_CHOICES } from "@/api/constants";

export const PolicyEdit = (): ReactElement => (
  <Edit mutationMode="pessimistic" redirect="edit">
    <TabbedForm>
      <TabbedForm.Tab label="Details">
        <PolicyDetailsFields />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Settings">
        <PolicySettingsFields />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Rules">
        <PolicyRulesInput />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Targets">
        <PolicyTargetsInput />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Machines">
        <ReferenceManyField reference="machines" target="policy_id" label={false}>
          <DataTable bulkActionButtons={false} rowClick="show">
            <DataTable.Col source="hostname" label="Hostname" />
            <DataTable.Col source="serial_number" label="Serial Number" />
            <DataTable.Col source="primary_user" label="Primary User" />
            <DataTable.Col source="policy_status" label="Policy Status">
              <SelectField source="policy_status" choices={POLICY_STATUS_CHOICES} optionText="name" />
            </DataTable.Col>
            <DataTable.Col source="last_seen" label="Last Seen">
              <DateField source="last_seen" showTime />
            </DataTable.Col>
          </DataTable>
        </ReferenceManyField>
      </TabbedForm.Tab>
    </TabbedForm>
  </Edit>
);
