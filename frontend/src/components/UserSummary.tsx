import { Chip, Divider, Stack, Typography, type StackProps } from "@mui/material";

import type { DirectoryUser } from "../api";
import { formatDateTime } from "../utils/dates";

export interface UserSummaryProps extends StackProps {
  user: DirectoryUser;
}

export function UserSummary({ user, ...stackProps }: UserSummaryProps) {
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

      <Divider />

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
    </Stack>
  );
}
