import { Chip, Tooltip } from "@mui/material";
import type { ReactElement } from "react";

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
