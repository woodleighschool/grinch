import { useMemo, useState, useEffect } from "react";
import { useDevices } from "../hooks/useQueries";
import { useDebouncedValue } from "../hooks/useDebouncedValue";
import { formatTimeAgo } from "../utils/dates";
import { Card, CardContent, CardHeader, Chip, InputAdornment, Stack, TextField, Typography, Snackbar, Box } from "@mui/material";
import { DataGrid, type GridColDef } from "@mui/x-data-grid";
import SearchIcon from "@mui/icons-material/Search";

export default function Devices() {
  const [searchTerm, setSearchTerm] = useState("");
  const debouncedSearch = useDebouncedValue(searchTerm, 300);
  const trimmedSearch = searchTerm.trim();
  const hasSearchTerm = trimmedSearch.length > 0;

  const { data: devices = [], isLoading, isFetching, error } = useDevices({ search: debouncedSearch });

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
        field: "primaryUser",
        headerName: "Primary User",
        flex: 1.5,
        minWidth: 200,
        renderCell: (p) => <Typography variant="body2">{p.value || "—"}</Typography>,
      },
      {
        field: "clientMode",
        headerName: "Mode",
        flex: 0.8,
        minWidth: 100,
        renderCell: (p) => (
          <Chip
            size="small"
            label={p.value || "Unknown"}
            color={p.value === "MONITOR" ? "primary" : p.value === "LOCKDOWN" ? "error" : "default"}
            variant="outlined"
          />
        ),
      },
      {
        field: "status",
        headerName: "Status",
        flex: 1,
        minWidth: 100,
        sortable: false,
        valueGetter: (_value, row) => getStatusCategory(row.lastSeen as string | null),
        renderCell: (p) => {
          const status = p.value as Status;
          const colorMap = {
            success: "success" as const,
            warning: "warning" as const,
            danger: "error" as const,
            muted: "default" as const,
          };
          const textMap = {
            success: "Online",
            warning: "Inactive",
            danger: "Offline",
            muted: "Unknown",
          };
          return <Chip size="small" color={colorMap[status]} variant={colorMap[status] === "default" ? "outlined" : "filled"} label={textMap[status]} />;
        },
      },
      {
        field: "lastSeen",
        headerName: "Last Seen",
        flex: 1,
        minWidth: 120,
        valueFormatter: (value) => (value ? formatTimeAgo(value as string) : "Never"),
      },
      {
        field: "lastPreflightAt",
        headerName: "Last Preflight",
        flex: 1,
        minWidth: 130,
        valueFormatter: (value) => (value ? formatTimeAgo(value as string) : "Never"),
      },
      {
        field: "cleanSyncRequested",
        headerName: "Clean Sync",
        flex: 0.8,
        minWidth: 100,
        renderCell: (p) => <Chip size="small" label={p.value ? "Yes" : "No"} color={p.value ? "warning" : "default"} variant="outlined" />,
      },
      {
        field: "machineIdentifier",
        headerName: "Machine ID",
        flex: 1.2,
        minWidth: 200,
        renderCell: (p) => (
          <Typography component="code" variant="body2" fontSize="0.75rem">
            {p.value || "—"}
          </Typography>
        ),
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
        primaryUser: d.primaryUser,
        clientMode: d.clientMode,
        lastSeen: d.lastSeen,
        lastPreflightAt: d.lastPreflightAt,
        cleanSyncRequested: d.cleanSyncRequested,
        machineIdentifier: d.machineIdentifier,
      })),
    [devices],
  );

  return (
    <Card elevation={1}>
      <CardHeader title="Devices" subheader="Monitor and manage Santa agents across your fleet." />
      <CardContent>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
            <Chip size="small" color="success" variant="filled" label={`Online: ${statusCounts.online}`} />
            <Chip size="small" color="warning" variant="filled" label={`Inactive: ${statusCounts.warning}`} />
            <Chip size="small" color="error" variant="filled" label={`Offline: ${statusCounts.offline}`} />
            <Chip label={`Total: ${statusCounts.total}`} size="small" />
          </Stack>

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
                    <SearchIcon />
                  </InputAdornment>
                ),
              },
            }}
          />
        </Stack>

        <Box sx={{ height: 600, width: "100%", mt: 2 }}>
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
        </Box>

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
