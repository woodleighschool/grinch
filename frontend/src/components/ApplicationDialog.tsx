import { useEffect } from "react";
import { Controller, useForm } from "react-hook-form";
import {
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  LinearProgress,
  TextField,
  Switch,
  FormControlLabel,
  FormHelperText,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Typography,
  Card,
  CardHeader,
  CardContent,
  Tooltip,
  Chip,
  Divider,
} from "@mui/material";
import HelpOutlineIcon from "@mui/icons-material/HelpOutline";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { ApiValidationError, validateApplication, type Application } from "../api";
import { useCreateApplication, useUpdateApplication } from "../hooks/useQueries";
import {
  applicationRuleTypeMetadata,
  applicationRuleTypes,
  defaultApplicationRuleType,
  primaryRuleTypeEntries,
  signingChainReference,
  type ApplicationRuleType,
} from "../constants/applicationRules";

const optionalText = z.string().transform((value) => value.trim());

const applicationSchema = z.object({
  name: z.string().trim().min(1, "Application name is required."),
  rule_type: z.enum(applicationRuleTypes),
  identifier: z.string().trim().min(1, "Identifier is required."),
  description: optionalText,
  block_message: optionalText,
  cel_enabled: z.boolean(),
  cel_expression: optionalText,
});

export type ApplicationFormValues = z.infer<typeof applicationSchema>;
type ApplicationDialogMode = "create" | "edit";

const defaultValues: ApplicationFormValues = {
  name: "",
  rule_type: defaultApplicationRuleType,
  identifier: "",
  description: "",
  block_message: "",
  cel_enabled: false,
  cel_expression: "",
};

interface ApplicationDialogProps {
  open: boolean;
  mode?: ApplicationDialogMode;
  application?: Application | null;
  onClose: () => void;
  onSuccess: () => void;
  onError: (message: string) => void;
}

export function ApplicationDialog({ open, mode = "create", application, onClose, onSuccess, onError }: ApplicationDialogProps) {
  const editing = mode === "edit";
  const createApplication = useCreateApplication();
  const updateApplication = useUpdateApplication();

  const form = useForm<ApplicationFormValues>({
    defaultValues,
    resolver: zodResolver(applicationSchema),
  });

  const {
    register,
    control,
    handleSubmit,
    watch,
    reset,
    setValue,
    setError,
    clearErrors,
    formState: { isSubmitting, errors },
  } = form;

  const getHelperText = (value: unknown) => (typeof value === "string" ? value : undefined);

  useEffect(() => {
    if (!open) return;

    if (editing && application) {
      reset({
        name: application.name,
        rule_type: application.rule_type as ApplicationFormValues["rule_type"],
        identifier: application.identifier,
        description: application.description ?? "",
        block_message: application.block_message ?? "",
        cel_enabled: application.cel_enabled,
        cel_expression: application.cel_expression ?? "",
      });
    } else {
      reset(defaultValues);
    }

    clearErrors();
  }, [open, editing, application, reset, clearErrors]);

  const watchedRuleType = watch("rule_type");
  const watchedCelEnabled = watch("cel_enabled");

  const celLocked = editing && Boolean(application?.cel_enabled);

  useEffect(() => {
    if (celLocked) return;
    if (!watchedCelEnabled) {
      setValue("cel_expression", "");
    }
  }, [watchedCelEnabled, celLocked, setValue]);

  const identifierPlaceholder = applicationRuleTypeMetadata[watchedRuleType].placeholder;
  const dialogTitle = editing ? "Edit Application Rule" : "Create Application Rule";
  const submitLabel = editing ? "Save Changes" : "Create";
  const submittingLabel = editing ? "Saving..." : "Creating...";

  const buildPayload = (values: ApplicationFormValues) => {
    const payload: {
      name: string;
      rule_type: ApplicationRuleType;
      identifier: string;
      cel_enabled: boolean;
      description?: string;
      block_message?: string;
      cel_expression?: string;
    } = {
      name: values.name.trim(),
      rule_type: values.rule_type,
      identifier: values.identifier.trim(),
      cel_enabled: values.cel_enabled,
    };

    if (values.description) payload.description = values.description;
    if (values.block_message) payload.block_message = values.block_message;
    if (values.cel_enabled && values.cel_expression) {
      payload.cel_expression = values.cel_expression;
    }

    return payload;
  };

  const onSubmit = async (formData: ApplicationFormValues) => {
    clearErrors();

    try {
      const payload = buildPayload(formData);

      if (editing) {
        if (!application?.id) {
          throw new Error("Missing application identifier.");
        }
        await updateApplication.mutateAsync({ appId: application.id, payload });
      } else {
        const { normalised } = await validateApplication(payload);
        await createApplication.mutateAsync(normalised);
      }

      onSuccess();
      onClose();
    } catch (err) {
      if (err instanceof ApiValidationError) {
        Object.entries(err.fieldErrors).forEach(([field, message]) => {
          if (
            field === "name" ||
            field === "rule_type" ||
            field === "identifier" ||
            field === "description" ||
            field === "block_message" ||
            field === "cel_enabled" ||
            field === "cel_expression"
          ) {
            setError(field, {
              type: "server",
              message,
            });
          }
        });
        return;
      }

      console.error("Application dialog failed", err);
      onError(editing ? "Failed to update application rule." : "Failed to create application rule.");
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
    >
      <DialogTitle>
        <Stack
          direction="row"
          alignItems="center"
          justifyContent="space-between"
          spacing={2}
        >
          <Typography variant="h6">{dialogTitle}</Typography>

          <Stack
            direction="row"
            spacing={1.5}
            alignItems="center"
          >
            <FormControlLabel
              label={
                <Typography
                  variant="body2"
                  component="span"
                >
                  CEL Mode
                </Typography>
              }
              control={
                <Controller
                  name="cel_enabled"
                  control={control}
                  render={({ field }) => (
                    <Switch
                      {...field}
                      checked={field.value}
                      onChange={(event) => {
                        if (celLocked) return;
                        field.onChange(event.target.checked);
                      }}
                      disabled={celLocked}
                      slotProps={{ input: { "aria-label": "Enable CEL mode" } }}
                    />
                  )}
                />
              }
            />
            <Tooltip
              arrow
              title={
                celLocked
                  ? "CEL mode is permanently enabled for this rule. To disable it, delete and recreate the rule."
                  : "Enable CEL mode to control this rule using a CEL expression. Once created with CEL mode enabled, it cannot be disabled without deleting and recreating the rule."
              }
            >
              <HelpOutlineIcon
                fontSize="small"
                color="action"
                sx={{ cursor: "default" }}
              />
            </Tooltip>
          </Stack>
        </Stack>
      </DialogTitle>

      {isSubmitting && <LinearProgress />}

      <DialogContent dividers>
        <form
          id="application-form"
          onSubmit={(e) => void handleSubmit(onSubmit)(e)}
        >
          <Stack
            direction={{ xs: "column", md: "row" }}
            spacing={3}
            alignItems={{ xs: "stretch", md: "flex-start" }}
          >
            <Stack
              spacing={2.5}
              flex={{ xs: "auto" }}
            >
              <TextField
                label="Application Name"
                placeholder="Santa"
                {...register("name")}
                error={!!errors.name}
                helperText={getHelperText(errors.name?.message)}
                autoComplete="off"
                fullWidth
                required
              />

              <FormControl
                fullWidth
                error={!!errors.rule_type}
              >
                <InputLabel id="rule-type-label">Rule Type</InputLabel>
                <Controller
                  name="rule_type"
                  control={control}
                  render={({ field }) => (
                    <Select
                      labelId="rule-type-label"
                      label="Rule Type"
                      value={field.value}
                      onChange={field.onChange}
                    >
                      {applicationRuleTypes.map((type) => (
                        <MenuItem
                          key={type}
                          value={type}
                        >
                          {type}
                        </MenuItem>
                      ))}
                    </Select>
                  )}
                />
                {errors.rule_type && <FormHelperText>{getHelperText(errors.rule_type.message)}</FormHelperText>}
              </FormControl>

              <TextField
                label="Identifier"
                placeholder={identifierPlaceholder}
                {...register("identifier")}
                error={!!errors.identifier}
                helperText={getHelperText(errors.identifier?.message)}
                spellCheck={false}
                autoComplete="off"
                fullWidth
                required
              />

              <TextField
                label="Description"
                placeholder="Describe what this rule does and why it exists."
                {...register("description")}
                error={!!errors.description}
                helperText={getHelperText(errors.description?.message)}
                multiline
                minRows={2}
                fullWidth
              />

              <TextField
                label="Block Message"
                placeholder="Message shown when this rule blocks an application."
                {...register("block_message")}
                error={!!errors.block_message}
                helperText={getHelperText(errors.block_message?.message)}
                multiline
                minRows={2}
                fullWidth
              />
            </Stack>

            <Card
              elevation={1}
              sx={{ alignSelf: "flex-start" }}
            >
              <CardHeader
                title="Field Reference Guide"
                action={
                  <Tooltip title="Binary Authorization Help">
                    <Button
                      size="small"
                      variant="outlined"
                      startIcon={<HelpOutlineIcon fontSize="small" />}
                      onClick={() => window.open("https://northpole.dev/features/binary-authorization/", "_blank", "noopener,noreferrer")}
                    >
                      Help
                    </Button>
                  </Tooltip>
                }
              />
              <CardContent>
                <Typography
                  variant="body2"
                  color="text.secondary"
                  gutterBottom
                >
                  Run <code>santactl fileinfo /path/to/app</code> to copy the identifiers below.
                </Typography>

                <Stack spacing={1.5}>
                  {primaryRuleTypeEntries.map(({ type, meta }) => (
                    <Stack
                      key={type}
                      direction="row"
                      spacing={1}
                      alignItems="center"
                    >
                      <Typography
                        variant="subtitle2"
                        width={90}
                      >
                        {meta.label}
                      </Typography>
                      <Tooltip
                        title={meta.description}
                        arrow
                        placement="top"
                      >
                        <Chip
                          variant={watchedRuleType === type ? "filled" : "outlined"}
                          label={meta.example}
                          size="small"
                        />
                      </Tooltip>
                    </Stack>
                  ))}

                  <Divider />

                  <Stack spacing={0.5}>
                    <Typography variant="subtitle2">Signing Chain</Typography>
                    <Stack
                      direction="row"
                      spacing={1}
                      alignItems="center"
                    >
                      <Typography variant="body2">SHA-256</Typography>
                      <Tooltip
                        title={signingChainReference.description}
                        arrow
                        placement="top"
                      >
                        <Chip
                          variant={watchedRuleType === "CERTIFICATE" ? "filled" : "outlined"}
                          label={signingChainReference.example}
                          size="small"
                        />
                      </Tooltip>
                    </Stack>
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
          </Stack>

          {watchedCelEnabled && (
            <Stack
              mt={3}
              spacing={1}
            >
              <TextField
                label="CEL expression"
                placeholder={`args.exists(arg, arg in ["--foo"]) ? BLOCKLIST : ALLOWLIST`}
                {...register("cel_expression")}
                error={!!errors.cel_expression}
                helperText={getHelperText(errors.cel_expression?.message) ?? "Define the rule using Common Expression Language (CEL)."}
                multiline
                minRows={4}
                required
                fullWidth
                slotProps={{ input: { style: { fontFamily: "monospace" } } }}
              />
            </Stack>
          )}
        </form>
      </DialogContent>

      <DialogActions>
        <Button
          onClick={onClose}
          disabled={isSubmitting}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          form="application-form"
          variant="contained"
          disabled={isSubmitting}
          startIcon={isSubmitting ? <CircularProgress size={16} /> : undefined}
        >
          {isSubmitting ? submittingLabel : submitLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
