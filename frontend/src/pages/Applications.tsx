import { useEffect, useState, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Alert, Box, Button, LinearProgress, Paper, Stack, Typography } from "@mui/material";
import { useConfirm } from "material-ui-confirm";
import { DataGrid, GridActionsCellItem, type GridColDef, type GridRowParams } from "@mui/x-data-grid";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import ShieldIcon from "@mui/icons-material/Shield";
import PauseCircleOutlineIcon from "@mui/icons-material/PauseCircleOutline";
import PlayCircleOutlineIcon from "@mui/icons-material/PlayCircleOutline";

import type { Application } from "../api";
import { useApplications, useDeleteApplication, useUpdateApplication } from "../hooks/useQueries";
import { ApplicationDialog, EmptyState, PageHeader } from "../components";
import { useToast } from "../hooks/useToast";

// Types
type DialogMode = "create" | "edit";

interface DialogConfig {
  mode: DialogMode;
  application?: Application | null;
}

interface ApplicationColumnOptions {
  onEdit: (application: Application) => void;
  onRequestDelete: (appId: string, appName: string) => void;
  onToggleStatus: (application: Application) => void;
  deletingAppId: string | null;
  togglingAppId: string | null;
}

function LinearProgressWithLabel({ value }: { value: number }) {
  const rounded = Math.round(value).toString().concat("%");

  return (
    <Box
      sx={{
        display: "flex",
        alignItems: "center",
        width: "100%",
        gap: 1,
      }}
    >
      <Box sx={{ flexGrow: 1 }}>
        <LinearProgress
          variant="determinate"
          value={value}
        />
      </Box>
      <Typography
        variant="body2"
        sx={{ minWidth: 40, color: "text.secondary" }}
      >
        {rounded}
      </Typography>
    </Box>
  );
}

// Column factory
function createApplicationColumns({
  onEdit,
  onRequestDelete,
  onToggleStatus,
  deletingAppId,
  togglingAppId,
}: ApplicationColumnOptions): GridColDef<Application>[] {
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
      flex: 1,
    },
    {
      field: "identifier",
      headerName: "Identifier",
      flex: 1,
    },
    {
      field: "assignment_stats",
      headerName: "Machine Coverage",
      flex: 1,
      renderCell: (params) => {
        const stats = params.row.assignment_stats;
        const totalMachines = stats?.total_machines ?? 0;
        const syncedMachines = stats?.synced_machines ?? 0;

        if (totalMachines === 0) {
          return <Typography variant="body2">No enrolled machines targeted</Typography>;
        }

        // TODO: reliability TBD?
        const deploymentPercent = Math.min(100, Math.max(0, (syncedMachines / totalMachines) * 100));

        return (
          <Stack
            spacing={0.75}
            width="100%"
          >
            <LinearProgressWithLabel value={deploymentPercent} />
            <Typography
              variant="body2"
              fontWeight="medium"
            >
              {syncedMachines.toLocaleString()} of {totalMachines.toLocaleString()} synced
            </Typography>
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
          key="toggle"
          icon={params.row.enabled ? <PauseCircleOutlineIcon /> : <PlayCircleOutlineIcon />}
          label={params.row.enabled ? "Pause Rule" : "Resume Rule"}
          onClick={(event) => {
            event.stopPropagation();
            onToggleStatus(params.row);
          }}
          disabled={togglingAppId === String(params.id)}
        />,
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

// Page component
export default function Applications() {
  const navigate = useNavigate();
  const confirm = useConfirm();
  const { showToast } = useToast();

  const { data: apps = [], error: appsError, isLoading } = useApplications();
  const deleteApplication = useDeleteApplication();
  const updateApplication = useUpdateApplication();

  // Local state
  const [dialogConfig, setDialogConfig] = useState<DialogConfig | null>(null);
  const [deletingAppId, setDeletingAppId] = useState<string | null>(null);
  const [togglingAppId, setTogglingAppId] = useState<string | null>(null);

  const dialogMode: DialogMode = dialogConfig?.mode ?? "create";

  // Effects
  useEffect(() => {
    if (!appsError) return;

    console.error("Applications query failed", appsError);
    showToast({
      message: appsError instanceof Error ? appsError.message : "Failed to load applications.",
      severity: "error",
    });
  }, [appsError, showToast]);

  const applicationErrorMessage = useMemo(() => {
    if (!appsError) return null;
    return appsError instanceof Error ? appsError.message : "Failed to load applications.";
  }, [appsError]);

  // Handlers
  const openCreateDialog = useCallback(() => {
    setDialogConfig({ mode: "create", application: null });
  }, []);

  const openEditDialog = useCallback((application: Application) => {
    setDialogConfig({ mode: "edit", application });
  }, []);

  const closeDialog = useCallback(() => {
    setDialogConfig(null);
  }, []);

  const handleDeleteApplication = useCallback(
    async (appId: string) => {
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
    },
    [deleteApplication, showToast],
  );

  const handleToggleApplication = useCallback(
    (application: Application) => {
      setTogglingAppId(application.id);
      void updateApplication
        .mutateAsync({
          appId: application.id,
          payload: { enabled: !application.enabled },
        })
        .then(() => {
          showToast({
            message: application.enabled ? "Application paused." : "Application resumed.",
            severity: "success",
          });
        })
        .catch((error: unknown) => {
          console.error("Toggle application failed", error);
          showToast({
            message: "Failed to update application status.",
            severity: "error",
          });
        })
        .finally(() => {
          setTogglingAppId(null);
        });
    },
    [showToast, updateApplication],
  );

  const handleRowClick = useCallback(
    (params: GridRowParams<Application>) => {
      void navigate(`/applications/${String(params.id)}`);
    },
    [navigate],
  );

  const handleConfirmDelete = useCallback(
    (appId: string, appName: string) => {
      confirm({
        title: "Delete Application Rule",
        description: `Are you sure you want to delete “${appName}”? This action cannot be undone.`,
        cancellationText: "Cancel",
        confirmationText: "Delete",
        confirmationButtonProps: { color: "error", variant: "contained" },
      })
        .then(() => void handleDeleteApplication(appId))
        .catch(() => {
          // user cancelled; no-op
        });
    },
    [confirm, handleDeleteApplication],
  );

  const handleDialogSuccess = useCallback(() => {
    const message = dialogMode === "edit" ? "Application rule updated." : "Application rule created.";

    showToast({ message, severity: "success" });
  }, [dialogMode, showToast]);

  const handleDialogError = useCallback(
    (message: string) => {
      showToast({ message, severity: "error" });
    },
    [showToast],
  );

  // Derived data
  const columns = useMemo(
    () =>
      createApplicationColumns({
        onEdit: openEditDialog,
        onRequestDelete: handleConfirmDelete,
        onToggleStatus: handleToggleApplication,
        deletingAppId,
        togglingAppId,
      }),
    [openEditDialog, handleConfirmDelete, handleToggleApplication, deletingAppId, togglingAppId],
  );

  const action = (
    <Button
      variant="contained"
      startIcon={<AddIcon />}
      onClick={openCreateDialog}
    >
      Add Application
    </Button>
  );

  // Render
  return (
    <>
      <Stack spacing={3}>
        <PageHeader
          title="Applications"
          subtitle="Define rules using reference-compatible identifiers."
          action={action}
        />

        {applicationErrorMessage && <Alert severity="error">{applicationErrorMessage}</Alert>}

        <Paper sx={{ height: 640, width: "100%" }}>
          <DataGrid
            rows={apps}
            columns={columns}
            loading={isLoading}
            disableRowSelectionOnClick
            onRowClick={handleRowClick}
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
        onError={handleDialogError}
      />
    </>
  );
}
