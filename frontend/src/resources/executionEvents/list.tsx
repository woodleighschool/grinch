import { EventDecisionField } from "@/resources/shared/decisionField";
import type { ReactElement } from "react";
import { DataTable, DateField, List, ReferenceField, SearchInput, TextField } from "react-admin";

const executionEventFilters = [<SearchInput key="search" source="search" alwaysOn />];

export const ExecutionEventList = (): ReactElement => (
  <List sort={{ field: "occurred_at", order: "DESC" }} filters={executionEventFilters}>
    <DataTable rowClick="show">
      <DataTable.Col source="occurred_at" label="Occurred At">
        <DateField source="occurred_at" showTime />
      </DataTable.Col>
      <DataTable.Col source="decision" label="Decision">
        <EventDecisionField />
      </DataTable.Col>
      <DataTable.Col label="Machine">
        <ReferenceField source="machine_id" reference="machines">
          <TextField source="hostname" />
        </ReferenceField>
      </DataTable.Col>
      <DataTable.Col source="file_name" label="File Name" />
      <DataTable.Col source="signing_id" label="Signing ID" />
    </DataTable>
  </List>
);
