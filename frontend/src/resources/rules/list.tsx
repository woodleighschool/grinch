import { RULE_TYPE_CHOICES } from "@/resources/rules/fields";
import type { ReactElement } from "react";
import { DataTable, List, SearchInput, SelectArrayInput, SelectField } from "react-admin";

const RULE_ENABLED_CHOICES = [
  { id: true, name: "Enabled" },
  { id: false, name: "Disabled" },
];

const ruleFilters = [
  <SearchInput key="search" source="search" alwaysOn />,
  <SelectArrayInput key="enabled" source="enabled" label="Enabled" choices={RULE_ENABLED_CHOICES} />,
  <SelectArrayInput key="rule_type" source="rule_type" label="Rule Type" choices={RULE_TYPE_CHOICES} />,
];

export const RuleList = (): ReactElement => (
  <List sort={{ field: "name", order: "ASC" }} filters={ruleFilters} perPage={25}>
    <DataTable rowClick="edit">
      <DataTable.Col source="name" label="Name" />
      <DataTable.Col source="enabled" label="Enabled" />
      <DataTable.Col source="rule_type" label="Rule Type">
        <SelectField source="rule_type" choices={RULE_TYPE_CHOICES} optionText="name" />
      </DataTable.Col>
      <DataTable.Col source="identifier" label="Identifier" />
      <DataTable.Col source="description" label="Description" />
    </DataTable>
  </List>
);
