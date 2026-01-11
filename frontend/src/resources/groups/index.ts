import GroupsIcon from "@mui/icons-material/Groups";
import type { ComponentType } from "react";
import type { ResourceProps } from "react-admin";
import { GroupList } from "@/resources/groups/list";
import { GroupShow } from "@/resources/groups/show";

const groups: Partial<ResourceProps> & { icon?: ComponentType } = {
  icon: GroupsIcon,
  recordRepresentation: "display_name",
  list: GroupList,
  show: GroupShow,
};

export default groups;
