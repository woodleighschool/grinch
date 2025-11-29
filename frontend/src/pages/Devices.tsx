import { useEffect, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Alert, Paper, Stack } from "@mui/material";
import { DataGrid, type GridColDef, type GridRowParams } from "@mui/x-data-grid";
import DevicesOtherIcon from "@mui/icons-material/DevicesOther";

import { EmptyState, PageHeader } from "../components";
import { useDevices } from "../hooks/useQueries";
import { useToast } from "../hooks/useToast";
import { formatDateTime } from "../utils/dates";

// Types
type DeviceRow = {
  id: string;
  hostname: string;
  serial: string;
  primaryUser: string | undefined;
  clientMode: string | undefined;
  lastSeen: string | undefined;
};

// Data grid configuration
const deviceColumns: GridColDef<DeviceRow>[] = [
  {
    field: "hostname",
    headerName: "Hostname",
    flex: 1,
  },
  {
    field: "serial",
    headerName: "Serial",
    flex: 1,
  },
  {
    field: "primaryUser",
    headerName: "Primary user",
    flex: 1,
  },
  {
    field: "clientMode",
    headerName: "Mode",
    flex: 1,
  },
  {
    field: "lastSeen",
    headerName: "Last seen",
    flex: 1,
    valueFormatter: (value) => formatDateTime(value as string),
  },
];

// Page component
export default function Devices() {
  // Hooks
  const navigate = useNavigate();
  const { data: devices = [], isLoading, error } = useDevices();
  const { showToast } = useToast();

  // Effects
  useEffect(() => {
    if (!error) return;

    showToast({
      message: error instanceof Error ? error.message : "Failed to load devices",
      severity: "error",
    });
  }, [error, showToast]);

  // Handlers
  const handleRowClick = useCallback(
    (params: GridRowParams<DeviceRow>) => {
      void navigate(`/devices/${String(params.id)}`);
    },
    [navigate],
  );

  const devicesErrorMessage = useMemo(() => {
    if (!error) return null;
    return error instanceof Error ? error.message : "Failed to load devices";
  }, [error]);

  // Derived data
  const rows: DeviceRow[] = useMemo(
    () =>
      devices.map((device) => ({
        id: device.id,
        hostname: device.hostname,
        serial: device.serial,
        primaryUser: device.primaryUser,
        clientMode: device.clientMode,
        lastSeen: device.lastSeen,
      })),
    [devices],
  );

  // Render
  return (
    <Stack spacing={3}>
      <PageHeader
        title="Devices"
        subtitle="Monitor Santa agents across your fleet."
      />

      {devicesErrorMessage && <Alert severity="error">{devicesErrorMessage}</Alert>}

      <Paper sx={{ height: 640, width: "100%" }}>
        <DataGrid
          rows={rows}
          columns={deviceColumns}
          showToolbar
          loading={isLoading}
          disableRowSelectionOnClick
          onRowClick={handleRowClick}
          initialState={{
            sorting: {
              sortModel: [{ field: "hostname", sort: "asc" }],
            },
          }}
          slots={{
            noRowsOverlay: () => (
              <EmptyState
                title="No Devices Found"
                description="No devices have checked in yet. Ensure your MDM profile is deployed."
                icon={<DevicesOtherIcon fontSize="inherit" />}
              />
            ),
          }}
        />
      </Paper>
    </Stack>
  );
}
