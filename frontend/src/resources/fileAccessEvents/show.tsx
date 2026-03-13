import type { FileAccessDecision, FileAccessEvent } from "@/api/types";
import { FILE_ACCESS_DECISION_DESCRIPTIONS, fileAccessDecisionName } from "@/resources/fileAccessEvents/choices";
import { DecisionChip } from "@/resources/shared/decisionField";
import ArrowDownwardIcon from "@mui/icons-material/ArrowDownward";
import { Box, Chip, Paper, Stack, Typography } from "@mui/material";
import type { ReactElement } from "react";
import {
  DateField,
  DeleteButton,
  FunctionField,
  Labeled,
  ListButton,
  ReferenceField,
  Show,
  TabbedShowLayout,
  TextField,
  TopToolbar,
  useRecordContext,
} from "react-admin";

type ProcessChainRow = FileAccessEvent["process_chain"][number] & { id: string; step: number };

const ProcessChainField = (): ReactElement | undefined => {
  const record = useRecordContext<FileAccessEvent>();
  if (!record) {
    return undefined;
  }

  const processChain: ProcessChainRow[] = record.process_chain.map(
    (row, index): ProcessChainRow => ({ ...row, id: String(index), step: index + 1 }),
  );

  return (
    <Labeled label="Process Chain" fullWidth>
      <Stack spacing={1.5}>
        {processChain.map(
          (row, index): ReactElement => (
            <Stack key={row.id} spacing={1}>
              <Paper sx={{ p: 2.25 }}>
                <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5} alignItems={{ sm: "center" }}>
                  <Chip
                    size="small"
                    color="primary"
                    label={`Step ${String(row.step)}`}
                    sx={{ alignSelf: { xs: "flex-start", sm: "center" } }}
                  />
                  <Stack spacing={0.25} sx={{ flex: 1, minWidth: 0 }}>
                    <TextField source="file_name" record={row} />
                    <Typography variant="body2" color="text.secondary">
                      PID {row.pid}
                    </Typography>
                  </Stack>
                  <Typography variant="body2" color="text.secondary" sx={{ minWidth: 0 }}>
                    {row.file_path}
                  </Typography>
                </Stack>
              </Paper>
              {index < processChain.length - 1 ? (
                <Box sx={{ display: "flex", justifyContent: "center", color: "text.secondary" }}>
                  <ArrowDownwardIcon fontSize="small" />
                </Box>
              ) : undefined}
            </Stack>
          ),
        )}
      </Stack>
    </Labeled>
  );
};

const FileAccessEventShowActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
    <DeleteButton redirect="list" mutationMode="optimistic" />
  </TopToolbar>
);

export const FileAccessEventShow = (): ReactElement => (
  <Show actions={<FileAccessEventShowActions />}>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <ReferenceField source="machine_id" reference="machines" label="Machine">
          <TextField source="hostname" />
        </ReferenceField>
        <TextField source="file_name" label="Primary Process" />
        <TextField source="rule_name" label="Rule Name" />
        <TextField source="rule_version" label="Rule Version" />
        <TextField source="target" label="Target" />
        <FunctionField
          label="Decision"
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
        <DateField source="occurred_at" label="Occurred At" showTime />
        <DateField source="created_at" label="Ingested At" showTime />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Process">
        <TextField source="file_name" label="File Name" />
        <TextField source="file_sha256" label="SHA-256" />
        <TextField source="signing_id" label="Signing ID" />
        <TextField source="team_id" label="Team ID" />
        <TextField source="cdhash" label="CDHash" />
        <ProcessChainField />
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
