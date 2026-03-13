import { SourceField } from "@/resources/shared/sourceField";
import type { ReactElement } from "react";
import { DataTable, List, SearchInput } from "react-admin";

const userFilters = [<SearchInput key="search" source="search" alwaysOn />];

export const UserList = (): ReactElement => (
  <List sort={{ field: "display_name", order: "ASC" }} filters={userFilters}>
    <DataTable
      rowClick="show"
      isRowSelectable={(record): boolean => (record as { source?: string }).source === "local"}
    >
      <DataTable.Col source="display_name" label="Name" />
      <DataTable.Col source="upn" label="UPN" />
      <DataTable.Col source="source" label="Source">
        <SourceField />
      </DataTable.Col>
    </DataTable>
  </List>
);
