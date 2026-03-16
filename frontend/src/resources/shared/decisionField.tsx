import type { EventDecision, FileAccessDecision } from "@/api/types";
import { EVENT_DECISION_DESCRIPTIONS, eventDecisionName } from "@/resources/executionEvents/choices";
import { FILE_ACCESS_DECISION_DESCRIPTIONS, fileAccessDecisionName } from "@/resources/fileAccessEvents/choices";
import { Chip, Tooltip } from "@mui/material";
import type { ReactElement } from "react";
import { useRecordContext } from "react-admin";

const getDecisionColor = (value: string): "default" | "error" | "info" | "success" | "warning" => {
  if (value.startsWith("allow") || value === "audit_only") {
    return "success";
  }

  if (value.startsWith("block") || value.startsWith("denied")) {
    return "error";
  }

  if (value.startsWith("bundle")) {
    return "info";
  }

  return "default";
};

interface DecisionChipProperties {
  decision: string;
  label: string;
  description: string;
}

export const DecisionChip = ({ decision, label, description }: DecisionChipProperties): ReactElement => (
  <Tooltip title={description} arrow>
    <Chip size="small" label={label} color={getDecisionColor(decision)} />
  </Tooltip>
);

export const EventDecisionField = (): ReactElement => {
  const record = useRecordContext<{ decision: EventDecision }>();

  if (!record) {
    return <></>;
  }

  return (
    <DecisionChip
      decision={record.decision}
      label={eventDecisionName(record.decision)}
      description={EVENT_DECISION_DESCRIPTIONS[record.decision]}
    />
  );
};

export const FileAccessDecisionField = (): ReactElement => {
  const record = useRecordContext<{ decision: FileAccessDecision }>();

  if (!record) {
    return <></>;
  }

  return (
    <DecisionChip
      decision={record.decision}
      label={fileAccessDecisionName(record.decision)}
      description={FILE_ACCESS_DECISION_DESCRIPTIONS[record.decision]}
    />
  );
};
