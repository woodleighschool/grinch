import { GroupCreate } from "@/resources/groups/create";
import { GroupEdit } from "@/resources/groups/edit";
import { GroupList } from "@/resources/groups/list";
import GroupsIcon from "@mui/icons-material/Groups";
import type { ResourceProps } from "react-admin";

const groups: Partial<ResourceProps> = {
  icon: GroupsIcon,
  recordRepresentation: "name",
  list: GroupList,
  create: GroupCreate,
  edit: GroupEdit,
};

export default groups;
