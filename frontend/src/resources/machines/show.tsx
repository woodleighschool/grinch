import { CLIENT_MODE_CHOICES, RULE_SYNC_STATUS_CHOICES } from "@/resources/machines/choices";
import { RULE_POLICY_CHOICES } from "@/resources/rules/fields";
import { EventDecisionField, FileAccessDecisionField } from "@/resources/shared/decisionField";
import {
  GroupMembershipGroupLinkField,
  GroupMembershipGroupSourceField,
} from "@/resources/shared/groupMembershipFields";
import type { ReactElement } from "react";
import {
  BooleanField,
  DataTable,
  DateField,
  DeleteButton,
  ListButton,
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
    <DeleteButton redirect="list" mutationMode="pessimistic" />
  </TopToolbar>
);

export const MachineShow = (): ReactElement => (
  <Show actions={<MachineShowActions />}>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <TextField source="hostname" label="Hostname" />
        <TextField source="serial_number" label="Serial Number" />
        <TextField source="model_identifier" label="Model Identifier" />
        <TextField source="os_version" label="OS Version" />
        <TextField source="os_build" label="OS Build" />
        <TextField source="santa_version" label="Santa Version" />
        <ReferenceField source="primary_user_id" reference="users" label="Primary User">
          <TextField source="display_name" />
        </ReferenceField>
        <TextField source="primary_user" label="Primary User UPN" />
        <SelectField source="rule_sync_status" label="Rule Sync Status" choices={RULE_SYNC_STATUS_CHOICES} />
        <SelectField source="client_mode" label="Client Mode" choices={CLIENT_MODE_CHOICES} />
        <TextField source="binary_rule_count" label="Binary Rules" />
        <TextField source="certificate_rule_count" label="Certificate Rules" />
        <TextField source="compiler_rule_count" label="Compiler Rules" />
        <TextField source="transitive_rule_count" label="Transitive Rules" />
        <TextField source="teamid_rule_count" label="Team ID Rules" />
        <TextField source="signingid_rule_count" label="Signing ID Rules" />
        <TextField source="cdhash_rule_count" label="CD Hash Rules" />
        <DateField source="last_seen_at" label="Last Seen" showTime />
        <DateField source="created_at" label="Created" showTime />
        <DateField source="updated_at" label="Updated" showTime />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Rules">
        <ReferenceManyField reference="machine-rules" target="machine_id" pagination={<Pagination />}>
          <DataTable bulkActionButtons={false}>
            <DataTable.Col source="rule_id" label="Rule">
              <ReferenceField source="rule_id" reference="rules" label="Rule">
                <TextField source="name" />
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
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Groups">
        <ReferenceManyField reference="group-memberships" target="machine_id" pagination={<Pagination />}>
          <DataTable bulkActionButtons={false}>
            <DataTable.Col source="group.name" label="Name">
              <GroupMembershipGroupLinkField />
            </DataTable.Col>
            <DataTable.Col source="group.source" label="Source">
              <GroupMembershipGroupSourceField />
            </DataTable.Col>
            <DataTable.Col source="kind" label="Membership" />
            <DataTable.Col source="member.name" label="Via" />
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Events">
        <ReferenceManyField reference="execution-events" target="machine_id" pagination={<Pagination />}>
          <DataTable rowClick="show" bulkActionButtons={false}>
            <DataTable.Col source="created_at" label="Ingested At">
              <DateField source="created_at" showTime />
            </DataTable.Col>
            <DataTable.Col source="file_name" label="File" />
            <DataTable.Col source="decision" label="Decision">
              <EventDecisionField />
            </DataTable.Col>
            <DataTable.Col source="signing_id" label="Signing ID" />
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="File Access">
        <ReferenceManyField reference="file-access-events" target="machine_id" pagination={<Pagination />}>
          <DataTable rowClick="show" bulkActionButtons={false}>
            <DataTable.Col source="created_at" label="Ingested At">
              <DateField source="created_at" showTime />
            </DataTable.Col>
            <DataTable.Col source="rule_name" label="Rule Name" />
            <DataTable.Col source="target" label="Target" />
            <DataTable.Col source="decision" label="Decision">
              <FileAccessDecisionField />
            </DataTable.Col>
            <DataTable.Col source="file_name" label="Primary Process" />
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
