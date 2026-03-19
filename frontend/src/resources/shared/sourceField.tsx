import type { components } from "@/api/openapi";
import MicrosoftIcon from "@mui/icons-material/Microsoft";
import StorageOutlinedIcon from "@mui/icons-material/StorageOutlined";
import { Tooltip } from "@mui/material";
import type { ReactElement } from "react";
import { useRecordContext } from "react-admin";

type Source = components["schemas"]["Source"];

interface SourceFieldProperties {
  source?: Source | null | undefined;
}

export const SourceField = ({ source: providedSource }: SourceFieldProperties = {}): ReactElement | undefined => {
  const record = useRecordContext<{ source?: string }>();
  const source = providedSource ?? record?.source;

  if (!source) {
    return undefined;
  }

  if (source === "entra") {
    return (
      <Tooltip title="Entra">
        <MicrosoftIcon fontSize="small" color="action" />
      </Tooltip>
    );
  }

  return (
    <Tooltip title="Local">
      <StorageOutlinedIcon fontSize="small" color="action" />
    </Tooltip>
  );
};
