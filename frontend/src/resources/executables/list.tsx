import type { ReactElement } from "react";
import { DataTable, List, NumberField, SearchInput } from "react-admin";

const executableFilters = [<SearchInput key="search" source="search" alwaysOn />];

export const ExecutableList = (): ReactElement => (
  <List sort={{ field: "file_name", order: "DESC" }} filters={executableFilters}>
    <DataTable rowClick="show" bulkActionButtons={false}>
      <DataTable.Col source="file_name" label="File Name" />
      <DataTable.Col source="occurrences" label="Occurrences">
        <NumberField source="occurrences" />
      </DataTable.Col>
      <DataTable.Col source="signing_id" label="Signing ID" />
      <DataTable.Col source="team_id" label="Team ID" />
      {/* <DataTable.Col source="file_sha256" label="SHA-256" /> */}
    </DataTable>
  </List>
);
