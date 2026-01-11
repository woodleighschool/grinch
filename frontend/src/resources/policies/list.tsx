import type { ReactElement } from "react";
import { BooleanField, DataTable, List, NumberField, SearchInput } from "react-admin";

const policyFilters: ReactElement[] = [<SearchInput key="q" source="q" alwaysOn />];

export const PolicyList = (): ReactElement => (
  <List sort={{ field: "priority", order: "DESC" }} filters={policyFilters}>
    <DataTable rowClick="edit" bulkActionButtons={false}>
      <DataTable.Col source="name" label="Name" />
      <DataTable.Col source="description" label="Description" />
      <DataTable.Col source="priority" label="Priority">
        <NumberField source="priority" />
      </DataTable.Col>
      <DataTable.Col source="enabled" label="Enabled">
        <BooleanField source="enabled" />
      </DataTable.Col>
    </DataTable>
  </List>
);
