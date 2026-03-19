import { FILE_ACCESS_DECISION_CHOICES } from "@/resources/fileAccessEvents/choices";
import { FileAccessDecisionField } from "@/resources/shared/decisionField";
import type { ReactElement } from "react";
import { DataTable, DateField, List, ReferenceField, SearchInput, SelectArrayInput, TextField } from "react-admin";

const fileAccessEventFilters = [
  <SearchInput key="search" source="search" alwaysOn />,
  <SelectArrayInput key="decision" source="decision" label="Decision" choices={FILE_ACCESS_DECISION_CHOICES} />,
];

export const FileAccessEventList = (): ReactElement => (
  <List sort={{ field: "created_at", order: "DESC" }} filters={fileAccessEventFilters}>
    <DataTable rowClick="show">
      <DataTable.Col source="occurred_at" label="Occurred At">
        <DateField source="occurred_at" showTime />
      </DataTable.Col>
      <DataTable.Col source="decision" label="Decision">
        <FileAccessDecisionField />
      </DataTable.Col>
      <DataTable.Col label="Machine">
        <ReferenceField source="machine_id" reference="machines">
          <TextField source="hostname" />
        </ReferenceField>
      </DataTable.Col>
      <DataTable.Col source="rule_name" label="Rule Name" />
      <DataTable.Col source="target" label="Target" />
      <DataTable.Col source="file_name" label="Primary Process" />
    </DataTable>
  </List>
);
