import { groupsApi } from "@/api/adminClient";
import type { components } from "@/api/openapi";
import { MEMBER_KIND_CHOICES, type MemberKind } from "@/resources/groups/choices";
import { searchFilterToQuery } from "@/resources/shared/search";
import AddIcon from "@mui/icons-material/Add";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import LaptopMacOutlinedIcon from "@mui/icons-material/LaptopMacOutlined";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Link as MuiLink,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Tooltip,
  Typography,
} from "@mui/material";
import type { ReactElement } from "react";
import { useMemo, useState } from "react";
import {
  AutocompleteInput,
  Form,
  FormDataConsumer,
  ReferenceInput,
  SelectInput,
  required,
  useGetMany,
  useNotify,
  useRecordContext,
  useRefresh,
} from "react-admin";
import { Link as RouterLink } from "react-router-dom";

type Group = components["schemas"]["Group"];
type MachineSummary = components["schemas"]["MachineSummary"];
type User = components["schemas"]["User"];

interface AddMemberDialogProperties {
  groupID: string;
  open: boolean;
  onClose: () => void;
}

interface MemberRow {
  id: string;
  kind: MemberKind;
  label?: string;
}

const AddMemberDialog = ({ groupID, open, onClose }: AddMemberDialogProperties): ReactElement => {
  const notify = useNotify();
  const refresh = useRefresh();
  const [isPending, setIsPending] = useState(false);

  const handleSubmit = async (values: { member_id?: string; member_kind?: MemberKind }): Promise<void> => {
    if (!values.member_kind || !values.member_id) {
      return;
    }

    setIsPending(true);
    try {
      await (values.member_kind === "machine"
        ? groupsApi.addMachine(groupID, values.member_id)
        : groupsApi.addUser(groupID, values.member_id));

      notify("Member added", { type: "success" });
      refresh();
      onClose();
    } catch (error) {
      notify(error instanceof Error && error.message.trim() !== "" ? error.message : "Failed to add member", {
        type: "error",
      });
    } finally {
      setIsPending(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Add member</DialogTitle>

      <Form defaultValues={{ member_kind: "user" }} onSubmit={handleSubmit}>
        <DialogContent sx={{ display: "grid", gap: 2 }}>
          <Typography variant="body2" color="text.secondary">
            Add a local user or machine to this group.
          </Typography>

          <SelectInput
            source="member_kind"
            label="Member kind"
            choices={MEMBER_KIND_CHOICES}
            validate={required()}
            disabled={isPending}
            fullWidth
          />

          <FormDataConsumer>
            {({ formData }): ReactElement => {
              const isMachine = formData.member_kind === "machine";

              return (
                <ReferenceInput reference={isMachine ? "machines" : "users"} source="member_id">
                  <AutocompleteInput
                    label={isMachine ? "Machine" : "User"}
                    optionText={isMachine ? "hostname" : "display_name"}
                    optionValue="id"
                    filterToQuery={searchFilterToQuery}
                    validate={required()}
                    disabled={isPending}
                    fullWidth
                  />
                </ReferenceInput>
              );
            }}
          </FormDataConsumer>
        </DialogContent>

        <DialogActions>
          <Button onClick={onClose} disabled={isPending}>
            Cancel
          </Button>
          <Button type="submit" variant="contained" disabled={isPending}>
            Add member
          </Button>
        </DialogActions>
      </Form>
    </Dialog>
  );
};

const GroupMembersTable = ({ canManage }: { canManage: boolean }): ReactElement | undefined => {
  const record = useRecordContext<Group>();
  const notify = useNotify();
  const refresh = useRefresh();
  const [pendingKey, setPendingKey] = useState<string | undefined>();

  const userIDs = record?.user_ids ?? [];
  const machineIDs = record?.machine_ids ?? [];

  const { data: users = [], isPending: usersPending } = useGetMany<User>("users", { ids: userIDs });
  const { data: machines = [], isPending: machinesPending } = useGetMany<MachineSummary>("machines", {
    ids: machineIDs,
  });

  const rows = useMemo((): MemberRow[] => {
    const items = [
      ...users.map((user): MemberRow => ({ id: user.id, kind: "user", label: user.display_name })),
      ...machines.map((machine): MemberRow => ({ id: machine.id, kind: "machine", label: machine.hostname })),
    ];

    return items.toSorted((a, b): number => (a.label ?? "").localeCompare(b.label ?? ""));
  }, [machines, users]);

  if (!record) {
    return undefined;
  }

  if (usersPending || machinesPending) {
    return <Typography color="text.secondary">Loading members...</Typography>;
  }

  if (rows.length === 0) {
    return <Typography color="text.secondary">No members.</Typography>;
  }

  const handleRemove = (row: MemberRow): void => {
    const key = `${row.kind}:${row.id}`;
    setPendingKey(key);

    const remove =
      row.kind === "machine" ? groupsApi.removeMachine(record.id, row.id) : groupsApi.removeUser(record.id, row.id);

    void remove
      .then((): void => {
        notify("Member removed", { type: "success" });
        refresh();
      })
      .catch((error: unknown): void => {
        notify(error instanceof Error && error.message.trim() !== "" ? error.message : "Failed to remove member", {
          type: "error",
        });
      })
      .finally((): void => {
        setPendingKey(undefined);
      });
  };

  return (
    <Table size="small">
      <TableHead>
        <TableRow>
          <TableCell>Kind</TableCell>
          <TableCell>Member</TableCell>
          {canManage ? <TableCell align="right">Actions</TableCell> : undefined}
        </TableRow>
      </TableHead>
      <TableBody>
        {rows.map((row): ReactElement => {
          const isUser = row.kind === "user";
          const Icon = isUser ? PersonOutlineIcon : LaptopMacOutlinedIcon;
          const actionKey = `${row.kind}:${row.id}`;

          return (
            <TableRow key={actionKey}>
              <TableCell>
                <Tooltip title={isUser ? "User" : "Machine"}>
                  <Icon fontSize="small" color="action" />
                </Tooltip>
              </TableCell>
              <TableCell>
                <MuiLink
                  component={RouterLink}
                  to={`/${isUser ? "users" : "machines"}/${row.id}`}
                  underline="hover"
                  color="primary"
                >
                  {row.label}
                </MuiLink>
              </TableCell>
              {canManage ? (
                <TableCell align="right">
                  <IconButton
                    onClick={(): void => {
                      handleRemove(row);
                    }}
                    disabled={pendingKey === actionKey}
                  >
                    <DeleteOutlineIcon />
                  </IconButton>
                </TableCell>
              ) : undefined}
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
};

export const GroupMembersTab = (): ReactElement | undefined => {
  const record = useRecordContext<Pick<Group, "id" | "source">>();
  const [dialogOpen, setDialogOpen] = useState(false);

  if (!record?.id) {
    return undefined;
  }

  const canManage = record.source === "local";

  return (
    <Stack spacing={2} sx={{ width: "100%" }}>
      <Stack direction="row" justifyContent="space-between" alignItems="flex-start" spacing={2}>
        {canManage ? (
          <Tooltip title="Add member">
            <IconButton
              onClick={(): void => {
                setDialogOpen(true);
              }}
            >
              <AddIcon />
            </IconButton>
          </Tooltip>
        ) : undefined}
      </Stack>

      <GroupMembersTable canManage={canManage} />

      {canManage ? (
        <AddMemberDialog
          groupID={record.id}
          open={dialogOpen}
          onClose={(): void => {
            setDialogOpen(false);
          }}
        />
      ) : undefined}
    </Stack>
  );
};
