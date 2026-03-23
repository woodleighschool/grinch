import { EditableShowActions } from "@/resources/shared/actions";
import { ExecutionDecisionField } from "@/resources/shared/decisionField";
import { ExecutableEntitlementsArrayField, SigningChainArrayField } from "@/resources/shared/executableFields";
import type { ReactElement } from "react";
import { DateField, Labeled, ReferenceField, Show, TabbedShowLayout, TextArrayField, TextField } from "react-admin";

export const ExecutionEventShow = (): ReactElement => (
  <Show actions={<EditableShowActions />}>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <ReferenceField source="machine_id" reference="machines" label="Machine">
          <TextField source="hostname" />
        </ReferenceField>
        <ReferenceField source="executable_id" reference="executables" label="Executable">
          <TextField source="file_name" />
        </ReferenceField>
        <Labeled label="Decision">
          <ExecutionDecisionField />
        </Labeled>
        <TextField source="executing_user" label="Executing User" />
        <DateField source="occurred_at" label="Occurred At" showTime />
        <DateField source="created_at" label="Ingested At" showTime />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Executable">
        <TextField source="file_path" label="File Path" />
        <TextField source="file_name" label="File Name" />
        <TextField source="file_sha256" label="SHA-256" />
        <TextField source="cdhash" label="CDHash" />
        <TextField source="file_bundle_id" label="Bundle ID" />
        <TextField source="file_bundle_path" label="Bundle Path" />
        <TextField source="signing_id" label="Signing ID" />
        <TextField source="team_id" label="Team ID" />
        <SigningChainArrayField source="signing_chain" />
        <ExecutableEntitlementsArrayField />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Sessions">
        <TextArrayField source="logged_in_users" label="Logged-in Users" />
        <TextArrayField source="current_sessions" label="Current Sessions" />
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
