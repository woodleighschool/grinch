import { useMemo, useState, useEffect } from "react";
import { Link as RouterLink } from "react-router-dom";
import type { DirectoryUser } from "../api";
import { useUsers } from "../hooks/useQueries";
import { useDebouncedValue } from "../hooks/useDebouncedValue";
import { Card, CardContent, CardHeader, Chip, InputAdornment, Stack, TextField, Typography, Button, Box, Snackbar, Alert } from "@mui/material";
import { DataGrid, type GridColDef } from "@mui/x-data-grid";
import SearchIcon from "@mui/icons-material/Search";
import PersonIcon from "@mui/icons-material/Person";

function getDisplayName(user: DirectoryUser): string {
  return user.displayName || user.upn;
}

export default function Users() {
  const [searchTerm, setSearchTerm] = useState("");
  const debouncedSearch = useDebouncedValue(searchTerm, 300);
  const trimmedSearch = searchTerm.trim();
  const hasSearchTerm = trimmedSearch.length > 0;

  const { data: users = [], isLoading, isFetching, error } = useUsers({ search: debouncedSearch });

  const [toastOpen, setToastOpen] = useState(false);
  useEffect(() => {
    if (error) {
      console.error("Failed to load users", error);
      setToastOpen(true);
    }
  }, [error]);

  const columns = useMemo<GridColDef[]>(
    () => [
      {
        field: "displayName",
        headerName: "Name",
        flex: 1.2,
        minWidth: 200,
        sortable: true,
        renderCell: (p) => (
          <Stack direction="row" spacing={1} alignItems="center">
            <PersonIcon />
            <Typography variant="body2" fontWeight={600}>
              {getDisplayName(p.row as DirectoryUser)}
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
          <Typography component="code" sx={{ fontSize: 13 }}>
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
        renderCell: (p) => (
          <Button component={RouterLink} to={`/users/${(p.row as DirectoryUser).id}`} size="small" variant="outlined">
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

        <Box sx={{ height: 600, width: "100%", mt: 2 }}>
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
                <Box sx={{ textAlign: "center", p: 3 }}>
                  <Typography variant="h6" gutterBottom>
                    No users found
                  </Typography>
                  <Typography color="text.secondary">
                    {hasSearchTerm ? `No users match "${trimmedSearch}".` : "No users are available in the directory."}
                  </Typography>
                </Box>
              ),
            }}
          />
        </Box>
      </CardContent>

      <Snackbar open={toastOpen} autoHideDuration={4000} onClose={() => setToastOpen(false)} anchorOrigin={{ vertical: "bottom", horizontal: "center" }}>
        <Alert severity="error" onClose={() => setToastOpen(false)} variant="filled">
          {error instanceof Error ? error.message : "Failed to load users"}
        </Alert>
      </Snackbar>
    </Card>
  );
}
