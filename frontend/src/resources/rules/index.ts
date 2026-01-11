import { RuleCreate } from "@/resources/rules/create";
import { RuleEdit } from "@/resources/rules/edit";
import { RuleList } from "@/resources/rules/list";
import RuleIcon from "@mui/icons-material/Rule";
import type { ComponentType } from "react";
import type { ResourceProps } from "react-admin";

const rules: Partial<ResourceProps> & { icon?: ComponentType } = {
  icon: RuleIcon,
  options: { label: "Rules" },
  recordRepresentation: "name",
  list: RuleList,
  create: RuleCreate,
  edit: RuleEdit,
};

export default rules;
