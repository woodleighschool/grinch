import type { ReactElement } from "react";
import { DataTable, DateField, List, ReferenceField, SearchInput, TextField } from "react-admin";

const machineFilters = [<SearchInput key="search" source="search" alwaysOn />];

export const MachineList = (): ReactElement => (
  <List sort={{ field: "last_seen_at", order: "DESC" }} filters={machineFilters}>
    <DataTable rowClick="show">
      <DataTable.Col source="hostname" label="Hostname" />
      <DataTable.Col source="serial_number" label="Serial Number" />
      <DataTable.Col label="Primary User">
        <ReferenceField source="primary_user_id" reference="users">
          <TextField source="display_name" />
        </ReferenceField>
      </DataTable.Col>
      <DataTable.Col source="os_version" label="OS Version" />
      <DataTable.Col source="last_seen_at" label="Last Seen">
        <DateField source="last_seen_at" showTime />
      </DataTable.Col>
    </DataTable>
  </List>
);
