import { useMemo, useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useBlockedEvents } from "../hooks/useQueries";
import { formatDateTime } from "../utils/dates";
import { Badge, Box, Card, CardContent, CardHeader, Chip, Divider, Stack, Tooltip, Typography } from "@mui/material";
import { DataGrid, GridActionsCellItem, type GridRowParams, type GridColDef, type GridActionsCellItemProps } from "@mui/x-data-grid";
import { PageSnackbar, type PageToast } from "../components";
import GroupIcon from "@mui/icons-material/Group";
import DevicesIcon from "@mui/icons-material/Devices";

export default function Events() {
	const navigate = useNavigate();
  const { events, loading, error } = useBlockedEvents();

  const [toast, setToast] = useState<PageToast>({
    open: false,
    message: "",
    severity: "error",
  });

  useEffect(() => {
    if (error) {
      setToast({
        open: true,
        message: "Failed to load events",
        severity: "error",
      });
    }
  }, [error]);

  const handleToastClose = () => setToast((prev) => ({ ...prev, open: false }));

  const columns = useMemo<GridColDef[]>(
    () => [
      { field: "occurredAt", headerName: "Occurred", flex: 1, sortable: true },
      { field: "file_path", headerName: "Process", flex: 1 },
      { field: "hostname", headerName: "Machine", flex: 1, sortable: false, filterable: true },
      { field: "userDisplayName", headerName: "User", flex: 1, sortable: false, filterable: true },
      {
        field: "kind",
        headerName: "Result",
        flex: 1,
        sortable: false,
        filterable: true,
        renderCell: (p) => {
          var eventStatus: "warning" | "success" | "error" | "info";
          var status = p.value as string;
          if (status.includes("ALLOW")) {
            eventStatus = status.includes("UNKNOWN") ? "warning" : "success";
          } else if (status.includes("BLOCK")) {
            eventStatus = status.includes("UNKNOWN") ? "warning" : "error";
          } else {
            eventStatus = "info";
          }
          return <Chip size="small" color={eventStatus} variant="filled" label={status} />;
        },
      },
      {
        field: "actions",
        type: "actions",
        getActions: (params: GridRowParams) => [
          <GridActionsCellItem showInMenu icon={<GroupIcon />} label="Go to user" onClick={() => navigate(`/users/${params.row.userId}`)} />,
          <GridActionsCellItem showInMenu icon={<DevicesIcon />} label="Go to device" onClick={() => navigate(`/devices/${params.row.machineId}`)} />,
        ],
      },
    ],
    [],
  );

  const rows = useMemo(
    () =>
      events.map((event) => ({
        id: event.id,
        occurredAt: event.occurredAt ? formatDateTime(event.occurredAt) : "-",
        file_path: typeof event.payload?.file_name == "string" ? event.payload.file_name : event.kind,
        hostname: event.hostname,
		machineId: event.machineId,
        userDisplayName: event.userDisplayName || "-",
		userId: event.userId || "-",
        kind: event.kind,
      })),
    [events],
  );

  return (
    <Card elevation={1}>
      <CardHeader title="Events" subheader="All events logged by Santa" />
      <CardContent>
        <Box style={{ height: "100%", width: "100%", marginTop: 16 }}>
          <DataGrid
            rows={rows}
            columns={columns}
            pageSizeOptions={[25, 50, 100]}
            showToolbar
            initialState={{
              pagination: { paginationModel: { pageSize: 100, page: 0 } },
              sorting: { sortModel: [{ field: "occurredAt", sort: "desc" }] },
            }}
            loading={loading}
            slots={{
              noRowsOverlay: () => (
                <Box textAlign="center" padding={3}>
                  <Stack spacing={1}>
                    <Typography variant="h6">No events found</Typography>
                    <Typography color="text.secondary">No events have been logged by Santa yet, check back later...</Typography>
                  </Stack>
                </Box>
              ),
            }}
          />
        </Box>
        <PageSnackbar toast={toast} onClose={handleToastClose} />
      </CardContent>
    </Card>
  );
}
