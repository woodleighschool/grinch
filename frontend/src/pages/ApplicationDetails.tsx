import { Fragment, useEffect, useMemo, useState } from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import { formatDate } from "../utils/dates";
import type { Application, ApplicationScope, DirectoryGroup, DirectoryUser } from "../api";
import { ApiValidationError, createScope, deleteApplication, deleteScope, validateScope } from "../api";
import { useApplicationDetail, useGroups, useUsers } from "../hooks/useQueries";
import { Virtuoso } from "react-virtuoso";
import { useDebouncedValue } from "../hooks/useDebouncedValue";
import {
  Alert,
  Avatar,
  Box,
  Grid,
  Button,
  Card,
  CardContent,
  CardHeader,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Divider,
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
  Tooltip,
  Typography,
  InputAdornment,
  LinearProgress,
  CircularProgress,
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import DeleteIcon from "@mui/icons-material/Delete";
import GroupIcon from "@mui/icons-material/Groups";
import PersonIcon from "@mui/icons-material/Person";
import SearchIcon from "@mui/icons-material/Search";
import { PageSnackbar, type PageToast } from "../components";

export interface SelectedTarget {
  type: "group" | "user";
  id: string;
  name: string;
}

interface ScopeAssignmentRowProps {
  scope: ApplicationScope;
  onDelete: () => void;
  disabled: boolean;
}

function ScopeAssignmentRow({ scope, onDelete, disabled }: ScopeAssignmentRowProps) {
  const memberUsers = Array.isArray(scope.effective_members) ? (scope.effective_members as DirectoryUser[]) : [];
  const maxDisplayedMembers = 10;
  const displayedMembers = memberUsers.slice(0, maxDisplayedMembers);
  const remainingCount = Math.max((scope.effective_member_count ?? memberUsers.length) - displayedMembers.length, 0);

  const targetName = (() => {
    if (scope.target_type === "group") return scope.target_display_name || `Group (${scope.target_id})`;
    return scope.target_display_name || scope.target_upn || `User (${scope.target_id})`;
  })();

  const isGroup = scope.target_type === "group";

  return (
    <ListItem
      alignItems="flex-start"
      secondaryAction={
        <Button size="small" color="error" variant="outlined" onClick={onDelete} disabled={disabled}>
          Remove
        </Button>
      }
    >
      <ListItemAvatar>
        <Avatar>{isGroup ? <GroupIcon fontSize="small" /> : <PersonIcon fontSize="small" />}</Avatar>
      </ListItemAvatar>
      <ListItemText
        primary={
          <Stack direction={{ sm: "row" }} spacing={1} alignItems={{ sm: "center" }}>
            <Typography variant="subtitle1" fontWeight={600}>
              {targetName}
            </Typography>
            <Chip size="small" label={isGroup ? "Group" : "User"} variant="outlined" />
            <Chip size="small" color={scope.action === "allow" ? "success" : "error"} label={scope.action.toUpperCase()} />
          </Stack>
        }
        secondary={
          <Stack spacing={1} mt={1} pr={2}>
            <Typography variant="body2" color="text.secondary">
              Added {formatDate(scope.created_at)}
            </Typography>
            {isGroup && (
              <Box>
                {memberUsers.length === 0 ? (
                  <Typography variant="body2" color="text.secondary">
                    No users currently in this group.
                  </Typography>
                ) : (
                  <Stack direction="row" flexWrap="wrap" gap={1}>
                    {displayedMembers.map((user) => (
                      <Chip key={user.id} size="small" variant="outlined" label={user.displayName} />
                    ))}
                    {remainingCount > 0 && <Chip size="small" label={`+${remainingCount} more`} />}
                  </Stack>
                )}
              </Box>
            )}
          </Stack>
        }
      />
    </ListItem>
  );
}

interface AssignmentSectionProps {
  label: string;
  scopes: ApplicationScope[];
  onDeleteScope: (scopeId: string) => void;
  deletingScopeId: string | null;
}

function AssignmentSection({ label, scopes, onDeleteScope, deletingScopeId }: AssignmentSectionProps) {
  if (scopes.length === 0) return null;

  return (
    <Box>
      <Stack direction="row" alignItems="center" spacing={1} mb={1.5}>
        <Chip label={label} color={label === "ALLOW" ? "success" : "error"} size="small" />
        <Typography variant="body2" color="text.secondary">
          {scopes.length} assignment{scopes.length === 1 ? "" : "s"}
        </Typography>
      </Stack>
      <Paper elevation={2}>
        <List disablePadding>
          {scopes.map((scope, index) => (
            <Fragment key={scope.id}>
              <ScopeAssignmentRow scope={scope} onDelete={() => onDeleteScope(scope.id)} disabled={deletingScopeId === scope.id} />
              {index < scopes.length - 1 && <Divider component="li" />}
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

function TargetSelector({ groups, users, onSelectTarget, selectedTarget, disabled }: TargetSelectorProps) {
  const [activeTab, setActiveTab] = useState<"groups" | "users">("groups");
  const [groupSearchTerm, setGroupSearchTerm] = useState("");
  const [userSearchTerm, setUserSearchTerm] = useState("");
  const debouncedGroupSearch = useDebouncedValue(groupSearchTerm, 250);
  const debouncedUserSearch = useDebouncedValue(userSearchTerm, 250);

  const {
    data: groupResults = [],
    isFetching: groupLoading,
    isError: groupError,
    error: groupErrorObj,
    isLoading: groupInitialLoading,
    refetch: refetchGroups,
  } = useGroups({ search: debouncedGroupSearch });

  const {
    data: userResults = [],
    isFetching: userLoading,
    isError: userError,
    error: userErrorObj,
    isLoading: userInitialLoading,
    refetch: refetchUsers,
  } = useUsers({ search: debouncedUserSearch });

  const activeSearchTerm = activeTab === "groups" ? groupSearchTerm : userSearchTerm;
  const setActiveSearchTerm = activeTab === "groups" ? setGroupSearchTerm : setUserSearchTerm;
  const activeItems = activeTab === "groups" ? groupResults : userResults;
  const activeLoading = activeTab === "groups" ? groupLoading : userLoading;
  const activeInitialLoading = activeTab === "groups" ? groupInitialLoading : userInitialLoading;
  const activeErrorObj = activeTab === "groups" ? groupErrorObj : userErrorObj;
  const activeRefetch = activeTab === "groups" ? refetchGroups : refetchUsers;
  const trimmedActiveSearch = activeSearchTerm.trim();

  const handleSelectItem = (item: DirectoryGroup | DirectoryUser) => {
    if (disabled) return;
    const type = activeTab === "groups" ? "group" : "user";
    const name = type === "group" ? (item as DirectoryGroup).displayName : (item as DirectoryUser).displayName;
    onSelectTarget({ type, id: item.id, name });
  };

  return (
    <Stack spacing={2}>
      <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)} aria-label="Target type selector">
        <Tab value="groups" label={`Groups (${groups.length})`} />
        <Tab value="users" label={`Users (${users.length})`} />
      </Tabs>

      <TextField
        type="search"
        size="small"
        label={`Search ${activeTab}`}
        value={activeSearchTerm}
        onChange={(e) => setActiveSearchTerm(e.target.value)}
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

      <Paper elevation={2}>
        <Box style={{ height: 360, overflow: "hidden" }}>
          {activeInitialLoading ? (
            <LinearProgress />
          ) : (groupError || userError) && activeErrorObj ? (
            <Box p={2} textAlign="center">
              <Typography color="error" variant="body2" gutterBottom>
                {activeErrorObj instanceof Error ? activeErrorObj.message : `Failed to load ${activeTab}.`}
              </Typography>
              <Button size="small" variant="outlined" onClick={() => void activeRefetch()}>
                Retry
              </Button>
            </Box>
          ) : activeItems.length === 0 ? (
            <Box p={2} textAlign="center">
              <Typography variant="body2" color="text.secondary">
                {trimmedActiveSearch ? `No ${activeTab} match “${trimmedActiveSearch}”` : `No ${activeTab} are available.`}
              </Typography>
            </Box>
          ) : (
            <Virtuoso<DirectoryGroup | DirectoryUser>
              data={activeItems as Array<DirectoryGroup | DirectoryUser>}
              overscan={200}
              style={{ height: 360 }}
              itemContent={(_, item) => {
                const type = activeTab === "groups" ? "group" : "user";
                const name = type === "group" ? (item as DirectoryGroup).displayName : (item as DirectoryUser).displayName;
                const subtitle = type === "group" ? (item as DirectoryGroup).description : (item as DirectoryUser).upn;
                const isSelected = selectedTarget?.id === item.id && selectedTarget.type === type;
                return (
                  <ListItem disablePadding divider>
                    <ListItemButton onClick={() => handleSelectItem(item)} selected={isSelected} disabled={disabled}>
                      <ListItemAvatar>
                        <Avatar>{type === "group" ? <GroupIcon fontSize="small" /> : <PersonIcon fontSize="small" />}</Avatar>
                      </ListItemAvatar>
                      <ListItemText
                        primary={
                          <Stack direction={{ sm: "row" }} spacing={1} alignItems={{ sm: "center" }}>
                            <Typography variant="body2" fontWeight={600}>
                              {name}
                            </Typography>
                            {activeLoading && <CircularProgress size={16} />}
                          </Stack>
                        }
                        secondary={<Typography variant="body2">{subtitle}</Typography>}
                      />
                      {isSelected && <Chip label="Selected" size="small" color="primary" />}
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

export default function ApplicationDetails() {
  const { appId } = useParams<{ appId: string }>();
  const navigate = useNavigate();

  const [selectedTarget, setSelectedTarget] = useState<SelectedTarget | null>(null);
  const [selectedAction, setSelectedAction] = useState<"allow" | "block">("allow");
  const [assignmentError, setAssignmentError] = useState<string | null>(null);
  const [assignmentBusy, setAssignmentBusy] = useState(false);
  const [deletingScopeId, setDeletingScopeId] = useState<string | null>(null);
  const [confirmDeleteApp, setConfirmDeleteApp] = useState<{ appId: string; appName: string } | null>(null);
  const [deletingApp, setDeletingApp] = useState(false);
  const [toast, setToast] = useState<PageToast>({ open: false, message: "", severity: "error" });

  const { data: groups = [] } = useGroups();
  const { data: users = [] } = useUsers();
  const { data: applicationDetail, isLoading, error, refetch } = useApplicationDetail(appId, { includeMembers: true });

  const app: Application | null = applicationDetail?.application ?? null;
  const scopes: ApplicationScope[] = applicationDetail?.scopes ?? [];
  const allowScopes = useMemo(() => scopes.filter((s) => s.action === "allow"), [scopes]);
  const blockScopes = useMemo(() => scopes.filter((s) => s.action === "block"), [scopes]);
  const totalAssignments = scopes.length;

  useEffect(() => {
    if (error) {
      console.error("Application detail load failed", error);
      setToast({ open: true, message: "Failed to load application details.", severity: "error" });
    }
  }, [error]);

  const handleToastClose = () => setToast((prev) => ({ ...prev, open: false }));

  useEffect(() => {
    if (assignmentError && selectedTarget) setAssignmentError(null);
  }, [selectedTarget, assignmentError]);

  async function handleAssignRule() {
    if (!selectedTarget) {
      setAssignmentError("Please select a user or group first.");
      return;
    }
    const targetAppId = app?.id ?? appId;
    if (!targetAppId) {
      setAssignmentError("Missing application identifier.");
      return;
    }

    setAssignmentBusy(true);
    try {
      const validation = await validateScope({
        application_id: targetAppId,
        target_type: selectedTarget.type,
        target_id: selectedTarget.id,
        action: selectedAction,
      });
      const n = validation.normalised;
      await createScope(targetAppId, { target_type: n.target_type, target_id: n.target_id, action: n.action });
      await refetch();
      setSelectedTarget(null);
      setToast({ open: true, message: "Assignment created.", severity: "success" });
    } catch (err) {
      console.error("Assign rule failed", err);
      if (err instanceof ApiValidationError) setAssignmentError(err.fieldErrors.target_id ?? err.message);
      else setAssignmentError(err instanceof Error ? err.message : "Failed to assign rule.");
    } finally {
      setAssignmentBusy(false);
    }
  }

  async function handleDeleteScope(scopeId: string) {
    const targetAppId = app?.id ?? appId;
    if (!targetAppId) return;

    setDeletingScopeId(scopeId);
    try {
      await deleteScope(targetAppId, scopeId);
      await refetch();
      setToast({ open: true, message: "Assignment removed.", severity: "success" });
    } catch (err) {
      console.error("Delete scope failed", err);
      setAssignmentError(err instanceof Error ? err.message : "Failed to remove assignment.");
    } finally {
      setDeletingScopeId(null);
    }
  }

  async function handleDeleteApplication() {
    const targetAppId = confirmDeleteApp?.appId ?? app?.id ?? appId;
    if (!targetAppId) return;

    setDeletingApp(true);
    try {
      await deleteApplication(targetAppId);
      navigate("/applications", { replace: true });
    } catch (err) {
      console.error("Delete application failed", err);
      setAssignmentError(err instanceof Error ? err.message : "Failed to delete application.");
      setDeletingApp(false);
    } finally {
      setConfirmDeleteApp(null);
    }
  }

  if (!appId) {
    return (
      <Card elevation={1}>
        <CardHeader title="Application Details" />
        <CardContent>
          <Stack spacing={2}>
            <Alert severity="error">Missing application identifier.</Alert>
            <Button component={RouterLink} to="/applications" variant="contained">
              Back to applications
            </Button>
          </Stack>
        </CardContent>
      </Card>
    );
  }

  return (
    <Stack spacing={3}>
      <Stack direction="row" spacing={1} alignItems="center">
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate(-1)}>
          Back
        </Button>
        <Typography variant="h4">Application Details</Typography>
      </Stack>

      <Card elevation={1}>
        <CardHeader
          title={app?.name || "Loading..."}
          subheader="Manage assignments, view current scopes, and adjust access for this application."
          action={
            <Tooltip title="Delete application">
              <span>
                <Button
                  color="error"
                  variant="outlined"
                  startIcon={<DeleteIcon />}
                  onClick={() => app && setConfirmDeleteApp({ appId: app.id, appName: app.name })}
                  disabled={!app || deletingApp}
                >
                  Delete
                </Button>
              </span>
            </Tooltip>
          }
        />
        {isLoading && <LinearProgress />}
        <CardContent>
          {app ? (
            <Stack spacing={2}>
              <Stack direction="row" flexWrap="wrap" gap={1.5}>
                <Chip label={app.rule_type} variant="outlined" />
                <Chip label={app.identifier ?? "—"} variant="outlined" />
                <Chip label={`Assignments: ${totalAssignments}`} />
              </Stack>
              {app.description && <Typography color="text.secondary">{app.description}</Typography>}
            </Stack>
          ) : (
            !error && <Typography color="text.secondary">Loading application...</Typography>
          )}
        </CardContent>
      </Card>

      <Grid container spacing={3}>
        <Grid size={{ xs: 12, lg: 6 }}>
          <Card elevation={1}>
            <CardHeader title="Assign to Groups or Users" />
            <CardContent>
              <Stack spacing={3}>
                {assignmentError && (
                  <Alert severity="error" onClose={() => setAssignmentError(null)}>
                    {assignmentError}
                  </Alert>
                )}
                <TargetSelector groups={groups} users={users} onSelectTarget={setSelectedTarget} selectedTarget={selectedTarget} disabled={assignmentBusy} />
                <Box>
                  <Typography variant="subtitle2" gutterBottom>
                    Action
                  </Typography>
                  <ToggleButtonGroup exclusive value={selectedAction} onChange={(_, v) => v && setSelectedAction(v)} size="small">
                    <ToggleButton value="allow" color="success">
                      Allow
                    </ToggleButton>
                    <ToggleButton value="block" color="error">
                      Block
                    </ToggleButton>
                  </ToggleButtonGroup>
                </Box>
                {/* TODO: Do something better on validation/duplicate error */}
                <Button variant="contained" onClick={handleAssignRule} disabled={assignmentBusy || !selectedTarget}>
                  Assign Rule
                </Button>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, lg: 6 }}>
          <Card elevation={1}>
            <CardHeader title="Current Assignments" subheader="Allow and block scopes targeting this application." />
            <CardContent>
              {/* TODO: To infinity and beyond... these needs to be a scrollable list */}
              <Stack spacing={3}>
                <AssignmentSection label="ALLOW" scopes={allowScopes} onDeleteScope={handleDeleteScope} deletingScopeId={deletingScopeId} />
                <AssignmentSection label="BLOCK" scopes={blockScopes} onDeleteScope={handleDeleteScope} deletingScopeId={deletingScopeId} />
                {allowScopes.length === 0 && blockScopes.length === 0 && <Alert severity="info">No assignments yet. Use the controls above to add one.</Alert>}
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Dialog open={!!confirmDeleteApp} onClose={() => setConfirmDeleteApp(null)} aria-labelledby="confirm-delete-title">
        <DialogTitle id="confirm-delete-title">Delete Application</DialogTitle>
        <DialogContent>
          <DialogContentText>Are you sure you want to delete "{confirmDeleteApp?.appName}"? This action cannot be undone.</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmDeleteApp(null)} disabled={deletingApp}>
            Cancel
          </Button>
          <Button onClick={handleDeleteApplication} color="error" variant="contained" disabled={deletingApp}>
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      <PageSnackbar toast={toast} onClose={handleToastClose} autoHideDuration={3500} />
    </Stack>
  );
}
