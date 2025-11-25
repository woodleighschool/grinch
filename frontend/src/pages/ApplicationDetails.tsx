import { Fragment, useCallback, useEffect, useMemo, useState, type ReactNode } from "react";
import Fuse, { type IFuseOptions } from "fuse.js";
import { useConfirm } from "material-ui-confirm";
import { useNavigate, useParams } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";

import type { Application, ApplicationScope, DirectoryGroup, DirectoryUser } from "../api";
import { ApiValidationError, createScope, deleteApplication, deleteScope, validateScope } from "../api";
import { useApplicationDetail, useGroups, useUsers } from "../hooks/useQueries";

import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Chip,
  Divider,
  Grid,
  InputAdornment,
  LinearProgress,
  List,
  ListItem,
  ListItemAvatar,
  ListItemButton,
  ListItemText,
  Paper,
  Stack,
  Tab,
  Tabs,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from "@mui/material";

import DeleteIcon from "@mui/icons-material/Delete";
import GroupIcon from "@mui/icons-material/Groups";
import PersonIcon from "@mui/icons-material/Person";
import SearchIcon from "@mui/icons-material/Search";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";

import { ApplicationDialog, EmptyState, PageHeader } from "../components";
import { useToast, type ToastOptions } from "../hooks/useToast";

export interface SelectedTarget {
  type: "group" | "user";
  id: string;
  name: string;
}

type ShowToast = (toast: ToastOptions) => void;

interface ScopeAssignmentRowProps {
  scope: ApplicationScope;
  onDelete: () => void;
  disabled: boolean;
}

function formatIdentifier(value: unknown) {
  if (typeof value === "string") return value;
  if (typeof value === "number") return value.toString();
  return "unknown";
}

function ScopeAssignmentRow({ scope, onDelete, disabled }: ScopeAssignmentRowProps) {
  const totalMemberCount = scope.effective_member_count ?? (Array.isArray(scope.effective_members) ? scope.effective_members.length : 0);

  const isGroup = scope.target_type === "group";
  const targetName = scope.target_display_name || (!isGroup && scope.target_upn) || `${isGroup ? "Group" : "User"} (${formatIdentifier(scope.target_id)})`;

  const formattedMemberCount = totalMemberCount.toLocaleString();
  const groupMemberCountLabel = isGroup && totalMemberCount > 0 ? `${formattedMemberCount} member${totalMemberCount === 1 ? "" : "s"}` : null;

  return (
    <ListItem
      alignItems="center"
      secondaryAction={
        <Button
          size="small"
          color="error"
          variant="outlined"
          onClick={onDelete}
          disabled={disabled}
        >
          Remove
        </Button>
      }
      sx={{ py: 1 }}
    >
      <ListItemAvatar sx={{ minWidth: 44 }}>
        <Avatar>{isGroup ? <GroupIcon fontSize="small" /> : <PersonIcon fontSize="small" />}</Avatar>
      </ListItemAvatar>

      <Box sx={{ flexGrow: 1 }}>
        <Stack
          direction="row"
          spacing={1}
          alignItems="center"
          flexWrap="wrap"
        >
          <Typography
            variant="subtitle1"
            fontWeight={600}
          >
            {targetName}
          </Typography>

          <Chip
            size="small"
            label={isGroup ? "Group" : "User"}
            variant="outlined"
          />
          {groupMemberCountLabel && (
            <Chip
              size="small"
              label={groupMemberCountLabel}
              variant="outlined"
            />
          )}
        </Stack>
      </Box>
    </ListItem>
  );
}

interface AssignmentSectionProps {
  label: "ALLOW" | "BLOCK" | "CEL";
  scopes: ApplicationScope[];
  onDeleteScope: (scopeId: string) => void;
  deletingScopeId: string | null;
}

function AssignmentSection({ label, scopes, onDeleteScope, deletingScopeId }: AssignmentSectionProps) {
  if (!scopes.length) return null;

  const chipColor = label === "ALLOW" ? "success" : label === "BLOCK" ? "error" : "info";

  return (
    <Box>
      <Stack
        direction="row"
        alignItems="center"
        spacing={1}
        mb={1.5}
      >
        <Chip
          label={label}
          color={chipColor}
          size="small"
        />
        <Typography
          variant="body2"
          color="text.secondary"
        >
          {scopes.length} assignment{scopes.length === 1 ? "" : "s"}
        </Typography>
      </Stack>

      <Paper elevation={2}>
        <List disablePadding>
          {scopes.map((scope, index) => (
            <Fragment key={scope.id}>
              <ScopeAssignmentRow
                scope={scope}
                onDelete={() => {
                  onDeleteScope(scope.id);
                }}
                disabled={deletingScopeId === scope.id}
              />
              {index < scopes.length - 1 && (
                <Divider
                  component="li"
                  role="presentation"
                />
              )}
            </Fragment>
          ))}
        </List>
      </Paper>
    </Box>
  );
}

interface TargetSelectorProps {
  groups: DirectoryGroup[];
  users: DirectoryUser[];
  onSelectTarget: (target: SelectedTarget) => void;
  selectedTarget: SelectedTarget | null;
  disabled: boolean;
}

const groupFuseOptions: IFuseOptions<DirectoryGroup> = {
  keys: [
    { name: "displayName", weight: 0.7 },
    { name: "description", weight: 0.3 },
  ],
  threshold: 0.3,
  ignoreLocation: true,
};

const userFuseOptions: IFuseOptions<DirectoryUser> = {
  keys: [
    { name: "displayName", weight: 0.7 },
    { name: "upn", weight: 0.3 },
  ],
  threshold: 0.3,
  ignoreLocation: true,
};

function TargetSelector({ groups, users, onSelectTarget, selectedTarget, disabled }: TargetSelectorProps) {
  const [activeTab, setActiveTab] = useState<"groups" | "users">("groups");
  const [groupSearchTerm, setGroupSearchTerm] = useState("");
  const [userSearchTerm, setUserSearchTerm] = useState("");

  const groupFuse = useMemo(() => new Fuse(groups, groupFuseOptions), [groups]);
  const userFuse = useMemo(() => new Fuse(users, userFuseOptions), [users]);

  const filterItems = (items: DirectoryGroup[] | DirectoryUser[], term: string) =>
    !term.trim() ? items : (activeTab === "groups" ? groupFuse : userFuse).search(term.trim()).map((r) => r.item);

  const activeSearchTerm = activeTab === "groups" ? groupSearchTerm : userSearchTerm;
  const setActiveSearchTerm = activeTab === "groups" ? setGroupSearchTerm : setUserSearchTerm;

  const activeItems =
    activeTab === "groups" ? (filterItems(groups, groupSearchTerm) as DirectoryGroup[]) : (filterItems(users, userSearchTerm) as DirectoryUser[]);

  const itemType: "group" | "user" = activeTab === "groups" ? "group" : "user";
  const trimmedActiveSearch = activeSearchTerm.trim();

  const handleSelectItem = (item: DirectoryGroup | DirectoryUser) => {
    if (disabled) return;
    onSelectTarget({ type: itemType, id: item.id, name: item.displayName });
  };

  return (
    <Stack
      spacing={2}
      sx={{ flex: 1, minHeight: 0 }}
    >
      <Tabs
        value={activeTab}
        onChange={(_, value: "groups" | "users" | null) => {
          if (value) setActiveTab(value);
        }}
        aria-label="Target type selector"
      >
        <Tab
          value="groups"
          label={`Groups (${groups.length.toLocaleString()})`}
        />
        <Tab
          value="users"
          label={`Users (${users.length.toLocaleString()})`}
        />
      </Tabs>

      <TextField
        type="search"
        size="small"
        label={`Search ${activeTab}`}
        value={activeSearchTerm}
        onChange={(event) => {
          setActiveSearchTerm(event.target.value);
        }}
        fullWidth
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

      <Paper
        elevation={2}
        sx={{ flex: 1, minHeight: 0, display: "flex", flexDirection: "column" }}
      >
        <Box sx={{ flex: 1, minHeight: 0, overflow: "hidden" }}>
          {!activeItems.length ? (
            <Box
              p={2}
              textAlign="center"
            >
              <Typography
                variant="body2"
                color="text.secondary"
              >
                {trimmedActiveSearch ? `No ${activeTab} match “${trimmedActiveSearch}”` : `No ${activeTab} are available.`}
              </Typography>
            </Box>
          ) : (
            <Virtuoso<DirectoryGroup | DirectoryUser>
              key={activeTab}
              data={activeItems}
              increaseViewportBy={{ top: 128, bottom: 256 }}
              style={{ height: "100%", width: "100%" }}
              itemContent={(index, item) => {
                const typedItem = item;
                const name = typedItem.displayName;
                const subtitle = itemType === "group" ? (typedItem as DirectoryGroup).description : (typedItem as DirectoryUser).upn;
                const isSelected = selectedTarget?.id === typedItem.id && selectedTarget.type === itemType;

                return (
                  <ListItem
                    component="div"
                    disablePadding
                    sx={(theme) => ({
                      borderBottom: index < activeItems.length - 1 ? `1px solid ${theme.palette.divider}` : "none",
                    })}
                  >
                    <ListItemButton
                      onClick={() => {
                        handleSelectItem(item);
                      }}
                      selected={isSelected}
                      disabled={disabled}
                    >
                      <ListItemAvatar>
                        <Avatar>{itemType === "group" ? <GroupIcon fontSize="small" /> : <PersonIcon fontSize="small" />}</Avatar>
                      </ListItemAvatar>

                      <ListItemText
                        disableTypography
                        primary={
                          <Typography
                            variant="body2"
                            fontWeight={600}
                            noWrap
                          >
                            {name}
                          </Typography>
                        }
                        secondary={
                          <Typography
                            variant="body2"
                            color="text.secondary"
                            noWrap
                          >
                            {subtitle}
                          </Typography>
                        }
                      />

                      {isSelected && (
                        <Chip
                          label="Selected"
                          size="small"
                          color="primary"
                        />
                      )}
                    </ListItemButton>
                  </ListItem>
                );
              }}
            />
          )}
        </Box>
      </Paper>
    </Stack>
  );
}

interface AssignmentManagerCardProps {
  assignment: AssignmentManagerState;
  groups: DirectoryGroup[];
  users: DirectoryUser[];
}

function AssignmentManagerCard({ assignment, groups, users }: AssignmentManagerCardProps) {
  const { selectedTarget, selectTarget, selectedAction, changeAction, assignmentError, clearAssignmentError, assignmentBusy, assignRule } = assignment;

  return (
    <Card
      elevation={1}
      sx={{ height: "100%", display: "flex", flexDirection: "column" }}
    >
      <CardHeader title="Assign to Groups or Users" />
      <CardContent sx={{ flex: 1, display: "flex", flexDirection: "column", minHeight: 0, p: 2 }}>
        <Stack
          spacing={2}
          sx={{ flex: 1, minHeight: 0 }}
        >
          {assignmentError && (
            <Alert
              severity="error"
              onClose={clearAssignmentError}
            >
              {assignmentError}
            </Alert>
          )}

          <TargetSelector
            groups={groups}
            users={users}
            onSelectTarget={selectTarget}
            selectedTarget={selectedTarget}
            disabled={assignmentBusy}
          />

          <Box sx={{ mt: "auto", pt: 2 }}>
            <Stack
              direction="row"
              spacing={2}
              alignItems="center"
              justifyContent="space-between"
            >
              <ToggleButtonGroup
                exclusive
                value={selectedAction}
                onChange={(_, value: "allow" | "block" | "cel" | null) => {
                  if (value) changeAction(value);
                }}
                size="small"
              >
                <ToggleButton
                  value="allow"
                  color="success"
                  disabled={assignment.celActionEnabled}
                >
                  Allow
                </ToggleButton>
                <ToggleButton
                  value="block"
                  color="error"
                  disabled={assignment.celActionEnabled}
                >
                  Block
                </ToggleButton>
                <ToggleButton
                  value="cel"
                  color="info"
                  disabled={!assignment.celActionEnabled}
                >
                  CEL
                </ToggleButton>
              </ToggleButtonGroup>

              <Button
                variant="contained"
                onClick={() => void assignRule()}
                disabled={assignmentBusy || !selectedTarget}
              >
                {assignmentBusy ? "Assigning..." : "Assign Rule"}
              </Button>
            </Stack>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}

interface AssignmentsCardProps {
  allowScopes: ApplicationScope[];
  blockScopes: ApplicationScope[];
  celScopes: ApplicationScope[];
  onDeleteScope: (scopeId: string) => void;
  deletingScopeId: string | null;
}

function AssignmentsCard({ allowScopes, blockScopes, celScopes, onDeleteScope, deletingScopeId }: AssignmentsCardProps) {
  const hasAssignments = allowScopes.length > 0 || blockScopes.length > 0 || celScopes.length > 0;

  return (
    <Card
      elevation={1}
      sx={{ height: "100%", display: "flex", flexDirection: "column" }}
    >
      <CardHeader
        title="Current Assignments"
        subheader="Allow and block scopes."
      />
      <CardContent sx={{ overflowY: "auto", flex: 1 }}>
        <Stack spacing={3}>
          <AssignmentSection
            label="ALLOW"
            scopes={allowScopes}
            onDeleteScope={onDeleteScope}
            deletingScopeId={deletingScopeId}
          />
          <AssignmentSection
            label="BLOCK"
            scopes={blockScopes}
            onDeleteScope={onDeleteScope}
            deletingScopeId={deletingScopeId}
          />
          <AssignmentSection
            label="CEL"
            scopes={celScopes}
            onDeleteScope={onDeleteScope}
            deletingScopeId={deletingScopeId}
          />
          {!hasAssignments && <Alert severity="info">No assignments yet.</Alert>}
        </Stack>
      </CardContent>
    </Card>
  );
}

interface AssignmentManagerHookParams {
  app: Application | null;
  appId: string | undefined;
  refetch: () => Promise<unknown>;
  showToast: ShowToast;
}

interface AssignmentManagerState {
  selectedTarget: SelectedTarget | null;
  selectTarget: (target: SelectedTarget | null) => void;
  selectedAction: "allow" | "block" | "cel";
  changeAction: (action: "allow" | "block" | "cel") => void;
  celActionEnabled: boolean;
  assignmentError: string | null;
  clearAssignmentError: () => void;
  assignmentBusy: boolean;
  deletingScopeId: string | null;
  assignRule: () => Promise<void>;
  removeScope: (scopeId: string) => void;
}

function useAssignmentManager({ app, appId, refetch, showToast }: AssignmentManagerHookParams): AssignmentManagerState {
  const [selectedTarget, setSelectedTarget] = useState<SelectedTarget | null>(null);
  const [selectedAction, setSelectedAction] = useState<"allow" | "block" | "cel">("allow");
  const [assignmentError, setAssignmentError] = useState<string | null>(null);
  const [assignmentBusy, setAssignmentBusy] = useState(false);
  const [deletingScopeId, setDeletingScopeId] = useState<string | null>(null);

  const applicationId = app?.id ?? appId;
  const celActionEnabled = Boolean(app?.cel_enabled);

  useEffect(() => {
    if (assignmentError && selectedTarget) setAssignmentError(null);
  }, [assignmentError, selectedTarget]);

  useEffect(() => {
    if (celActionEnabled) {
      setSelectedAction("cel");
    } else if (selectedAction === "cel") {
      setSelectedAction("allow");
    }
  }, [celActionEnabled, selectedAction]);

  const assignRule = useCallback(async () => {
    if (!selectedTarget) {
      setAssignmentError("Please select a user or group first.");
      return;
    }

    if (!applicationId) {
      setAssignmentError("Missing application identifier.");
      return;
    }

    setAssignmentBusy(true);

    try {
      const validation = await validateScope(applicationId, {
        target_type: selectedTarget.type,
        target_id: selectedTarget.id,
        action: selectedAction,
      });

      const { target_type, target_id, action } = validation.normalised;

      await createScope(applicationId, {
        target_type,
        target_id,
        action,
      });

      await refetch();
      setSelectedTarget(null);

      showToast({ message: "Assignment created.", severity: "success" });
    } catch (error) {
      console.error("Assign rule failed", error);
      if (error instanceof ApiValidationError) {
        setAssignmentError(error.fieldErrors.target_id ?? error.message);
      } else {
        setAssignmentError(error instanceof Error ? error.message : "Failed to assign rule.");
      }
    } finally {
      setAssignmentBusy(false);
    }
  }, [selectedTarget, selectedAction, applicationId, refetch, showToast]);

  const handleDeleteScope = useCallback(
    async (scopeId: string) => {
      if (!applicationId) return;

      setDeletingScopeId(scopeId);

      try {
        await deleteScope(applicationId, scopeId);
        await refetch();

        showToast({ message: "Assignment removed.", severity: "success" });
      } catch (error) {
        console.error("Delete scope failed", error);
        setAssignmentError(error instanceof Error ? error.message : "Failed to remove assignment.");
      } finally {
        setDeletingScopeId(null);
      }
    },
    [applicationId, refetch, showToast],
  );

  return {
    selectedTarget,
    selectTarget: setSelectedTarget,
    selectedAction,
    changeAction: setSelectedAction,
    celActionEnabled,
    assignmentError,
    clearAssignmentError: () => {
      setAssignmentError(null);
    },
    assignmentBusy,
    deletingScopeId,
    assignRule,
    removeScope: (scopeId) => {
      void handleDeleteScope(scopeId);
    },
  };
}

export default function ApplicationDetails() {
  const { appId } = useParams<{ appId: string }>();
  const navigate = useNavigate();
  const confirm = useConfirm();

  const [deletingApp, setDeletingApp] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);

  const { showToast } = useToast();

  const { data: groups = [] } = useGroups();
  const { data: users = [] } = useUsers();
  const {
    data: applicationDetail,
    isLoading,
    error,
    refetch,
  } = useApplicationDetail(appId, {
    includeMembers: true,
  });

  const app = applicationDetail?.application ?? null;
  const scopes = applicationDetail?.scopes ?? [];

  const allowScopes = scopes.filter((scope) => scope.action === "allow");
  const blockScopes = scopes.filter((scope) => scope.action === "block");
  const celScopes = scopes.filter((scope) => scope.action === "cel");

  const assignedGroupIds = new Set(scopes.filter((scope) => scope.target_type === "group").map((scope) => scope.target_id));
  const assignedUserIds = new Set(scopes.filter((scope) => scope.target_type === "user").map((scope) => scope.target_id));

  const availableGroups = groups.filter((group) => !assignedGroupIds.has(group.id));
  const availableUsers = users.filter((user) => !assignedUserIds.has(user.id));

  const assignment = useAssignmentManager({
    app,
    appId,
    refetch,
    showToast,
  });

  useEffect(() => {
    if (!error) return;
    console.error("Application detail load failed", error);
    showToast({
      message: error instanceof Error ? error.message : "Failed to load application details.",
      severity: "error",
    });
  }, [error, showToast]);

  const handleEditSuccess = useCallback(() => {
    showToast({ message: "Application updated.", severity: "success" });
    void refetch();
  }, [showToast, refetch]);

  const handleDeleteApplication = useCallback(async () => {
    const targetAppId = app?.id ?? appId;
    if (!targetAppId) return;

    setDeletingApp(true);
    try {
      await deleteApplication(targetAppId);
      void navigate("/applications", { replace: true });
    } catch (error) {
      console.error("Delete application failed", error);
      showToast({
        message: error instanceof Error ? error.message : "Failed to delete application.",
        severity: "error",
      });
    } finally {
      setDeletingApp(false);
    }
  }, [app, appId, navigate, showToast]);

  const action = useMemo<ReactNode | undefined>(() => {
    if (!app) return undefined;

    return (
      <Stack
        direction="row"
        spacing={1}
      >
        <Button
          variant="contained"
          onClick={() => {
            setEditDialogOpen(true);
          }}
        >
          Edit
        </Button>
        <Button
          color="error"
          variant="outlined"
          startIcon={<DeleteIcon />}
          onClick={() =>
            void confirm({
              title: "Delete Application",
              description: `Are you sure you want to delete "${app.name}"? This action cannot be undone.`,
              cancellationText: "Cancel",
              confirmationText: "Delete",
              confirmationButtonProps: { color: "error", variant: "contained" },
            })
              .then(() => void handleDeleteApplication())
              .catch(() => {})
          }
          disabled={deletingApp}
        >
          Delete
        </Button>
      </Stack>
    );
  }, [app, deletingApp, confirm, handleDeleteApplication]);

  const pageTitle = app?.name ?? "Application Details";
  const subtitleParts = [app?.rule_type, app?.identifier].filter(Boolean) as string[];
  const pageSubtitle = subtitleParts.length ? subtitleParts.join(" • ") : undefined;

  const breadcrumbs = [{ label: "Applications", to: "/applications" }, { label: pageTitle }];

  let content: ReactNode = null;

  if (!appId) {
    content = <Alert severity="error">Missing application identifier.</Alert>;
  } else if (isLoading) {
    content = <LinearProgress />;
  } else if (error) {
    content = <Alert severity="error">{error instanceof Error ? error.message : "Failed to load application details."}</Alert>;
  } else if (!app) {
    content = (
      <EmptyState
        title="Application not found"
        description="The requested application could not be found."
      />
    );
  } else {
    content = (
      <Stack spacing={2}>
        <Alert
          severity="info"
          variant="outlined"
          icon={<InfoOutlinedIcon fontSize="small" />}
        >
          Assignments are scoped based on the device&apos;s <strong>primary user</strong> received during preflight, not the currently logged-in user.
        </Alert>

        <Grid
          container
          spacing={2}
          sx={{ height: "calc(100vh - 280px)", minHeight: 500 }}
        >
          <Grid
            size={{ xs: 12, lg: 7 }}
            sx={{ height: "100%" }}
          >
            <AssignmentManagerCard
              assignment={assignment}
              groups={availableGroups}
              users={availableUsers}
            />
          </Grid>

          <Grid
            size={{ xs: 12, lg: 5 }}
            sx={{ height: "100%" }}
          >
            <AssignmentsCard
              allowScopes={allowScopes}
              blockScopes={blockScopes}
              celScopes={celScopes}
              onDeleteScope={assignment.removeScope}
              deletingScopeId={assignment.deletingScopeId}
            />
          </Grid>
        </Grid>
      </Stack>
    );
  }

  return (
    <>
      <Stack spacing={3}>
        <PageHeader
          title={pageTitle}
          subtitle={pageSubtitle}
          breadcrumbs={breadcrumbs}
          action={action}
        />
        {content}
      </Stack>

      <ApplicationDialog
        open={editDialogOpen}
        mode="edit"
        application={app ?? null}
        onClose={() => {
          setEditDialogOpen(false);
        }}
        onSuccess={handleEditSuccess}
        onError={(message) => {
          showToast({
            message,
            severity: "error",
          });
        }}
      />
    </>
  );
}
