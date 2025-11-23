import { useEffect, useState } from "react";
import { useNavigate, type NavigateFunction } from "react-router-dom";
import { Button, Card, CardContent, Dialog, DialogActions, DialogContent, DialogTitle, Paper, Chip, Typography, Stack, TextField } from "@mui/material";
import { DataGrid, GridActionsCellItem, type GridColDef, type GridRowParams } from "@mui/x-data-grid";
import ArticleIcon from "@mui/icons-material/Article";
import DevicesIcon from "@mui/icons-material/Devices";
import EventNoteIcon from "@mui/icons-material/EventNote";
import GroupIcon from "@mui/icons-material/Group";

import { EmptyState, PageHeader } from "../components";
import { useBlockedEvents } from "../hooks/useQueries";
import { useToast } from "../hooks/useToast";
import { formatDateTime } from "../utils/dates";

type EventRow = {
  id: string;
  occurredAt: string | undefined;
  file_path: string;
  payload: Record<string, unknown>;
  hostname: string;
  machineId: string | undefined;
  user: string | undefined;
  userId: string | undefined;
  kind: string;
};

function createEventColumns(navigate: NavigateFunction, showPayload: (payload: Record<string, unknown>) => void): GridColDef<EventRow>[] {
  return [
    {
      field: "occurredAt",
      headerName: "Occurred",
      flex: 1,
      sortable: true,
      renderCell: (params) => formatDateTime(params.row.occurredAt),
    },
    {
      field: "file_path",
      headerName: "Process",
      flex: 1,
    },
    {
      field: "hostname",
      headerName: "Machine",
      flex: 1,
      sortable: false,
      filterable: true,
    },
    {
      field: "user",
      headerName: "User",
      flex: 1,
      sortable: false,
      filterable: true,
    },
    {
      field: "kind",
      headerName: "Result",
      flex: 1,
      sortable: false,
      filterable: true,
      renderCell: (params) => {
        const status = params.value as string;

        let eventStatus: "warning" | "success" | "error" | "info";

        if (status.includes("ALLOW")) {
          eventStatus = status.includes("UNKNOWN") ? "warning" : "success";
        } else if (status.includes("BLOCK")) {
          eventStatus = status.includes("UNKNOWN") ? "warning" : "error";
        } else {
          eventStatus = "info";
        }

        return (
          <Chip
            size="small"
            color={eventStatus}
            variant="filled"
            label={status}
          />
        );
      },
    },
    {
      field: "actions",
      type: "actions",
      getActions: (params: GridRowParams<EventRow>) => {
        const actions = [];

        const userId = params.row.userId;
        if (userId) {
          actions.push(
            <GridActionsCellItem
              key="user"
              showInMenu
              icon={<GroupIcon />}
              label="View User"
              onClick={() => void navigate(`/users/${userId}`)}
            />,
          );
        }

        const machineId = params.row.machineId;
        if (machineId) {
          actions.push(
            <GridActionsCellItem
              key="device"
              showInMenu
              icon={<DevicesIcon />}
              label="View Device"
              onClick={() => void navigate(`/devices/${machineId}`)}
            />,
          );
        }

        actions.push(
          <GridActionsCellItem
            key="details"
            showInMenu
            icon={<ArticleIcon />}
            label="View Payload"
            onClick={() => {
              showPayload(params.row.payload);
            }}
          />,
        );

        return actions;
      },
    },
  ];
}

export default function Events() {
  const navigate = useNavigate();
  const { events, loading, error } = useBlockedEvents();
  const { showToast } = useToast();
  const [eventPayload, setEventPayload] = useState<string | null>(null);

  useEffect(() => {
    if (!error) return;

    const message = error || "Failed to load events";
    showToast({
      message,
      severity: "error",
    });
  }, [error, showToast]);

  const columns = createEventColumns(navigate, (payload) => {
    setEventPayload(JSON.stringify(payload, null, 2));
  });
  const handleClosePayloadDialog = () => {
    setEventPayload(null);
  };

  const rows: EventRow[] = events.map((event) => ({
    id: event.id,
    occurredAt: event.occurredAt,
    file_path: typeof event.payload?.file_name === "string" ? event.payload.file_name : event.kind,
    payload: event.payload ?? {},
    hostname: event.hostname,
    machineId: event.machineId,
    user: event.email,
    userId: event.userId,
    kind: typeof event.payload?.decision === "string" ? event.payload.decision : event.kind,
  }));

  return (
    <>
      <Stack spacing={3}>
        <PageHeader
          title="Events"
          subtitle="Audit log of all Santa agent activity."
        />

        <Paper sx={{ height: 640, width: "100%" }}>
          <DataGrid
            rows={rows}
            columns={columns}
            showToolbar
            loading={loading}
            disableRowSelectionOnClick
            initialState={{
              sorting: {
                sortModel: [{ field: "occurredAt", sort: "desc" }],
              },
            }}
            slots={{
              noRowsOverlay: () => (
                <EmptyState
                  title="No Events Found"
                  description="No events have been logged by Santa yet. Check back later."
                  icon={<EventNoteIcon fontSize="inherit" />}
                />
              ),
            }}
          />
        </Paper>
      </Stack>

      <Dialog
        open={Boolean(eventPayload)}
        onClose={handleClosePayloadDialog}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Event Payload</DialogTitle>
        <DialogContent dividers>
          <Card variant="outlined">
            <CardContent sx={{ overflowX: "auto" }}>
              {eventPayload ? (
                <TextField
                  fullWidth
                  multiline
                  minRows={10}
                  value={eventPayload}
                  slotProps={{ input: { readOnly: true } }}
                />
              ) : (
                <Typography color="text.secondary">No payload data available.</Typography>
              )}
            </CardContent>
          </Card>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClosePayloadDialog}>Close</Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
