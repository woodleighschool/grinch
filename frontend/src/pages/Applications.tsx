import { useEffect, useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import { useForm, Controller } from "react-hook-form";
import { ApiValidationError, validateApplication } from "../api";
import { useApplications, useCreateApplication, useUpdateApplication, useDeleteApplication } from "../hooks/useQueries";
import { useDebouncedValue } from "../hooks/useDebouncedValue";
import {
  Avatar,
  Button,
  Card,
  CardActionArea,
  CardContent,
  CardHeader,
  Chip,
  Dialog,
  DialogActions,
  CardActions,
  DialogContent,
  DialogTitle,
  Divider,
  FormControl,
  Grid,
  IconButton,
  InputAdornment,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Tooltip,
  Typography,
  LinearProgress,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import HelpIcon from "@mui/icons-material/Help";
import ShieldIcon from "@mui/icons-material/Shield";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import PauseIcon from "@mui/icons-material/Pause";
import DeleteIcon from "@mui/icons-material/Delete";
import { PageSnackbar, type PageToast } from "../components";

const applicationRuleTypes = ["BINARY", "CERTIFICATE", "SIGNINGID", "TEAMID", "CDHASH"] as const;
type ApplicationRuleType = (typeof applicationRuleTypes)[number];

type ApplicationRuleTypeMetadata = {
  label: string;
  placeholder: string;
  example: string;
  description: string;
  referenceGroup?: "signingChain";
};

const applicationRuleTypeMetadata: Record<ApplicationRuleType, ApplicationRuleTypeMetadata> = {
  BINARY: {
    label: "SHA-256",
    placeholder: "f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef",
    example: "f820d4f4ed9aade09e1810314f21e4152988c54e489245670cc9de5639bc14ef",
    description: "Matches one exact copy of a program using its full file hash, so nothing else counts as that app.",
  },
  CERTIFICATE: {
    label: "SHA-256",
    placeholder: "1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64",
    example: "1afd16f5b920f0d3b5f841aace6e948d6190ea8b5156b02deb36572d1d082f64",
    referenceGroup: "signingChain",
    description: "Covers every app signed with this certificate, so trusting or blocking it affects all software that uses that signer.",
  },
  SIGNINGID: {
    label: "Signing ID",
    placeholder: "ZMCG7MLDV9:com.northpolesec.santa",
    example: "ZMCG7MLDV9:com.northpolesec.santa",
    description: "Groups every version of the same app when its bundle ID and team stay the same, letting you treat that app together across updates.",
  },
  TEAMID: {
    label: "Team ID",
    placeholder: "ZMCG7MLDV9",
    example: "ZMCG7MLDV9",
    description: "Applies to every app from one developer team, useful when you want to trust or block that entire publisher.",
  },
  CDHASH: {
    label: "CDHash",
    placeholder: "a9fdcbc0427a0a585f91bbc7342c261c8ead1942",
    example: "a9fdcbc0427a0a585f91bbc7342c261c8ead1942",
    description: "Matches the internal hash stored in the code signature so it points to one specific build of the app.",
  },
};

type ApplicationFormData = {
  name: string;
  rule_type: ApplicationRuleType;
  identifier: string;
  description?: string;
};

const applicationFormDefaultValues: ApplicationFormData = {
  name: "",
  rule_type: "BINARY",
  identifier: "",
  description: "",
};

type RuleTypeEntry = { type: ApplicationRuleType; meta: ApplicationRuleTypeMetadata };

const signingChainReference = applicationRuleTypeMetadata.CERTIFICATE;
const primaryRuleTypeEntries: RuleTypeEntry[] = applicationRuleTypes
  .filter((t) => applicationRuleTypeMetadata[t].referenceGroup !== "signingChain")
  .map((t) => ({ type: t, meta: applicationRuleTypeMetadata[t] }));

interface AssignmentStats {
  allowCount: number;
  blockCount: number;
  totalUsersCovered: number;
}

export default function Applications() {
  const [appSearch, setAppSearch] = useState("");
  const debouncedAppSearch = useDebouncedValue(appSearch, 300);
  const { data: apps = [], isLoading: appsLoading, isFetching: appsFetching, error: appsError } = useApplications({ search: debouncedAppSearch });
  const createApplication = useCreateApplication();
  const updateApplication = useUpdateApplication();
  const deleteApplication = useDeleteApplication();

  const [deletingAppId, setDeletingAppId] = useState<string | null>(null);
  const [updatingAppId, setUpdatingAppId] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ appId: string; appName: string } | null>(null);
  const [toast, setToast] = useState<PageToast>({ open: false, message: "", severity: "error" });

  const {
    register,
    handleSubmit,
    watch,
    reset,
    control,
    setError,
    clearErrors,
    formState: { errors, isSubmitting },
  } = useForm<ApplicationFormData>({ defaultValues: applicationFormDefaultValues });

  const watchedRuleType = watch("rule_type");
  const identifierPlaceholder = applicationRuleTypeMetadata[watchedRuleType]?.placeholder ?? "Enter identifier...";

  const trimmedSearch = appSearch.trim();
  const hasSearchTerm = trimmedSearch.length > 0;

  const totalScopes = useMemo(() => apps.reduce((sum, app) => sum + (app.assignment_stats?.total_scopes ?? 0), 0), [apps]);

  function getAssignmentStats(app: { assignment_stats?: { allow_scopes?: number; block_scopes?: number; total_users?: number } }): AssignmentStats {
    const stats = app.assignment_stats;
    return { allowCount: stats?.allow_scopes ?? 0, blockCount: stats?.block_scopes ?? 0, totalUsersCovered: stats?.total_users ?? 0 };
  }

  async function handleCreateApp(data: ApplicationFormData) {
    clearErrors();
    try {
      const { normalised } = await validateApplication(data);
      const payload: { name: string; rule_type: string; identifier: string; description?: string } = {
        name: normalised.name,
        rule_type: normalised.rule_type,
        identifier: normalised.identifier,
      };
      if (normalised.description) {
        payload.description = normalised.description;
      }
      await createApplication.mutateAsync(payload);
      reset(applicationFormDefaultValues);
      setToast({ open: true, message: "Application rule created.", severity: "success" });
    } catch (err) {
      if (err instanceof ApiValidationError) {
        Object.entries(err.fieldErrors).forEach(([field, message]) => {
          if (field === "name" || field === "rule_type" || field === "identifier" || field === "description") {
            setError(field as keyof ApplicationFormData, { type: "server", message });
          }
        });
        return;
      }
      console.error("Create application failed", err);
      setToast({ open: true, message: "Failed to create application rule.", severity: "error" });
    }
  }

  function requestDeleteApplication(appId: string, appName: string) {
    setConfirmDelete({ appId, appName });
  }

  async function handleDeleteApplication(appId: string) {
    setDeletingAppId(appId);
    try {
      await deleteApplication.mutateAsync(appId);
      setToast({ open: true, message: "Application rule deleted.", severity: "success" });
    } catch (err) {
      console.error("Delete application failed", err);
      setToast({ open: true, message: "Failed to delete application rule.", severity: "error" });
    } finally {
      setDeletingAppId(null);
      setConfirmDelete(null);
    }
  }

  async function handleToggleEnabled(appId: string, currentEnabled: boolean) {
    setUpdatingAppId(appId);
    try {
      await updateApplication.mutateAsync({ appId, payload: { enabled: !currentEnabled } });
      setToast({ open: true, message: currentEnabled ? "Disabled." : "Enabled.", severity: "success" });
    } catch (err) {
      console.error("Toggle enabled failed", err);
      setToast({ open: true, message: "Failed to update application.", severity: "error" });
    } finally {
      setUpdatingAppId(null);
    }
  }

  useEffect(() => {
    if (appsError) {
      console.error("Applications query failed", appsError);
      setToast({ open: true, message: "Failed to load applications.", severity: "error" });
    }
  }, [appsError]);

  const handleToastClose = () => setToast((prev) => ({ ...prev, open: false }));

  return (
    <Stack spacing={3}>
      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Card elevation={1}>
            <CardHeader
              title="Add Application Rule"
              subheader="Define rules using reference-compatible identifiers. Assign them to groups or users from the detail page."
            />
            {isSubmitting && <LinearProgress />}
            <CardContent>
              <form onSubmit={handleSubmit(handleCreateApp)}>
                <Stack spacing={2.5}>
                  <TextField
                    label="Application Name"
                    placeholder="Santa"
                    {...register("name")}
                    error={!!errors.name}
                    helperText={errors.name?.message}
                    autoComplete="off"
                    fullWidth
                    required
                  />

                  <Grid container spacing={2}>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <FormControl fullWidth error={!!errors.rule_type}>
                        <InputLabel id="rule-type-label">Rule Type</InputLabel>
                        <Controller
                          name="rule_type"
                          control={control}
                          render={({ field }) => (
                            <Select labelId="rule-type-label" label="Rule Type" value={field.value} onChange={field.onChange}>
                              {applicationRuleTypes.map((type) => (
                                <MenuItem key={type} value={type}>
                                  {type}
                                </MenuItem>
                              ))}
                            </Select>
                          )}
                        />
                        {errors.rule_type && (
                          <Typography variant="caption" color="error">
                            {errors.rule_type.message}
                          </Typography>
                        )}
                      </FormControl>
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <TextField
                        label="Identifier"
                        placeholder={identifierPlaceholder}
                        spellCheck={false}
                        autoComplete="off"
                        {...register("identifier")}
                        error={!!errors.identifier}
                        helperText={errors.identifier?.message}
                        fullWidth
                        required
                      />
                    </Grid>
                  </Grid>

                  <TextField
                    label="Description"
                    placeholder="Explain why this rule exists or what it covers..."
                    {...register("description")}
                    error={!!errors.description}
                    helperText={errors.description?.message}
                    multiline
                    minRows={2}
                    fullWidth
                  />

                  <Button type="submit" variant="contained" loading={isSubmitting}>
                    Create Application Rule
                  </Button>
                </Stack>
              </form>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <Card elevation={1}>
            <CardHeader
              title="Field Reference Guide"
              action={
                <Tooltip title="Binary Authorisation Help">
                  <Button
                    variant="outlined"
                    onClick={() => window.open("https://northpole.dev/features/binary-authorization/", "_blank", "noopener,noreferrer")}
                    startIcon={<HelpIcon />}
                  >
                    Help
                  </Button>
                </Tooltip>
              }
            />
            <CardContent>
              <Typography variant="body2" gutterBottom>
                Use <code>santactl fileinfo /path/to/app</code> to get these values:
              </Typography>

              <Stack spacing={1.5}>
                {primaryRuleTypeEntries.map(({ type, meta }) => (
                  <Stack key={type} direction="row" spacing={1} alignItems="center">
                    <Typography variant="subtitle2">{meta.label}</Typography>
                    <Tooltip title={meta.description} arrow placement="top">
                      <Chip variant={watchedRuleType === type ? "filled" : "outlined"} label={meta.example} />
                    </Tooltip>
                  </Stack>
                ))}

                <Divider />

                <Typography variant="subtitle2">Signing Chain:</Typography>
                <Stack direction="row" spacing={1} alignItems="center">
                  <Typography variant="body2">1. SHA-256</Typography>
                  <Tooltip title={signingChainReference.description} arrow placement="top">
                    <Chip variant={watchedRuleType === "CERTIFICATE" ? "filled" : "outlined"} label={signingChainReference.example} />
                  </Tooltip>
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12 }}>
          <Card elevation={1}>
            <CardHeader title="Application Rules & Assignments" subheader="Manage who can access each application. Click a card to open the detail view." />
            <CardContent>
              <Stack spacing={2}>
                <Stack direction="row" spacing={1.5} alignItems="center" justifyContent="space-between" flexWrap="wrap">
                  <TextField
                    type="search"
                    label="Search applications..."
                    value={appSearch}
                    onChange={(e) => setAppSearch(e.target.value)}
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

                  <Typography variant="body2" color="text.secondary">
                    Showing {apps.length} application{apps.length === 1 ? "" : "s"} · {totalScopes} total assignments
                  </Typography>
                </Stack>

                {apps.length === 0 ? (
                  <Stack alignItems="center" spacing={1}>
                    <Typography variant="h6" gutterBottom>
                      {hasSearchTerm ? "No matching applications" : "No application rules yet"}
                    </Typography>
                    <Typography color="text.secondary" align="center">
                      {hasSearchTerm
                        ? `We couldn't find any applications matching “${trimmedSearch}”. Try a different search or clear the filter.`
                        : "Create your first application rule above to get started."}
                    </Typography>
                  </Stack>
                ) : (
                  <Grid container spacing={2}>
                    {apps.map((app) => {
                      const stats = getAssignmentStats(app);
                      const allowScopes = app.assignment_stats?.allow_scopes ?? 0;
                      const blockScopes = app.assignment_stats?.block_scopes ?? 0;

                      return (
                        <Grid key={app.id} size={{ xs: 12, sm: 6, md: 4, lg: 3 }}>
                          <Card elevation={2} sx={{ opacity: app.enabled ? 1 : 0.6 }}>
                            <CardActionArea component={RouterLink} to={`/applications/${app.id}`} focusRipple>
                              <CardContent>
                                <Stack direction="row" spacing={1} alignItems="center">
                                  <Avatar sx={{ bgcolor: app.enabled ? "success.main" : "grey.400" }} aria-hidden>
                                    {/* TODO: Add custom icons */}
                                    <ShieldIcon fontSize="small" />
                                  </Avatar>
                                  <Typography variant="subtitle1" noWrap title={app.name}>
                                    {app.name}
                                  </Typography>
                                  <Chip size="small" variant="outlined" label={app.rule_type} />
                                </Stack>

                                <Typography variant="caption" color="text.secondary" title={app.identifier}>
                                  {app.identifier}
                                </Typography>

                                {app.description && (
                                  <Typography variant="body2" color="text.secondary">
                                    {app.description}
                                  </Typography>
                                )}
                              </CardContent>

                              <Divider />

                              <CardActions disableSpacing sx={{ justifyContent: "space-between" }}>
                                <Stack direction="row" spacing={1} alignItems="center">
                                  <Chip size="small" color="success" label={`Allow ${allowScopes}`} />
                                  <Chip size="small" color="error" label={`Block ${blockScopes}`} />
                                  <Chip size="small" variant="outlined" label={`Total ${stats.totalUsersCovered}`} />
                                </Stack>

                                <Stack direction="row" spacing={0.5}>
                                  <IconButton
                                    size="small"
                                    aria-label={app.enabled ? "Pause application" : "Play application"}
                                    onClick={(e) => {
                                      e.preventDefault();
                                      e.stopPropagation();
                                      handleToggleEnabled(app.id, app.enabled);
                                    }}
                                    disabled={updatingAppId === app.id}
                                  >
                                    {app.enabled ? <PauseIcon fontSize="small" /> : <PlayArrowIcon fontSize="small" />}
                                  </IconButton>

                                  <IconButton
                                    size="small"
                                    aria-label="Delete application"
                                    onClick={(e) => {
                                      e.preventDefault();
                                      e.stopPropagation();
                                      requestDeleteApplication(app.id, app.name);
                                    }}
                                    disabled={deletingAppId === app.id}
                                    color="error"
                                  >
                                    <DeleteIcon fontSize="small" />
                                  </IconButton>
                                </Stack>
                              </CardActions>
                            </CardActionArea>
                          </Card>
                        </Grid>
                      );
                    })}
                  </Grid>
                )}
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Dialog open={!!confirmDelete} onClose={() => setConfirmDelete(null)} aria-labelledby="confirm-delete-title">
        <DialogTitle id="confirm-delete-title">Delete Application Rule</DialogTitle>
        <DialogContent dividers>
          <Typography>Are you sure you want to delete “{confirmDelete?.appName}”? This action cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmDelete(null)}>Cancel</Button>
          <Button
            loading={!!(deletingAppId && confirmDelete && deletingAppId === confirmDelete.appId)}
            color="error"
            variant="contained"
            onClick={() => confirmDelete && handleDeleteApplication(confirmDelete.appId)}
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      <PageSnackbar toast={toast} onClose={handleToastClose} />
    </Stack>
  );
}
