import { Chip, Divider, Stack, Typography, type StackProps } from "@mui/material";

import type { DirectoryUser } from "../api";
import { formatDateTime } from "../utils/dates";

export interface UserSummaryProps extends StackProps {
  user: DirectoryUser;
  showMetadata?: boolean;
}

export function UserSummary({ user, showMetadata = true, ...stackProps }: UserSummaryProps) {
  const hasMetadata = Boolean(showMetadata && (user.createdAt || user.updatedAt));

  return (
    <Stack
      spacing={1.5}
      {...stackProps}
    >
      <Typography
        variant="h5"
        fontWeight={600}
      >
        {user.displayName}
      </Typography>
      <Typography
        variant="body2"
        color="text.secondary"
      >
        {user.upn}
      </Typography>

      {hasMetadata && <Divider />}

      {hasMetadata && (
        <Stack
          direction="row"
          spacing={1}
          flexWrap="wrap"
        >
          {user.createdAt && (
            <Chip
              size="small"
              variant="outlined"
              label={`Created ${formatDateTime(user.createdAt)}`}
            />
          )}
          {user.updatedAt && (
            <Chip
              size="small"
              variant="outlined"
              label={`Updated ${formatDateTime(user.updatedAt)}`}
            />
          )}
        </Stack>
      )}
    </Stack>
  );
}
