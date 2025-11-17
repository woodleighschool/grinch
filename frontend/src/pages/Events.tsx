import { useMemo, useState, useEffect, type MouseEvent, type ReactElement } from "react";
import { useNavigate } from "react-router";
import { useBlockedEvents } from "../hooks/useQueries";
import { type EventRecord } from "../api";
import { formatDateTime } from "../utils/dates";
import { Badge, Box, Card, CardContent, CardHeader, Chip, Divider, Popover, Stack, Tooltip, Typography } from "@mui/material";
import { DataGrid, GridActionsCellItem, type GridRowParams, type GridColDef, type GridActionsCellItemProps } from "@mui/x-data-grid";
import { PageSnackbar, type PageToast } from "../components";
import GroupIcon from "@mui/icons-material/Group";
import DevicesIcon from "@mui/icons-material/Devices";
import AppsIcon from "@mui/icons-material/Apps";
import ArticleIcon from "@mui/icons-material/Article";

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

  const [eventPayload, setEventPayload] = useState<String>("");

  const handlePopoverClose = () => {
    setEventPayload("");
  };

  const columns = useMemo<GridColDef[]>(
    () => [
      { field: "occurredAt", headerName: "Occurred", flex: 1, sortable: true },
      { field: "file_path", headerName: "Process", flex: 1 },
      { field: "hostname", headerName: "Machine", flex: 1, sortable: false, filterable: true },
      { field: "user", headerName: "User", flex: 1, sortable: false, filterable: true },
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
        getActions: (params: GridRowParams) => {
          var actions: ReactElement[] = [];
          if (params.row.userId != "") {
            actions.push(<GridActionsCellItem showInMenu icon={<GroupIcon />} label="Go to user" onClick={() => navigate(`/users/${params.row.userId}`)} />);
          }
          if (params.row.machineId != "") {
            actions.push(
              <GridActionsCellItem showInMenu icon={<DevicesIcon />} label="Go to device" onClick={() => navigate(`/devices/${params.row.machineId}`)} />,
            );
          }
          actions.push(
            <GridActionsCellItem
              icon={<ArticleIcon />}
              label="Show details"
              onClick={() => {setEventPayload(params.row.payload)}}
            />,
          );
          return actions;
        },
      },
    ],
    [],
  );

  const rows = useMemo(
    () =>
      events.map((event) => ({
        id: event.id,
        occurredAt: event.occurredAt ? formatDateTime(event.occurredAt) : null,
        file_path: typeof event.payload?.file_name == "string" ? event.payload.file_name : event.kind,
        payload: JSON.stringify(event.payload, null, "\t"),
        hostname: event.hostname,
        machineId: event.machineId,
        user: event.email,
        userId: event.userId,
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
      <Popover
        open={eventPayload != ""}
        anchorPosition={{ top: parent.innerHeight/2 ,left: parent.innerWidth/2 }}
        onClose={handlePopoverClose}
        anchorOrigin={{
          vertical: "bottom",
          horizontal: "left",
        }}
      >
        <Typography>
          <pre>{eventPayload}</pre>
        </Typography>
      </Popover>
    </Card>
  );
}
