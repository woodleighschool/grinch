import { trimmedRequired } from "@/resources/shared/validation";
import type { ReactElement } from "react";
import { Create, ListButton, SimpleForm, TextInput, TopToolbar } from "react-admin";

const GroupCreateActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const GroupCreate = (): ReactElement => (
  <Create redirect="edit" actions={<GroupCreateActions />}>
    <SimpleForm>
      <TextInput source="name" label="Name" validate={[trimmedRequired("Name")]} />
      <TextInput source="description" label="Description" multiline minRows={2} />
    </SimpleForm>
  </Create>
);
