import type { ReactElement } from "react";
import {
  DataTable,
  DateField,
  FunctionField,
  List,
  ReferenceField,
  SearchInput,
  SelectField,
  SelectInput,
  TextField,
} from "react-admin";
import { Chip, Tooltip } from "@mui/material";
import { DECISION, DECISION_CHOICES, enumDescription, enumName } from "@/api/constants";
import type { EventRecord } from "@/api/types";

const eventFilters: ReactElement[] = [
  <SearchInput key="q" source="q" alwaysOn />,
  <SelectInput key="decision" source="decision" choices={DECISION_CHOICES} optionText="name" />,
];

export const EventList = (): ReactElement => (
  <List sort={{ field: "execution_time", order: "DESC" }} filters={eventFilters}>
    <DataTable rowClick="show" bulkActionButtons={false}>
      <DataTable.Col source="execution_time" label="Time">
        <DateField source="execution_time" showTime />
      </DataTable.Col>
      <DataTable.Col source="machine_id" label="Machine">
        <ReferenceField source="machine_id" reference="machines" link="show">
          <TextField source="hostname" />
        </ReferenceField>
      </DataTable.Col>
      <DataTable.Col source="decision" label="Decision">
        <FunctionField<EventRecord>
          render={(record): ReactElement => (
            <Tooltip title={enumDescription(DECISION, record.decision)} arrow>
              <Chip label={enumName(DECISION, record.decision)} size="small" />
            </Tooltip>
          )}
        />
      </DataTable.Col>
      <DataTable.Col source="file_name" label="Name" />
      <DataTable.Col source="signing_id" label="Signing ID" />
    </DataTable>
  </List>
);
