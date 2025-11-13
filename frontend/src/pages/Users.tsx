import { useMemo, useState, useEffect } from "react";
import { Link as RouterLink } from "react-router-dom";
import { useUsers } from "../hooks/useQueries";
import { useDebouncedValue } from "../hooks/useDebouncedValue";
import { Box, Card, CardContent, CardHeader, Chip, InputAdornment, Stack, TextField, Typography, Button } from "@mui/material";
import { DataGrid, type GridColDef } from "@mui/x-data-grid";
import SearchIcon from "@mui/icons-material/Search";
import PersonIcon from "@mui/icons-material/Person";
import { PageSnackbar, type PageToast } from "../components";

export default function Users() {
  const [searchTerm, setSearchTerm] = useState("");
  const debouncedSearch = useDebouncedValue(searchTerm, 300);
  const trimmedSearch = searchTerm.trim();
  const hasSearchTerm = trimmedSearch.length > 0;

  const { data: users = [], isLoading, isFetching, error } = useUsers({ search: debouncedSearch });

  const [toast, setToast] = useState<PageToast>({ open: false, message: "", severity: "error" });
  useEffect(() => {
    if (error) {
      console.error("Failed to load users", error);
      setToast({ open: true, message: error instanceof Error ? error.message : "Failed to load users", severity: "error" });
    }
  }, [error]);
  const handleToastClose = () => setToast((prev) => ({ ...prev, open: false }));

  const columns = useMemo<GridColDef[]>(
    () => [
      {
        field: "displayName",
        headerName: "Name",
        flex: 1.2,
        minWidth: 200,
        sortable: true,
        renderCell: ({ row }) => (
          <Stack direction="row" spacing={1} alignItems="center">
            <PersonIcon />
            <Typography variant="body2" fontWeight={600}>
              {row.displayName}
            </Typography>
          </Stack>
        ),
      },
      {
        field: "upn",
        headerName: "UPN",
        flex: 1,
        minWidth: 200,
        renderCell: (p) => (
          <Typography component="code" variant="body2">
            {p.value}
          </Typography>
        ),
      },
      {
        field: "actions",
        headerName: "Actions",
        flex: 0.8,
        minWidth: 140,
        sortable: false,
        filterable: false,
        renderCell: ({ row }) => (
          <Button component={RouterLink} to={`/users/${row.id}`} size="small" variant="outlined">
            View Details
          </Button>
        ),
      },
    ],
    [],
  );

  const rows = useMemo(
    () =>
      users.map((u) => ({
        id: u.id,
        displayName: u.displayName,
        upn: u.upn,
      })),
    [users],
  );

  return (
    <Card elevation={1}>
      <CardHeader title="Users" subheader="Manage user access from your Entra ID directory." />
      <CardContent>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack direction="row" spacing={1} alignItems="center">
            <Chip label={`Total: ${users.length}`} size="small" />
          </Stack>

          <TextField
            type="search"
            size="small"
            label="Search users..."
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

        <Box style={{ height: 600, width: "100%", marginTop: 16 }}>
          <DataGrid
            rows={rows}
            columns={columns}
            disableColumnMenu
            pageSizeOptions={[25, 50, 100]}
            initialState={{
              pagination: { paginationModel: { pageSize: 100, page: 0 } },
              sorting: { sortModel: [{ field: "displayName", sort: "asc" }] },
            }}
            loading={isLoading || isFetching}
            slots={{
              noRowsOverlay: () => (
                <Box textAlign="center" padding={3}>
                  <Stack spacing={1}>
                    <Typography variant="h6">No users found</Typography>
                    <Typography color="text.secondary">
                      {hasSearchTerm ? `No users match "${trimmedSearch}".` : "No users are available in the directory."}
                    </Typography>
                  </Stack>
                </Box>
              ),
            }}
          />
        </Box>
      </CardContent>

      <PageSnackbar toast={toast} onClose={handleToastClose} />
    </Card>
  );
}
