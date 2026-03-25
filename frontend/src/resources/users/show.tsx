import { ShowActions } from "@/resources/shared/actions";
import { ExecutionDecisionField } from "@/resources/shared/decisionField";
import { SourceField } from "@/resources/shared/sourceField";
import type { ReactElement } from "react";
import {
  DataTable,
  DateField,
  Pagination,
  ReferenceArrayField,
  ReferenceField,
  ReferenceManyField,
  Show,
  TabbedShowLayout,
  TextField,
} from "react-admin";

export const UserShow = (): ReactElement => (
  <Show actions={<ShowActions />}>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <TextField source="display_name" label="Name" />
        <TextField source="upn" label="UPN" />
        <TextField source="source" label="Source" />
        <DateField source="created_at" label="Created" showTime />
        <DateField source="updated_at" label="Updated" showTime />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Groups">
        <ReferenceArrayField source="group_ids" reference="groups">
          <DataTable rowClick="show" bulkActionButtons={false}>
            <DataTable.Col source="name" label="Name" />
            <DataTable.Col source="source" label="Source">
              <SourceField />
            </DataTable.Col>
          </DataTable>
        </ReferenceArrayField>
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
              <ExecutionDecisionField />
            </DataTable.Col>
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
