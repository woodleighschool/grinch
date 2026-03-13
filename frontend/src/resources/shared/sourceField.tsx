import MicrosoftIcon from "@mui/icons-material/Microsoft";
import StorageOutlinedIcon from "@mui/icons-material/StorageOutlined";
import { Tooltip } from "@mui/material";
import type { ReactElement } from "react";
import { useRecordContext } from "react-admin";

export const SourceField = (): ReactElement => {
  const record = useRecordContext<{ source?: string }>();

  if (!record) {
    return <></>;
  }

  if (record.source === "entra") {
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
