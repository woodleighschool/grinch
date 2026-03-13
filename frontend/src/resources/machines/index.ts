import { MachineList } from "@/resources/machines/list";
import { MachineShow } from "@/resources/machines/show";
import ComputerIcon from "@mui/icons-material/Computer";
import type { ResourceProps } from "react-admin";

const machines: Partial<ResourceProps> = {
  icon: ComputerIcon,
  recordRepresentation: "hostname",
  list: MachineList,
  show: MachineShow,
};

export default machines;
