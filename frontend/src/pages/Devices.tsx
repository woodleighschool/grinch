import { useMemo, useState, useEffect } from "react";
import { useDevices } from "../hooks/useQueries";
import { useDebouncedValue } from "../hooks/useDebouncedValue";
import { formatTimeAgo } from "../utils/dates";
import { Alert, Button, Card, CardContent, CardHeader, Chip, CircularProgress, InputAdornment, Stack, TextField, Typography, Snackbar } from "@mui/material";
import { DataGrid, type GridColDef } from "@mui/x-data-grid";
import SearchIcon from "@mui/icons-material/Search";

export default function Devices() {
  const [searchTerm, setSearchTerm] = useState("");
  const debouncedSearch = useDebouncedValue(searchTerm, 300);
  const trimmedSearch = searchTerm.trim();
  const hasSearchTerm = trimmedSearch.length > 0;

  const { data: devices = [], isLoading, isFetching, error, refetch } = useDevices({ search: debouncedSearch });

  const [toast, setToast] = useState<{ open: boolean; message: string }>({ open: false, message: "" });

  useEffect(() => {
    if (error) {
      console.error("Failed to load devices", error);
      setToast({ open: true, message: error instanceof Error ? error.message : "Failed to load devices" });
    }
  }, [error]);

  type Status = "success" | "warning" | "danger" | "muted";
  function getStatusCategory(lastSeen?: string | null): Status {
    if (!lastSeen) return "muted";
    const diffMinutes = (Date.now() - new Date(lastSeen).getTime()) / (1000 * 60);
    if (diffMinutes <= 15) return "success";
    if (diffMinutes <= 120) return "warning";
    return "danger";
  }

  const statusCounts = useMemo(() => {
    let online = 0;
    let warning = 0;
    let offline = 0;
    devices.forEach((d) => {
      const c = getStatusCategory(d.lastSeen);
      if (c === "success") online += 1;
      else if (c === "warning") warning += 1;
      else if (c === "danger") offline += 1;
    });
    return { online, warning, offline, total: devices.length };
  }, [devices]);

  const statusChip = (cat: Status, label?: string) => {
    const map: Record<Status, { color: "success" | "warning" | "error" | "default"; text: string }> = {
      success: { color: "success", text: "Online" },
      warning: { color: "warning", text: "Inactive" },
      danger: { color: "error", text: "Offline" },
      muted: { color: "default", text: "Unknown" },
    };
    const { color, text } = map[cat];
    return <Chip size="small" color={color} variant={color === "default" ? "outlined" : "filled"} label={label ?? text} />;
  };

  const columns = useMemo<GridColDef[]>(
    () => [
      {
        field: "hostname",
        headerName: "Hostname",
        flex: 1.2,
        minWidth: 160,
        renderCell: (params) => (
          <Typography variant="body2" fontWeight={600}>
            {params.value || "—"}
          </Typography>
        ),
      },
      {
        field: "serial",
        headerName: "Serial",
        flex: 1,
        minWidth: 140,
        renderCell: (p) => <Typography component="code">{p.value || "—"}</Typography>,
      },
      {
        field: "status",
        headerName: "Status",
        flex: 1,
        minWidth: 140,
        sortable: false,
        valueGetter: (_value, row) => getStatusCategory(row.lastSeen as string | null),
        renderCell: (p) => statusChip(p.value as Status),
      },
      {
        field: "ruleCursor",
        headerName: "Rule Cursor",
        flex: 1,
        minWidth: 140,
        renderCell: (p) => <Typography variant="body2">{p.value || "—"}</Typography>,
      },
      {
        field: "syncCursor",
        headerName: "Sync Cursor",
        flex: 1,
        minWidth: 140,
        renderCell: (p) => <Typography variant="body2">{p.value || "—"}</Typography>,
      },
      {
        field: "lastSeen",
        headerName: "Last Seen",
        flex: 0.8,
        minWidth: 120,
        valueFormatter: (value) => (value ? formatTimeAgo(value as string) : "Never"),
      },
    ],
    [],
  );

  const rows = useMemo(
    () =>
      devices.map((d) => ({
        id: d.serial || d.hostname || crypto.randomUUID(),
        hostname: d.hostname,
        serial: d.serial,
        lastSeen: d.lastSeen,
        ruleCursor: d.ruleCursor,
        syncCursor: d.syncCursor,
      })),
    [devices],
  );

  return (
    <Card elevation={1}>
      <CardHeader title="Devices" subheader="Monitor and manage Santa agents across your fleet." />
      <CardContent>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
            {statusChip("success", `Online: ${statusCounts.online}`)}
            {statusChip("warning", `Inactive: ${statusCounts.warning}`)}
            {statusChip("danger", `Offline: ${statusCounts.offline}`)}
            <Chip label={`Total: ${statusCounts.total}`} size="small" />
            {isFetching && <CircularProgress size={16} />}
          </Stack>

          <Stack direction="row" spacing={1} alignItems="center">
            <Button size="small" variant="outlined" onClick={() => void refetch()}>
              Refresh
            </Button>
            <TextField
              type="search"
              size="small"
              label="Search devices..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              slotProps={{
                input: {
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchIcon fontSize="small" />
                    </InputAdornment>
                  ),
                },
              }}
              inputProps={{ "aria-label": "Search devices" }}
            />
          </Stack>
        </Stack>

        <div style={{ height: 600, width: "100%", marginTop: 16 }}>
          <DataGrid
            rows={rows}
            columns={columns}
            disableColumnMenu
            pageSizeOptions={[25, 50, 100]}
            initialState={{
              pagination: { paginationModel: { pageSize: 100, page: 0 } },
              sorting: { sortModel: [{ field: "hostname", sort: "asc" }] },
            }}
            loading={isLoading || isFetching}
            slots={{
              noRowsOverlay: () => (
                <Stack alignItems="center" spacing={0.5} p={3}>
                  <Typography variant="h6">No devices found</Typography>
                  <Typography color="text.secondary">
                    {hasSearchTerm ? `No devices match "${trimmedSearch}".` : "No devices are available in the fleet."}
                  </Typography>
                </Stack>
              ),
            }}
          />
        </div>

        <Snackbar
          open={toast.open}
          autoHideDuration={4000}
          onClose={() => setToast((t) => ({ ...t, open: false }))}
          anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
          message={toast.message}
        />
      </CardContent>
    </Card>
  );
}
