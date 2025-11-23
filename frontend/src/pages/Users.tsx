import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Paper, Stack } from "@mui/material";
import { DataGrid, type GridColDef, type GridRowParams } from "@mui/x-data-grid";
import GroupOffIcon from "@mui/icons-material/GroupOff";

import { PageHeader, EmptyState } from "../components";
import { useToast } from "../hooks/useToast";
import { useUsers } from "../hooks/useQueries";

type UserRow = {
  id: string;
  displayName: string;
  upn: string;
};

const columns: GridColDef<UserRow>[] = [
  { field: "displayName", headerName: "Name", flex: 1 },
  { field: "upn", headerName: "UPN", flex: 1 },
];

export default function Users() {
  const navigate = useNavigate();
  const { data: users = [], isLoading, error } = useUsers({});
  const { showToast } = useToast();

  useEffect(() => {
    if (!error) return;

    showToast({
      message: error instanceof Error ? error.message : "Failed to load users",
      severity: "error",
    });
  }, [error, showToast]);

  const handleRowClick = (params: GridRowParams<UserRow>) => {
    void navigate(`/users/${String(params.id)}`);
  };

  const rows: UserRow[] = users.map((user) => ({
    id: user.id,
    displayName: user.displayName,
    upn: user.upn,
  }));

  return (
    <Stack spacing={3}>
      <PageHeader
        title="Users"
        subtitle="Manage user access from your Entra ID directory."
      />

      <Paper sx={{ height: 640 }}>
        <DataGrid
          rows={rows}
          columns={columns}
          showToolbar
          loading={isLoading}
          disableRowSelectionOnClick
          onRowClick={handleRowClick}
          initialState={{
            sorting: {
              sortModel: [{ field: "displayName", sort: "asc" }],
            },
          }}
          slots={{
            noRowsOverlay: () => (
              <EmptyState
                title="No Users Found"
                description="No users are available yet. Try again after syncing with your directory."
                icon={<GroupOffIcon fontSize="inherit" />}
              />
            ),
          }}
        />
      </Paper>
    </Stack>
  );
}
