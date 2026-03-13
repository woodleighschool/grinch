import { SourceField } from "@/resources/shared/sourceField";
import type { ReactElement } from "react";
import { DataTable, List, SearchInput } from "react-admin";

const groupFilters = [<SearchInput key="search" source="search" alwaysOn />];

export const GroupList = (): ReactElement => (
  <List sort={{ field: "name", order: "ASC" }} filters={groupFilters}>
    <DataTable
      rowClick="edit"
      isRowSelectable={(record): boolean => (record as { source?: string }).source === "local"}
    >
      <DataTable.Col source="name" label="Name" />
      <DataTable.Col source="description" label="Description" />
      <DataTable.Col source="member_count" label="Members" />
      <DataTable.Col source="source" label="Source">
        <SourceField />
      </DataTable.Col>
    </DataTable>
  </List>
);
