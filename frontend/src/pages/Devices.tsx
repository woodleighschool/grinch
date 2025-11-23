import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Paper, Stack } from "@mui/material";
import { DataGrid, type GridColDef, type GridRowParams } from "@mui/x-data-grid";
import DevicesOtherIcon from "@mui/icons-material/DevicesOther";

import { EmptyState, PageHeader } from "../components";
import { useDevices } from "../hooks/useQueries";
import { useToast } from "../hooks/useToast";
import { formatDateTime } from "../utils/dates";

type DeviceRow = {
  id: string;
  hostname: string;
  serial: string;
  primaryUser: string | undefined;
  clientMode: string | undefined;
  lastSeen: string | undefined;
};

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

export default function Devices() {
  const navigate = useNavigate();
  const { data: devices = [], isLoading, error } = useDevices({});
  const { showToast } = useToast();

  useEffect(() => {
    if (!error) return;

    showToast({
      message: error instanceof Error ? error.message : "Failed to load devices",
      severity: "error",
    });
  }, [error, showToast]);

  const rows: DeviceRow[] = devices.map((device) => ({
    id: device.id,
    hostname: device.hostname,
    serial: device.serial,
    primaryUser: device.primaryUser,
    clientMode: device.clientMode,
    lastSeen: device.lastSeen,
  }));

  return (
    <Stack spacing={3}>
      <PageHeader
        title="Devices"
        subtitle="Monitor Santa agents across your fleet."
      />

      <Paper sx={{ height: 640, width: "100%" }}>
        <DataGrid
          rows={rows}
          columns={deviceColumns}
          showToolbar
          loading={isLoading}
          disableRowSelectionOnClick
          onRowClick={(params: GridRowParams<DeviceRow>) => {
            void navigate(`/devices/${String(params.id)}`);
          }}
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
