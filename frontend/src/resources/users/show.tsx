import type { ReactElement } from "react";
import {
  DataTable,
  DateField,
  ListButton,
  Pagination,
  ReferenceField,
  ReferenceManyField,
  Show,
  TabbedShowLayout,
  TextField,
  TopToolbar,
} from "react-admin";

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
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Groups">
        <ReferenceManyField reference="memberships" target="user_id" pagination={<Pagination />}>
          <DataTable bulkActionButtons={false}>
            <DataTable.Col source="group_id" label="Group">
              <ReferenceField source="group_id" reference="groups" link="show">
                <TextField source="display_name" />
              </ReferenceField>
            </DataTable.Col>
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Machines">
        <ReferenceManyField reference="machines" target="user_id">
          <DataTable rowClick="show" bulkActionButtons={false}>
            <DataTable.Col source="hostname" label="Hostname" />
            <DataTable.Col source="serial_number" label="Serial Number" />
            <DataTable.Col source="last_seen" label="Last Seen">
              <DateField source="last_seen" showTime />
            </DataTable.Col>
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
