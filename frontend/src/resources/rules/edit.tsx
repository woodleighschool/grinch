import { RuleFields } from "@/resources/rules/fields";
import type { ReactElement } from "react";
import { Edit, SimpleForm } from "react-admin";

export const RuleEdit = (): ReactElement => (
  <Edit mutationMode="pessimistic" redirect="edit">
    <SimpleForm>
      <RuleFields />
    </SimpleForm>
  </Edit>
);
