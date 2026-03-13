import { GroupMembersTab } from "@/resources/groups/membersTab";
import { trimmedRequired } from "@/resources/shared/validation";
import type { ReactElement } from "react";
import {
  DateField,
  Edit,
  Labeled,
  ListButton,
  NumberField,
  TabbedForm,
  TextInput,
  TopToolbar,
  useRecordContext,
} from "react-admin";

interface GroupRecord {
  source?: string;
}

const GroupOverviewTab = (): ReactElement => {
  const record = useRecordContext<GroupRecord>();
  if (!record) return <></>;
  const isLocal = record.source === "local";

  return (
    <>
      <TextInput source="name" label="Name" validate={[trimmedRequired("Name")]} fullWidth disabled={!isLocal} />
      <TextInput source="description" label="Description" fullWidth disabled={!isLocal} />
      <Labeled label="Members">
        <NumberField source="member_count" />
      </Labeled>
      <Labeled label="Created">
        <DateField source="created_at" showTime />
      </Labeled>
      <Labeled label="Updated">
        <DateField source="updated_at" showTime />
      </Labeled>
    </>
  );
};

const GroupEditActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const GroupEdit = (): ReactElement => (
  <Edit mutationMode="optimistic" redirect="edit" actions={<GroupEditActions />}>
    <TabbedForm>
      <TabbedForm.Tab label="Overview">
        <GroupOverviewTab />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Members">
        <GroupMembersTab />
      </TabbedForm.Tab>
    </TabbedForm>
  </Edit>
);
