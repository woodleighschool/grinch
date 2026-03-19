import { ShowActions } from "@/resources/shared/actions";
import type { ReactElement } from "react";
import { Create, SimpleForm, TextInput } from "react-admin";

export const GroupCreate = (): ReactElement => (
  <Create mutationMode="pessimistic" redirect="edit" actions={<ShowActions />}>
    <SimpleForm>
      <TextInput source="name" label="Name" />
      <TextInput source="description" label="Description" multiline minRows={2} />
    </SimpleForm>
  </Create>
);
