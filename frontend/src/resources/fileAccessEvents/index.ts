import { FileAccessEventList } from "@/resources/fileAccessEvents/list";
import { FileAccessEventShow } from "@/resources/fileAccessEvents/show";
import FolderOpenIcon from "@mui/icons-material/FolderOpen";
import type { ResourceProps } from "react-admin";

const fileAccessEvents: Partial<ResourceProps> = {
  icon: FolderOpenIcon,
  list: FileAccessEventList,
  show: FileAccessEventShow,
};

export default fileAccessEvents;
