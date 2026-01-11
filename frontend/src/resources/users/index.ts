import { UserList } from "@/resources/users/list";
import { UserShow } from "@/resources/users/show";
import PeopleIcon from "@mui/icons-material/People";
import type { ComponentType } from "react";
import type { ResourceProps } from "react-admin";

const users: Partial<ResourceProps> & { icon?: ComponentType } = {
  icon: PeopleIcon,
  recordRepresentation: "display_name",
  list: UserList,
  show: UserShow,
};

export default users;
