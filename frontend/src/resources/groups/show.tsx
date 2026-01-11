import type { ReactElement } from "react";
import {
  DataTable,
  ListButton,
  NumberField,
  Pagination,
  ReferenceField,
  ReferenceManyField,
  Show,
  TabbedShowLayout,
  TextField,
  TopToolbar,
} from "react-admin";

const GroupShowActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const GroupShow = (): ReactElement => (
  <Show actions={<GroupShowActions />}>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <TextField source="display_name" label="Name" />
        <TextField source="description" label="Description" />
        <NumberField source="member_count" label="Members" />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Members">
        <ReferenceManyField reference="memberships" target="group_id" pagination={<Pagination />}>
          <DataTable bulkActionButtons={false}>
            <DataTable.Col source="user_id" label="User">
              <ReferenceField source="user_id" reference="users" link="show">
                <TextField source="display_name" />
              </ReferenceField>
            </DataTable.Col>
          </DataTable>
        </ReferenceManyField>
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
