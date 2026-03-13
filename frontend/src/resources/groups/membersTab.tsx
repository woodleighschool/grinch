import type { GroupMembership } from "@/api/types";
import { MEMBER_KIND_CHOICES } from "@/resources/groups/choices";
import { getErrorMessage } from "@/resources/shared/errors";
import { ResourceLink } from "@/resources/shared/resourceLinks";
import AddIcon from "@mui/icons-material/Add";
import LaptopMacOutlinedIcon from "@mui/icons-material/LaptopMacOutlined";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import type { ReactElement } from "react";
import { useState } from "react";
import {
  AutocompleteInput,
  DataTable,
  Form,
  FormDataConsumer,
  Pagination,
  ReferenceInput,
  ReferenceManyField,
  SelectInput,
  required,
  useCreate,
  useNotify,
  useRecordContext,
  useRefresh,
} from "react-admin";

type MemberKind = "user" | "machine";

interface GroupRecord {
  id?: string;
  source?: string;
}

interface AddMemberDialogProperties {
  groupID: string;
  open: boolean;
  onClose: () => void;
}

const AddMemberDialog = ({ groupID, open, onClose }: AddMemberDialogProperties): ReactElement => {
  const notify = useNotify();
  const refresh = useRefresh();
  const [create, { isPending }] = useCreate();

  const handleSubmit = async (values: { member_id?: string; member_kind?: MemberKind }): Promise<void> => {
    if (!values.member_kind || !values.member_id) {
      return;
    }

    try {
      await create("group-memberships", {
        data: {
          group_id: groupID,
          member_kind: values.member_kind,
          member_id: values.member_id,
        },
      });

      notify("Member added", { type: "success" });
      refresh();
      onClose();
    } catch (error) {
      notify(getErrorMessage(error, "Failed to add member"), { type: "error" });
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
                    filterToQuery={(searchText: string): { search: string } => ({ search: searchText })}
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

const MemberKindField = (): ReactElement => {
  const membership = useRecordContext<GroupMembership>();

  if (!membership) {
    return <></>;
  }

  const isUser = membership.member.kind === "user";
  const Icon = isUser ? PersonOutlineIcon : LaptopMacOutlinedIcon;

  return (
    <Tooltip title={isUser ? "User" : "Machine"}>
      <Icon fontSize="small" color="action" />
    </Tooltip>
  );
};

const MemberNameField = (): ReactElement => {
  const membership = useRecordContext<GroupMembership>();

  if (!membership) {
    return <></>;
  }

  return (
    <ResourceLink
      resource={membership.member.kind === "user" ? "users" : "machines"}
      id={membership.member.id}
      label={membership.member.name}
    />
  );
};

export const GroupMembersTab = (): ReactElement => {
  const record = useRecordContext<GroupRecord>();
  const [dialogOpen, setDialogOpen] = useState(false);

  if (!record?.id) {
    return <></>;
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

      <ReferenceManyField
        reference="group-memberships"
        target="group_id"
        perPage={25}
        pagination={<Pagination />}
        sort={{ field: "member_name", order: "ASC" }}
      >
        <DataTable bulkActionButtons={canManage}>
          <DataTable.Col source="member.kind" label="Kind">
            <MemberKindField />
          </DataTable.Col>

          <DataTable.Col label="Member">
            <MemberNameField />
          </DataTable.Col>
        </DataTable>
      </ReferenceManyField>

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
