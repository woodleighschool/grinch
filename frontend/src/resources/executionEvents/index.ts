import { ExecutionEventList } from "@/resources/executionEvents/list";
import { ExecutionEventShow } from "@/resources/executionEvents/show";
import EventIcon from "@mui/icons-material/Event";
import type { ResourceProps } from "react-admin";

const executionEvents: Partial<ResourceProps> = {
  icon: EventIcon,
  options: { label: "Execution Events" },
  recordRepresentation: "name",
  list: ExecutionEventList,
  show: ExecutionEventShow,
};

export default executionEvents;
