import type { ReactElement } from "react";
import { DataTable, List, NumberField, SearchInput } from "react-admin";

const groupFilters: ReactElement[] = [<SearchInput key="q" source="q" alwaysOn />];

export const GroupList = (): ReactElement => (
  <List sort={{ field: "display_name", order: "ASC" }} filters={groupFilters}>
    <DataTable rowClick="show" bulkActionButtons={false}>
      <DataTable.Col source="display_name" label="Name" />
      <DataTable.Col source="description" label="Description" />
      <DataTable.Col source="member_count" label="Members">
        <NumberField source="member_count" />
      </DataTable.Col>
    </DataTable>
  </List>
);
