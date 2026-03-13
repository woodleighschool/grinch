import { RuleCreate } from "@/resources/rules/create";
import { RuleEdit } from "@/resources/rules/edit";
import { RuleList } from "@/resources/rules/list";
import RuleIcon from "@mui/icons-material/Rule";
import type { ResourceProps } from "react-admin";

const rules: Partial<ResourceProps> = {
  icon: RuleIcon,
  recordRepresentation: "name",
  list: RuleList,
  create: RuleCreate,
  edit: RuleEdit,
};

export default rules;
