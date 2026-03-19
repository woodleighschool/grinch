import { CLIENT_MODE_CHOICES, RULE_SYNC_STATUS_CHOICES } from "@/resources/machines/choices";
import type { ReactElement } from "react";
import { DataTable, DateField, List, ReferenceField, SearchInput, SelectArrayInput, TextField } from "react-admin";

const machineFilters = [
  <SearchInput key="search" source="search" alwaysOn />,
  <SelectArrayInput
    key="rule_sync_status"
    source="rule_sync_status"
    label="Rule Sync Status"
    choices={RULE_SYNC_STATUS_CHOICES}
  />,
  <SelectArrayInput key="client_mode" source="client_mode" label="Client Mode" choices={CLIENT_MODE_CHOICES} />,
];

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
