import { CLIENT_MODE_CHOICES, DECISION_CHOICES, POLICY_STATUS_CHOICES } from "@/api/constants";
import type { ReactElement } from "react";
import {
  BooleanField,
  DataTable,
  DateField,
  ListButton,
  NumberField,
  Pagination,
  ReferenceField,
  ReferenceManyField,
  SelectField,
  Show,
  TabbedShowLayout,
  TextField,
  TopToolbar,
} from "react-admin";

const MachineShowActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const MachineShow = (): ReactElement => (
  <Show actions={<MachineShowActions />}>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <TextField source="hostname" label="Hostname" />
        <TextField source="serial_number" label="Serial Number" />
        <ReferenceField source="user_id" reference="users" label="Primary User" link="show">
          <TextField source="display_name" />
        </ReferenceField>
        <TextField source="model" label="Model" />
        <TextField source="os_version" label="OS Version" />
        <TextField source="os_build" label="OS Build" />
        <TextField source="santa_version" label="Santa Version" />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Sync State">
        <DateField source="last_seen" label="Last Seen" showTime />
        <SelectField source="policy_status" label="Policy Status" choices={POLICY_STATUS_CHOICES} optionText="name" />
        <ReferenceField source="policy_id" reference="policies" label="Policy" link="edit">
          <TextField source="name" />
        </ReferenceField>
        <NumberField source="binary_rule_count" label="Binary Rules" />
        <NumberField source="certificate_rule_count" label="Certificate Rules" />
        <NumberField source="compiler_rule_count" label="Compiler Rules" />
        <NumberField source="transitive_rule_count" label="Transitive Rules" />
        <NumberField source="team_id_rule_count" label="Team ID Rules" />
        <NumberField source="signing_id_rule_count" label="Signing ID Rules" />
        <NumberField source="cdhash_rule_count" label="CDHash Rules" />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Client Configuration">
        <SelectField source="client_mode" label="Client Mode" choices={CLIENT_MODE_CHOICES} optionText="name" />
        <TextField source="push_token" label="Push Token" />
        <NumberField source="sip_status" label="SIP Status" />
        <BooleanField source="request_clean_sync" label="Request Clean Sync" />
        <BooleanField source="push_notification_sync" label="Push Notification Sync" />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Events">
        <ReferenceManyField reference="events" target="machine_id" pagination={<Pagination />}>
          <DataTable rowClick="show" bulkActionButtons={false}>
            <DataTable.Col source="execution_time" label="Time">
              <DateField source="execution_time" showTime />
            </DataTable.Col>
            <DataTable.Col source="decision" label="Decision">
              <SelectField source="decision" choices={DECISION_CHOICES} optionText="name" />
            </DataTable.Col>
            <DataTable.Col source="file_path" label="Path" />
            <DataTable.Col source="signing_id" label="Signing ID" />
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
