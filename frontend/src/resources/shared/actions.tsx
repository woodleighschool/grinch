import type { ReactElement } from "react";
import { DeleteButton, ListButton, TopToolbar } from "react-admin";

export const ShowActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
  </TopToolbar>
);

export const EditableShowActions = (): ReactElement => (
  <TopToolbar>
    <ListButton />
    <DeleteButton redirect="list" mutationMode="pessimistic" />
  </TopToolbar>
);
