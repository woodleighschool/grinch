import type { ReactElement } from "react";
import { DataTable, List, SearchInput } from "react-admin";

const userFilters: ReactElement[] = [<SearchInput key="q" source="q" alwaysOn />];

export const UserList = (): ReactElement => (
  <List sort={{ field: "display_name", order: "ASC" }} filters={userFilters}>
    <DataTable rowClick="show" bulkActionButtons={false}>
      <DataTable.Col source="display_name" label="Name" />
      <DataTable.Col source="upn" label="UPN" />
    </DataTable>
  </List>
);
