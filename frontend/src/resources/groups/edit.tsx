import type { components } from "@/api/openapi";
import { GroupMembersTab } from "@/resources/groups/membersTab";
import { ShowActions } from "@/resources/shared/actions";
import type { ReactElement } from "react";
import { DateField, Edit, Labeled, NumberField, TabbedForm, TextInput, useRecordContext } from "react-admin";

type Group = components["schemas"]["Group"];

const GroupOverviewTab = (): ReactElement | undefined => {
  const record = useRecordContext<Pick<Group, "source">>();
  if (!record) return undefined;
  const isLocal = record.source === "local";

  return (
    <>
      <TextInput source="name" label="Name" fullWidth disabled={!isLocal} />
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

export const GroupEdit = (): ReactElement => (
  <Edit mutationMode="pessimistic" redirect="edit" actions={<ShowActions />}>
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
