import { RULE_TYPE_CHOICES } from "@/api/constants";
import type { ReactElement } from "react";
import { DataTable, List, SearchInput, SelectField, SelectInput } from "react-admin";

const ruleFilters: ReactElement[] = [
  <SearchInput key="q" source="q" alwaysOn />,
  <SelectInput key="rule_type" source="rule_type" choices={RULE_TYPE_CHOICES} optionText="name" />,
];

export const RuleList = (): ReactElement => (
  <List sort={{ field: "name", order: "ASC" }} filters={ruleFilters}>
    <DataTable rowClick="edit" bulkActionButtons={false}>
      <DataTable.Col source="name" label="Name" />
      <DataTable.Col source="rule_type" label="Type">
        <SelectField source="rule_type" choices={RULE_TYPE_CHOICES} optionText="name" />
      </DataTable.Col>
      <DataTable.Col source="identifier" label="Identifier" />
      <DataTable.Col source="description" label="Description" />
    </DataTable>
  </List>
);
