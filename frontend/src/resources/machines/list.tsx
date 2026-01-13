import { POLICY_STATUS_CHOICES } from "@/api/constants";
import type { ReactElement } from "react";
import { DataTable, DateField, List, ReferenceField, SearchInput, SelectField, TextField } from "react-admin";

const machineFilters: ReactElement[] = [<SearchInput key="q" source="q" alwaysOn />];

export const MachineList = (): ReactElement => (
  <List sort={{ field: "hostname", order: "ASC" }} filters={machineFilters}>
    <DataTable rowClick="show" bulkActionButtons={false}>
      <DataTable.Col source="hostname" label="Hostname" />
      <DataTable.Col source="serial_number" label="Serial Number" />
      <DataTable.Col source="primary_user" label="Primary User" />
      <DataTable.Col source="user_id" label="User">
        <ReferenceField source="user_id" reference="users" link="show">
          <TextField source="display_name" />
        </ReferenceField>
      </DataTable.Col>
      <DataTable.Col source="last_seen" label="Last Seen">
        <DateField source="last_seen" showTime />
      </DataTable.Col>
      <DataTable.Col source="policy_id" label="Policy">
        <ReferenceField source="policy_id" reference="policies" link="edit">
          <TextField source="name" />
        </ReferenceField>
      </DataTable.Col>
      <DataTable.Col source="policy_status" label="Policy Status">
        <SelectField source="policy_status" choices={POLICY_STATUS_CHOICES} optionText="name" />
      </DataTable.Col>
    </DataTable>
  </List>
);
