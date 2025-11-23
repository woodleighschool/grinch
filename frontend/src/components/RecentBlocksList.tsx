import { Alert, Card, CardContent, Chip, Stack, Typography } from "@mui/material";

import type { EventRecord } from "../api";
import { formatDateTime } from "../utils/dates";

export interface RecentBlocksListProps {
  events: EventRecord[];
  emptyMessage?: string;
}

export function RecentBlocksList({ events, emptyMessage = "No recent Santa blocks recorded." }: RecentBlocksListProps) {
  if (events.length === 0) {
    return <Alert severity="info">{emptyMessage}</Alert>;
  }

  return (
    <Stack spacing={1.5}>
      {events.map((event) => {
        const occurredAt = event.occurredAt ? formatDateTime(event.occurredAt) : "—";
        const processPath = typeof event.payload?.file_name === "string" ? event.payload.file_name : event.kind;

        return (
          <Card
            key={event.id}
            elevation={0}
            variant="outlined"
          >
            <CardContent>
              <Stack spacing={0.75}>
                <Stack
                  direction="row"
                  spacing={1}
                  alignItems="center"
                  flexWrap="wrap"
                >
                  <Typography fontWeight={600}>{processPath}</Typography>
                  {event.kind && (
                    <Chip
                      size="small"
                      label={event.kind}
                      color="error"
                    />
                  )}
                </Stack>
                <Typography
                  variant="body2"
                  color="text.secondary"
                >
                  Host: {event.hostname || "—"} · {occurredAt}
                </Typography>
              </Stack>
            </CardContent>
          </Card>
        );
      })}
    </Stack>
  );
}
