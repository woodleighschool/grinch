import type { components } from "@/api/openapi";
import {
  EVENT_DECISION_DESCRIPTIONS,
  FILE_ACCESS_DECISION_DESCRIPTIONS,
  eventDecisionName,
  fileAccessDecisionName,
} from "@/resources/shared/decisionChoices";
import { Chip, Tooltip } from "@mui/material";
import type { ReactElement } from "react";
import { useRecordContext } from "react-admin";

type EventDecision = components["schemas"]["EventDecision"];
type FileAccessDecision = components["schemas"]["FileAccessDecision"];

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

export const EventDecisionField = (): ReactElement | undefined => {
  const record = useRecordContext<{ decision: EventDecision }>();

  if (!record) {
    return undefined;
  }

  return (
    <DecisionChip
      decision={record.decision}
      label={eventDecisionName(record.decision)}
      description={EVENT_DECISION_DESCRIPTIONS[record.decision]}
    />
  );
};

export const FileAccessDecisionField = (): ReactElement | undefined => {
  const record = useRecordContext<{ decision: FileAccessDecision }>();

  if (!record) {
    return undefined;
  }

  return (
    <DecisionChip
      decision={record.decision}
      label={fileAccessDecisionName(record.decision)}
      description={FILE_ACCESS_DECISION_DESCRIPTIONS[record.decision]}
    />
  );
};
