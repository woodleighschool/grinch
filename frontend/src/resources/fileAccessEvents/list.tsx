import type { FileAccessDecision } from "@/api/types";
import { FILE_ACCESS_DECISION_DESCRIPTIONS, fileAccessDecisionName } from "@/resources/fileAccessEvents/choices";
import { DecisionChip } from "@/resources/shared/decisionField";
import type { ReactElement } from "react";
import { DataTable, DateField, FunctionField, List, ReferenceField, SearchInput, TextField } from "react-admin";

const fileAccessEventFilters = [<SearchInput key="search" source="search" alwaysOn />];

export const FileAccessEventList = (): ReactElement => (
  <List sort={{ field: "created_at", order: "DESC" }} filters={fileAccessEventFilters}>
    <DataTable rowClick="show">
      <DataTable.Col source="occurred_at" label="Occurred At">
        <DateField source="occurred_at" showTime />
      </DataTable.Col>
      <DataTable.Col source="decision" label="Decision">
        <FunctionField
          render={(record): ReactElement => {
            const { decision } = record as { decision: FileAccessDecision };
            return (
              <DecisionChip
                decision={decision}
                label={fileAccessDecisionName(decision)}
                description={FILE_ACCESS_DECISION_DESCRIPTIONS[decision]}
              />
            );
          }}
        />
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
