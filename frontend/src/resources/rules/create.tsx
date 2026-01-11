import { RULE_TYPE } from "@/api/constants";
import { RuleFields } from "@/resources/rules/fields";
import type { ReactElement } from "react";
import { Create, SimpleForm } from "react-admin";

const defaultRuleValues = {
  rule_type: RULE_TYPE.BINARY,
};

export const RuleCreate = (): ReactElement => (
  <Create redirect="edit">
    <SimpleForm defaultValues={defaultRuleValues}>
      <RuleFields />
    </SimpleForm>
  </Create>
);
