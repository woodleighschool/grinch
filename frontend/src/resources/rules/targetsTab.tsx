import type { Rule, RulePolicy, RuleTarget } from "@/api/types";
import { getErrorMessage } from "@/resources/shared/errors";
import { SANTA_CEL_PLAYGROUND_URL, SANTA_RULE_POLICY_DOCS } from "@/resources/shared/externalLinks";
import { searchFilterToQuery } from "@/resources/shared/search";
import AddIcon from "@mui/icons-material/Add";
import ArrowDownwardIcon from "@mui/icons-material/ArrowDownward";
import ArrowUpwardIcon from "@mui/icons-material/ArrowUpward";
import DescriptionOutlinedIcon from "@mui/icons-material/DescriptionOutlined";
import {
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Link as MuiLink,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import type { ReactElement } from "react";
import { useMemo, useState } from "react";
import {
  AutocompleteInput,
  Datagrid,
  Form,
  FormDataConsumer,
  FunctionField,
  List,
  ReferenceField,
  ReferenceInput,
  SelectInput,
  TextField,
  TextInput,
  required,
  useCreate,
  useDataProvider,
  useGetOne,
  useListContext,
  useNotify,
  useRecordContext,
  useRefresh,
  useUpdate,
} from "react-admin";

type Assignment = "include" | "exclude";

interface RuleTargetEditorValues {
  subject_id?: string;
  policy?: RulePolicy;
  cel_expression?: string;
}

interface RuleTargetDialogProperties {
  assignment: Assignment;
  nextPriority: number;
  open: boolean;
  onClose: () => void;
  ruleID: string;
  targetID?: string | undefined;
}

interface RuleTargetsSectionProperties {
  assignment: Assignment;
  ruleID: string;
}

interface RuleTargetsGridProperties {
  assignment: Assignment;
  dialogOpen: boolean;
  editingTargetID?: string | undefined;
  onClose: () => void;
  onEdit: (targetID: string) => void;
  pendingMoveID?: string | undefined;
  ruleID: string;
}

const RULE_POLICY_CHOICES: { id: RulePolicy; name: string }[] = [
  { id: "allowlist", name: "Allowlist" },
  { id: "blocklist", name: "Blocklist" },
  { id: "silent_blocklist", name: "Silent Blocklist" },
  { id: "cel", name: "CEL" },
];

const getPolicyLabel = (policy?: RulePolicy): string =>
  RULE_POLICY_CHOICES.find((choice): boolean => choice.id === policy)?.name ?? "";

const getDefaultValues = (assignment: Assignment, target?: RuleTarget): RuleTargetEditorValues => {
  if (target) {
    return {
      subject_id: target.subject_id,
      ...(target.policy === undefined ? {} : { policy: target.policy }),
      ...(target.cel_expression === undefined ? {} : { cel_expression: target.cel_expression }),
    };
  }

  if (assignment === "include") {
    return { policy: "blocklist", cel_expression: "" };
  }

  return {};
};

const buildTargetData = ({
  assignment,
  ruleID,
  nextPriority,
  values,
}: {
  assignment: Assignment;
  ruleID?: string;
  nextPriority?: number;
  values: RuleTargetEditorValues;
}): Record<string, unknown> => ({
  ...(ruleID ? { rule_id: ruleID } : {}),
  subject_id: values.subject_id,
  assignment,
  ...(assignment === "include"
    ? {
        ...(nextPriority === undefined ? {} : { priority: nextPriority }),
        policy: values.policy ?? "blocklist",
        cel_expression: values.policy === "cel" ? (values.cel_expression ?? "").trim() : "",
      }
    : {}),
});

const SectionActions = ({ assignment, onAdd }: { assignment: Assignment; onAdd: () => void }): ReactElement => (
  <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ width: "100%", px: 0.5 }}>
    <Typography variant="h6">{assignment === "include" ? "Include" : "Exclude"}</Typography>
    <Tooltip title="Add">
      <IconButton onClick={onAdd} aria-label="Add">
        <AddIcon />
      </IconButton>
    </Tooltip>
  </Stack>
);

const RuleTargetDialog = ({
  assignment,
  nextPriority,
  open,
  onClose,
  ruleID,
  targetID,
}: RuleTargetDialogProperties): ReactElement => {
  const notify = useNotify();
  const refresh = useRefresh();
  const [create, { isPending: isCreating }] = useCreate();
  const [update, { isPending: isUpdating }] = useUpdate();
  const isEditing = Boolean(targetID);
  const isSubmitting = isCreating || isUpdating;

  const { data: target, isPending: isLoadingTarget } = useGetOne<RuleTarget>(
    "rule-targets",
    { id: targetID ?? "" },
    { enabled: open && isEditing },
  );

  const defaultValues = useMemo(
    (): RuleTargetEditorValues => getDefaultValues(assignment, target),
    [assignment, target],
  );

  const handleSubmit = async (values: RuleTargetEditorValues): Promise<void> => {
    if (!values.subject_id) {
      return;
    }

    try {
      await (isEditing && targetID
        ? update("rule-targets", {
            id: targetID,
            data: buildTargetData({ assignment, values }),
            previousData: (target ?? { id: targetID }) as RuleTarget,
          })
        : create("rule-targets", {
            data: buildTargetData({ assignment, ruleID, nextPriority, values }),
          }));

      refresh();
      onClose();
    } catch (error) {
      notify(getErrorMessage(error, isEditing ? "Failed to update target" : "Failed to add target"), {
        type: "error",
      });
    }
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>{isEditing ? "Edit target" : "Add target"}</DialogTitle>

      {isEditing && isLoadingTarget ? (
        <DialogContent>
          <Box sx={{ display: "flex", justifyContent: "center", py: 4 }}>
            <CircularProgress size={24} />
          </Box>
        </DialogContent>
      ) : (
        <Form key={targetID ?? `new-${assignment}`} defaultValues={defaultValues} onSubmit={handleSubmit}>
          <DialogContent sx={{ display: "grid", gap: 2 }}>
            <ReferenceInput reference="groups" source="subject_id">
              <AutocompleteInput
                label="Group"
                optionText="name"
                optionValue="id"
                filterToQuery={searchFilterToQuery}
                validate={required()}
                disabled={isSubmitting}
                fullWidth
              />
            </ReferenceInput>

            {assignment === "include" ? (
              <>
                <FormDataConsumer>
                  {({ formData }): ReactElement => {
                    const policy = (formData.policy as RulePolicy | undefined) ?? "blocklist";

                    return (
                      <Stack direction="row" alignItems="center" gap={1}>
                        <SelectInput
                          source="policy"
                          label="Policy"
                          choices={RULE_POLICY_CHOICES}
                          validate={required()}
                          disabled={isSubmitting}
                          fullWidth
                          sx={{ flex: 1 }}
                        />

                        <Tooltip title="Open third-party documentation">
                          <IconButton
                            href={SANTA_RULE_POLICY_DOCS[policy]}
                            target="_blank"
                            rel="noreferrer"
                            color="primary"
                          >
                            <DescriptionOutlinedIcon />
                          </IconButton>
                        </Tooltip>
                      </Stack>
                    );
                  }}
                </FormDataConsumer>

                <FormDataConsumer>
                  {({ formData }): ReactElement | undefined =>
                    formData.policy === "cel" ? (
                      <Stack spacing={1}>
                        <Stack direction="row" alignItems="center" justifyContent="space-between">
                          <Typography variant="body2" color="text.secondary">
                            CEL expression
                          </Typography>
                          <MuiLink href={SANTA_CEL_PLAYGROUND_URL} target="_blank" rel="noreferrer">
                            Open CEL playground
                          </MuiLink>
                        </Stack>

                        <TextInput
                          source="cel_expression"
                          label={false}
                          helperText="This is not validated, malformed expressions will be sent to clients."
                          validate={required()}
                          disabled={isSubmitting}
                          fullWidth
                          multiline
                          minRows={4}
                        />
                      </Stack>
                    ) : undefined
                  }
                </FormDataConsumer>
              </>
            ) : undefined}
          </DialogContent>

          <DialogActions>
            <Button onClick={onClose} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button type="submit" variant="contained" disabled={isSubmitting}>
              {isEditing ? "Save" : "Add"}
            </Button>
          </DialogActions>
        </Form>
      )}
    </Dialog>
  );
};

const RuleTargetRowActions = ({
  assignment,
  disableReorder,
  onMove,
  pendingMoveID,
}: {
  assignment: Assignment;
  disableReorder: boolean;
  onMove: (target: RuleTarget, direction: -1 | 1) => Promise<void>;
  pendingMoveID?: string | undefined;
}): ReactElement => {
  const record = useRecordContext<RuleTarget>();
  const { data = [] } = useListContext<RuleTarget>();

  if (!record) {
    return <></>;
  }

  const index = data.findIndex((item): boolean => item.id === record.id);
  const isMoving = pendingMoveID !== undefined;
  const canMoveUp = assignment === "include" && index > 0 && !disableReorder;
  const canMoveDown = assignment === "include" && index >= 0 && index < data.length - 1 && !disableReorder;

  return (
    <Stack direction="row" spacing={0.5} justifyContent="flex-end">
      {assignment === "include" ? (
        <>
          <IconButton
            size="small"
            disabled={!canMoveUp || isMoving}
            onClick={(): void => {
              onMove(record, -1).catch((): void => undefined);
            }}
          >
            <ArrowUpwardIcon fontSize="inherit" />
          </IconButton>

          <IconButton
            size="small"
            disabled={!canMoveDown || isMoving}
            onClick={(): void => {
              onMove(record, 1).catch((): void => undefined);
            }}
          >
            <ArrowDownwardIcon fontSize="inherit" />
          </IconButton>
        </>
      ) : undefined}
    </Stack>
  );
};

const RuleTargetsGrid = ({
  assignment,
  dialogOpen,
  editingTargetID,
  onClose,
  onEdit,
  pendingMoveID,
  ruleID,
}: RuleTargetsGridProperties): ReactElement => {
  const dataProvider = useDataProvider();
  const notify = useNotify();
  const refresh = useRefresh();
  const { data = [], error, isPending } = useListContext<RuleTarget>();

  const nextPriority =
    assignment === "include" ? Math.max(0, ...data.map((target): number => target.priority ?? 0)) + 1 : 0;

  const moveTarget = async (target: RuleTarget, direction: -1 | 1): Promise<void> => {
    if (assignment !== "include") {
      return;
    }

    const index = data.findIndex((item): boolean => item.id === target.id);
    const nextIndex = index + direction;

    if (index < 0 || nextIndex < 0 || nextIndex >= data.length) {
      return;
    }

    const reordered = [...data];
    const [moved] = reordered.splice(index, 1);

    if (!moved) {
      return;
    }

    reordered.splice(nextIndex, 0, moved);

    try {
      await Promise.all(
        reordered.flatMap((item, position): Promise<unknown>[] => {
          const priority = position + 1;

          return item.priority === priority
            ? []
            : [
                dataProvider.update("rule-targets", {
                  id: item.id,
                  data: { priority },
                  previousData: item,
                }),
              ];
        }),
      );

      refresh();
    } catch (error_) {
      notify(getErrorMessage(error_, "Failed to reorder include groups"), { type: "error" });
    }
  };

  let grid: ReactElement;

  if (isPending) {
    grid = (
      <Box sx={{ display: "flex", justifyContent: "center", py: 3 }}>
        <CircularProgress size={24} />
      </Box>
    );
  } else if (error) {
    grid = <Typography color="error">{getErrorMessage(error, "Failed to load targets")}</Typography>;
  } else if (data.length === 0) {
    grid = <Typography color="text.secondary">No targets.</Typography>;
  } else {
    grid = (
      <Datagrid
        bulkActionButtons
        rowClick={(id): false => {
          onEdit(String(id));
          return false;
        }}
      >
        <ReferenceField source="subject_id" reference="groups" label="Group">
          <TextField source="name" />
        </ReferenceField>

        {assignment === "include" ? (
          <FunctionField<RuleTarget>
            source="policy"
            label="Policy"
            render={(record): ReactElement => (
              <Chip size="small" variant="outlined" label={getPolicyLabel(record.policy)} />
            )}
          />
        ) : undefined}

        {assignment === "include" ? (
          <FunctionField<RuleTarget> label="Priority" render={(record): string => String(record.priority ?? 0)} />
        ) : undefined}

        <FunctionField<RuleTarget>
          label=""
          render={(): ReactElement => (
            <RuleTargetRowActions
              assignment={assignment}
              disableReorder={data.length < 2}
              onMove={moveTarget}
              pendingMoveID={pendingMoveID}
            />
          )}
        />
      </Datagrid>
    );
  }

  return (
    <>
      {grid}
      <RuleTargetDialog
        assignment={assignment}
        nextPriority={nextPriority}
        open={dialogOpen}
        onClose={onClose}
        ruleID={ruleID}
        targetID={editingTargetID}
      />
    </>
  );
};

const RuleTargetsSection = ({ assignment, ruleID }: RuleTargetsSectionProperties): ReactElement => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingTargetID, setEditingTargetID] = useState<string>();

  const closeDialog = (): void => {
    setDialogOpen(false);
    setEditingTargetID(undefined);
  };

  const openCreate = (): void => {
    setEditingTargetID(undefined);
    setDialogOpen(true);
  };

  const openEdit = (targetID: string): void => {
    setEditingTargetID(targetID);
    setDialogOpen(true);
  };

  return (
    <List
      resource="rule-targets"
      filter={{ rule_id: ruleID, assignment }}
      sort={assignment === "include" ? { field: "priority", order: "ASC" } : { field: "subject_name", order: "ASC" }}
      storeKey={`rule-targets-${ruleID}-${assignment}`}
      disableSyncWithLocation
      actions={<SectionActions assignment={assignment} onAdd={openCreate} />}
      title={false}
      empty={false}
    >
      <RuleTargetsGrid
        assignment={assignment}
        dialogOpen={dialogOpen}
        editingTargetID={editingTargetID}
        onClose={closeDialog}
        onEdit={openEdit}
        ruleID={ruleID}
      />
    </List>
  );
};

export const RuleTargetsTab = (): ReactElement => {
  const record = useRecordContext<Pick<Rule, "id">>();

  if (!record?.id) {
    return <></>;
  }

  return (
    <Stack spacing={4} sx={{ width: "100%" }}>
      <RuleTargetsSection assignment="include" ruleID={record.id} />
      <RuleTargetsSection assignment="exclude" ruleID={record.id} />
    </Stack>
  );
};
