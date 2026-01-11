import GavelIcon from "@mui/icons-material/Gavel";
import type { ComponentType } from "react";
import type { ResourceProps } from "react-admin";
import { PolicyCreate } from "@/resources/policies/create";
import { PolicyEdit } from "@/resources/policies/edit";
import { PolicyList } from "@/resources/policies/list";

const policies: Partial<ResourceProps> & { icon?: ComponentType } = {
  icon: GavelIcon,
  options: { label: "Policies" },
  recordRepresentation: "name",
  list: PolicyList,
  create: PolicyCreate,
  edit: PolicyEdit,
};

export default policies;
