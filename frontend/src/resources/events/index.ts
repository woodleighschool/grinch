import EventIcon from "@mui/icons-material/Event";
import { EventList } from "@/resources/events/list";
import { EventShow } from "@/resources/events/show";
import type { ComponentType } from "react";
import type { ResourceProps } from "react-admin";

const events: Partial<ResourceProps> & { icon?: ComponentType } = {
  icon: EventIcon,
  list: EventList,
  show: EventShow,
};

export default events;
