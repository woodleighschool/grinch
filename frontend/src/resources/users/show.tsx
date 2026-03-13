import type { EventDecision } from "@/api/types";
import { EVENT_DECISION_DESCRIPTIONS, eventDecisionName } from "@/resources/executionEvents/choices";
import { DecisionChip } from "@/resources/shared/decisionField";
import { ResourceLink } from "@/resources/shared/resourceLinks";
import { SourceField } from "@/resources/shared/sourceField";
import type { ReactElement } from "react";
import {
  DataTable,
  DateField,
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

const UserShowActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const UserShow = (): ReactElement => (
  <Show actions={<UserShowActions />}>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <TextField source="display_name" label="Name" />
        <TextField source="upn" label="UPN" />
        <TextField source="source" label="Source" />
        <DateField source="created_at" label="Created" showTime />
        <DateField source="updated_at" label="Updated" showTime />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Groups">
        <ReferenceManyField reference="group-memberships" target="user_id" pagination={<Pagination />}>
          <DataTable bulkActionButtons={false}>
            <DataTable.Col source="group.name" label="Name">
              <GroupLinkField />
            </DataTable.Col>
            <DataTable.Col source="group.source" label="Source">
              <GroupSourceField />
            </DataTable.Col>
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Machines">
        <ReferenceManyField reference="machines" target="user_id" pagination={<Pagination />}>
          <DataTable rowClick="show" bulkActionButtons={false}>
            <DataTable.Col source="hostname" label="Hostname" />
            <DataTable.Col source="serial_number" label="Serial Number" />
            <DataTable.Col source="os_version" label="OS Version" />
            <DataTable.Col source="last_seen_at" label="Last Seen">
              <DateField source="last_seen_at" showTime />
            </DataTable.Col>
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Events">
        <ReferenceManyField reference="execution-events" target="user_id" pagination={<Pagination />}>
          <DataTable rowClick="show" bulkActionButtons={false}>
            <DataTable.Col source="created_at" label="Ingested At">
              <DateField source="created_at" showTime />
            </DataTable.Col>
            <DataTable.Col label="Machine">
              <ReferenceField source="machine_id" reference="machines" label="Machine">
                <TextField source="hostname" />
              </ReferenceField>
            </DataTable.Col>
            <DataTable.Col source="file_name" label="File Name" />
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
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
