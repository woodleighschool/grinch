import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button, Chip, Paper, Stack, Typography } from "@mui/material";
import { useConfirm } from "material-ui-confirm";
import { DataGrid, GridActionsCellItem, type GridColDef, type GridRowParams } from "@mui/x-data-grid";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import ShieldIcon from "@mui/icons-material/Shield";

import type { Application } from "../api";
import { useApplications, useDeleteApplication } from "../hooks/useQueries";
import { ApplicationDialog, EmptyState, PageHeader } from "../components";
import { useToast } from "../hooks/useToast";

type DialogMode = "create" | "edit";

interface DialogConfig {
  mode: DialogMode;
  application?: Application | null;
}

export default function Applications() {
  const navigate = useNavigate();

  const { data: apps = [], error: appsError, isLoading } = useApplications();

  const deleteApplication = useDeleteApplication();

  const [dialogConfig, setDialogConfig] = useState<DialogConfig | null>(null);
  const [deletingAppId, setDeletingAppId] = useState<string | null>(null);
  const confirm = useConfirm();

  const { showToast } = useToast();
  const dialogMode: DialogMode = dialogConfig?.mode ?? "create";

  const openCreateDialog = () => {
    setDialogConfig({ mode: "create", application: null });
  };

  const openEditDialog = (application: Application) => {
    setDialogConfig({ mode: "edit", application });
  };

  const closeDialog = () => {
    setDialogConfig(null);
  };

  useEffect(() => {
    if (!appsError) return;

    console.error("Applications query failed", appsError);
    showToast({
      message: appsError instanceof Error ? appsError.message : "Failed to load applications.",
      severity: "error",
    });
  }, [appsError, showToast]);

  const handleDeleteApplication = async (appId: string) => {
    setDeletingAppId(appId);

    try {
      await deleteApplication.mutateAsync(appId);
      showToast({
        message: "Application rule deleted.",
        severity: "success",
      });
    } catch (error) {
      console.error("Delete application failed", error);
      showToast({
        message: "Failed to delete application rule.",
        severity: "error",
      });
    } finally {
      setDeletingAppId(null);
    }
  };

  const handleRowClick = (params: GridRowParams<Application>) => {
    void navigate(`/applications/${String(params.id)}`);
  };

  const handleConfirmDelete = (appId: string, appName: string) => {
    confirm({
      title: "Delete Application Rule",
      description: `Are you sure you want to delete “${appName}”? This action cannot be undone.`,
      cancellationText: "Cancel",
      confirmationText: "Delete",
      confirmationButtonProps: { color: "error", variant: "contained" },
    })
      .then(() => void handleDeleteApplication(appId))
      .catch(() => {});
  };

  const columns = createApplicationColumns({
    onEdit: openEditDialog,
    onRequestDelete: handleConfirmDelete,
    deletingAppId,
  });

  const handleDialogSuccess = () => {
    const message = dialogMode === "edit" ? "Application rule updated." : "Application rule created.";

    showToast({ message, severity: "success" });
  };

  const action = (
    <Button
      variant="contained"
      startIcon={<AddIcon />}
      onClick={openCreateDialog}
    >
      Add Application
    </Button>
  );

  return (
    <>
      <Stack spacing={3}>
        <PageHeader
          title="Applications"
          subtitle="Define rules using reference-compatible identifiers."
          action={action}
        />

        <Paper sx={{ height: 640, width: "100%" }}>
          <DataGrid
            rows={apps}
            columns={columns}
            loading={isLoading}
            disableRowSelectionOnClick={true}
            onRowClick={handleRowClick}
            sx={{ "& .MuiDataGrid-row": { cursor: "pointer" } }}
            showToolbar
            slots={{
              noRowsOverlay: () => (
                <EmptyState
                  title="No application rules yet"
                  description="Create your first application rule to get started."
                  icon={<ShieldIcon fontSize="inherit" />}
                />
              ),
            }}
            initialState={{
              sorting: {
                sortModel: [{ field: "name", sort: "asc" }],
              },
            }}
          />
        </Paper>
      </Stack>

      <ApplicationDialog
        open={Boolean(dialogConfig)}
        mode={dialogMode}
        application={dialogConfig?.application ?? null}
        onClose={closeDialog}
        onSuccess={handleDialogSuccess}
        onError={(message) => {
          showToast({ message, severity: "error" });
        }}
      />
    </>
  );
}

interface ApplicationColumnOptions {
  onEdit: (application: Application) => void;
  onRequestDelete: (appId: string, appName: string) => void;
  deletingAppId: string | null;
}

function createApplicationColumns({ onEdit, onRequestDelete, deletingAppId }: ApplicationColumnOptions): GridColDef<Application>[] {
  return [
    {
      field: "name",
      headerName: "Name",
      flex: 1,
      renderCell: (params) => (
        <Stack
          direction="row"
          spacing={1}
          alignItems="center"
          height="100%"
        >
          <ShieldIcon
            fontSize="small"
            color={params.row.enabled ? "primary" : "disabled"}
          />
          <Typography
            variant="body2"
            fontWeight="medium"
          >
            {params.value}
          </Typography>
        </Stack>
      ),
    },
    {
      field: "rule_type",
      headerName: "Type",
      width: 120,
    },
    {
      field: "identifier",
      headerName: "Identifier",
      flex: 1,
    },
    {
      field: "assignment_stats",
      headerName: "Assignments",
      width: 250,
      renderCell: (params) => {
        const stats = params.row.assignment_stats;

        return (
          <Stack
            direction="row"
            spacing={1}
            alignItems="center"
            height="100%"
          >
            <Chip
              size="small"
              color="success"
              label={`Allow ${String(stats?.allow_scopes ?? 0)}`}
            />
            <Chip
              size="small"
              color="error"
              label={`Block ${String(stats?.block_scopes ?? 0)}`}
            />
            <Chip
              size="small"
              variant="outlined"
              label={`Total ${String(stats?.total_users ?? 0)}`}
            />
          </Stack>
        );
      },
    },
    {
      field: "actions",
      type: "actions",
      width: 150,
      getActions: (params) => [
        <GridActionsCellItem
          key="edit"
          icon={<EditIcon />}
          label="Edit"
          onClick={(event) => {
            event.stopPropagation();
            onEdit(params.row);
          }}
        />,
        <GridActionsCellItem
          key="delete"
          icon={<DeleteIcon color="error" />}
          label="Delete"
          disabled={deletingAppId === String(params.id)}
          onClick={(event) => {
            event.stopPropagation();
            onRequestDelete(String(params.id), params.row.name);
          }}
          showInMenu
        />,
      ],
    },
  ];
}
