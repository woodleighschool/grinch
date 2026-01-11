import { DECISION_CHOICES, SIGNING_STATUS_CHOICES } from "@/api/constants";
import type { ReactElement } from "react";
import {
  ArrayField,
  ChipField,
  DataTable,
  DateField,
  NumberField,
  ReferenceField,
  SelectField,
  Show,
  TabbedShowLayout,
  TextArrayField,
  TextField,
} from "react-admin";

export const EventShow = (): ReactElement => (
  <Show>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <SelectField source="decision" label="Decision" choices={DECISION_CHOICES} optionText="name" />
        <DateField source="execution_time" label="Execution Time" showTime />
        <ReferenceField source="machine_id" reference="machines" label="Machine" link="show">
          <TextField source="hostname" />
        </ReferenceField>
        <TextField source="executing_user" label="Executing User" />
      </TabbedShowLayout.Tab>

      <TabbedShowLayout.Tab label="Process">
        <TextField source="parent_name" label="Parent Process" />
        <NumberField source="pid" label="PID" options={{ useGrouping: false }} />
        <NumberField source="ppid" label="PPID" options={{ useGrouping: false }} />
        <NumberField source="cs_flags" label="CS Flags" options={{ useGrouping: false }} />
        <SelectField
          source="signing_status"
          label="Signing Status"
          choices={SIGNING_STATUS_CHOICES}
          optionText="name"
        />
      </TabbedShowLayout.Tab>

      <TabbedShowLayout.Tab label="File">
        <TextField source="file_name" label="File Name" />
        <TextField source="file_path" label="File Path" />
        <TextField source="file_sha256" label="SHA-256" />
        <TextField source="cdhash" label="CDHash" />
      </TabbedShowLayout.Tab>

      <TabbedShowLayout.Tab label="Bundle">
        <TextField source="file_bundle_name" label="Bundle Name" />
        <TextField source="file_bundle_id" label="Bundle ID" />
        <TextField source="file_bundle_path" label="Bundle Path" />
        <TextField source="file_bundle_version" label="Bundle Version" />
        <TextField source="file_bundle_version_string" label="Bundle Version (string)" />
        <NumberField source="file_bundle_binary_count" label="Binary Count" />
      </TabbedShowLayout.Tab>

      <TabbedShowLayout.Tab label="Signing">
        <TextField source="signing_id" label="Signing ID" />
        <TextField source="team_id" label="Team ID" />
        <SelectField
          source="signing_status"
          label="Signing Status"
          choices={SIGNING_STATUS_CHOICES}
          optionText="name"
        />
        <DateField source="signing_time" label="Signing Time" showTime />
        <DateField source="secure_signing_time" label="Secure Signing Time" showTime />
        <ArrayField source="signing_chain">
          <DataTable bulkActionButtons={false} rowClick={false}>
            <DataTable.Col source="sha256" label="SHA256" />
            <DataTable.Col source="cn" label="CN" />
            <DataTable.Col source="org" label="Org" />
            <DataTable.Col source="ou" label="OU" />
            <DataTable.Col source="valid_from" label="Valid From">
              <DateField source="valid_from" showTime />
            </DataTable.Col>
            <DataTable.Col source="valid_until" label="Valid Until">
              <DateField source="valid_until" showTime />
            </DataTable.Col>
          </DataTable>
        </ArrayField>
      </TabbedShowLayout.Tab>

      <TabbedShowLayout.Tab label="Entitlements">
        <ArrayField source="entitlements">
          <DataTable bulkActionButtons={false} rowClick={false}>
            <DataTable.Col source="key" label="Key">
              <ChipField source="key" />
            </DataTable.Col>
            <DataTable.Col source="value" label="Value" />
          </DataTable>
        </ArrayField>
      </TabbedShowLayout.Tab>

      <TabbedShowLayout.Tab label="Sessions">
        <TextArrayField source="logged_in_users" label="Logged-in Users" />
        <TextArrayField source="current_sessions" label="Current Sessions" />
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
