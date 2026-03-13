import { ExecutableList } from "@/resources/executables/list";
import { ExecutableShow } from "@/resources/executables/show";
import DescriptionIcon from "@mui/icons-material/Description";
import type { ResourceProps } from "react-admin";

const executables: Partial<ResourceProps> = {
  icon: DescriptionIcon,
  recordRepresentation: "file_name",
  list: ExecutableList,
  show: ExecutableShow,
};

export default executables;
