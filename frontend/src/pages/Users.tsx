import { useEffect, useState, useMemo } from "react";
import { Link as RouterLink } from "react-router-dom";
import { Box, Button, Card, CardContent, CardHeader } from "@mui/material";
import { DataGrid, type GridColDef } from "@mui/x-data-grid";

import { useUsers } from "../hooks/useQueries";
import { PageSnackbar, type PageToast } from "../components";

export default function Users() {
  const { data: users = [], isLoading, error } = useUsers({});

  const [toast, setToast] = useState<PageToast>({
    open: false,
    message: "",
    severity: "error",
  });

  useEffect(() => {
    if (error) {
      setToast({
        open: true,
        message: error instanceof Error ? error.message : "Failed to load users",
        severity: "error",
      });
    }
  }, [error]);

  const columns = useMemo<GridColDef[]>(
    () => [
      { field: "displayName", headerName: "Name", flex: 1 },
      { field: "upn", headerName: "UPN", flex: 1 },
      {
        field: "actions",
        headerName: "Actions",
        flex: 1,
        sortable: false,
        filterable: false,
        renderCell: ({ row }) => (
          <Button component={RouterLink} to={`/users/${row.id}`} size="small" variant="outlined">
            View details
          </Button>
        ),
      },
    ],
    [],
  );

  const rows = useMemo(
    () =>
      users.map((user) => ({
        id: user.id,
        displayName: user.displayName,
        upn: user.upn,
      })),
    [users],
  );

  return (
    <Card elevation={1}>
      <CardHeader title="Users" subheader="Manage user access from your Entra ID directory." />
      <CardContent>
        <Box height={600}>
          <DataGrid
            rows={rows}
            columns={columns}
            showToolbar
            loading={isLoading}
            initialState={{
              sorting: { sortModel: [{ field: "displayName", sort: "asc" }] },
            }}
          />
        </Box>

        <PageSnackbar toast={toast} onClose={() => setToast((prev) => ({ ...prev, open: false }))} />
      </CardContent>
    </Card>
  );
}
