import { useEffect, useState, useMemo } from "react";
import { Box, Card, CardContent, CardHeader } from "@mui/material";
import { DataGrid, type GridColDef } from "@mui/x-data-grid";

import { useDevices } from "../hooks/useQueries";
import { formatDateTime } from "../utils/dates";
import { PageSnackbar, type PageToast } from "../components";

export default function Devices() {
  const { data: devices = [], isLoading, error } = useDevices({});

  const [toast, setToast] = useState<PageToast>({
    open: false,
    message: "",
    severity: "error",
  });

  useEffect(() => {
    if (error) {
      setToast({
        open: true,
        message: error instanceof Error ? error.message : "Failed to load devices",
        severity: "error",
      });
    }
  }, [error]);

  const columns = useMemo<GridColDef[]>(
    () => [
      { field: "hostname", headerName: "Hostname", flex: 1 },
      { field: "serial", headerName: "Serial", flex: 1 },
      { field: "primaryUser", headerName: "Primary user", flex: 1 },
      { field: "clientMode", headerName: "Mode", flex: 1 },
      {
        field: "lastSeen",
        headerName: "Last seen",
        flex: 1,
        valueFormatter: (value) => formatDateTime(value as string),
      },
      { field: "machineIdentifier", headerName: "Machine ID", flex: 1 },
    ],
    [],
  );

  const rows = useMemo(
    () =>
      devices.map((device) => ({
        id: device.serial,
        hostname: device.hostname,
        serial: device.serial,
        primaryUser: device.primaryUser,
        clientMode: device.clientMode,
        lastSeen: device.lastSeen,
        machineIdentifier: device.machineIdentifier,
      })),
    [devices],
  );

  return (
    <Card elevation={1}>
      <CardHeader title="Devices" subheader="Monitor Santa agents across your fleet." />
      {/* TODO: add device overview Chips? */}
      <CardContent>
        <Box height={600}>
          <DataGrid
            rows={rows}
            columns={columns}
            showToolbar
            loading={isLoading}
            initialState={{
              sorting: { sortModel: [{ field: "hostname", sort: "asc" }] },
            }}
          />
        </Box>

        <PageSnackbar toast={toast} onClose={() => setToast((prev) => ({ ...prev, open: false }))} />
      </CardContent>
    </Card>
  );
}
