import type { EventDecision, FileAccessDecision } from "@/api/types";
import { EVENT_DECISION_DESCRIPTIONS, eventDecisionName } from "@/resources/executionEvents/choices";
import { FILE_ACCESS_DECISION_DESCRIPTIONS, fileAccessDecisionName } from "@/resources/fileAccessEvents/choices";
import { DecisionChip } from "@/resources/shared/decisionField";
import { ResourceLink } from "@/resources/shared/resourceLinks";
import { SourceField } from "@/resources/shared/sourceField";
import type { ReactElement } from "react";
import {
  BooleanField,
  DataTable,
  DateField,
  DeleteButton,
  FunctionField,
  ListButton,
  Pagination,
  RecordContextProvider,
  ReferenceField,
  ReferenceManyField,
  Show,
  TabbedShowLayout,
  TextField,
  TopToolbar,
  useRecordContext,
} from "react-admin";

interface GroupMembershipRecord {
  group: {
    id?: string | null;
    name?: string | null;
    source?: string;
  };
}

const GroupLinkField = (): ReactElement => {
  const membership = useRecordContext<GroupMembershipRecord>();

  if (!membership) {
    return <></>;
  }

  return <ResourceLink resource="groups" id={membership.group.id} label={membership.group.name} />;
};

const GroupSourceField = (): ReactElement => {
  const membership = useRecordContext<GroupMembershipRecord>();

  if (!membership) {
    return <></>;
  }

  return (
    <RecordContextProvider value={{ source: membership.group.source }}>
      <SourceField />
    </RecordContextProvider>
  );
};

const MachineShowActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
    <DeleteButton redirect="list" mutationMode="optimistic" />
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
        <BooleanField source="request_clean_sync" label="Request Clean Sync" />
        <DateField source="last_seen_at" label="Last Seen" showTime />
        <DateField source="created_at" label="Created" showTime />
        <DateField source="updated_at" label="Updated" showTime />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Groups">
        <ReferenceManyField reference="group-memberships" target="machine_id" pagination={<Pagination />}>
          <DataTable bulkActionButtons={false}>
            <DataTable.Col source="group.name" label="Name">
              <GroupLinkField />
            </DataTable.Col>
            <DataTable.Col source="group.source" label="Source">
              <GroupSourceField />
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
              <FunctionField
                render={(record): ReactElement => {
                  const { decision } = record as { decision: EventDecision };
                  return (
                    <DecisionChip
                      decision={decision}
                      label={eventDecisionName(decision)}
                      description={EVENT_DECISION_DESCRIPTIONS[decision]}
                    />
                  );
                }}
              />
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
              <FunctionField
                render={(record): ReactElement => {
                  const { decision } = record as { decision: FileAccessDecision };
                  return (
                    <DecisionChip
                      decision={decision}
                      label={fileAccessDecisionName(decision)}
                      description={FILE_ACCESS_DECISION_DESCRIPTIONS[decision]}
                    />
                  );
                }}
              />
            </DataTable.Col>
            <DataTable.Col source="file_name" label="Primary Process" />
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
