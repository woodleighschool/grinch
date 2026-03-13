import { Link as MuiLink } from "@mui/material";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

interface ResourceLinkProperties {
  id?: string | null | undefined;
  label?: string | null | undefined;
  resource: string;
}

export const ResourceLink = ({ id, label, resource }: ResourceLinkProperties): ReactElement => {
  if (!id || !label || label.trim() === "") {
    return <></>;
  }

  return (
    <MuiLink component={RouterLink} to={`/${resource}/${id}`} underline="hover" color="primary">
      {label}
    </MuiLink>
  );
};
