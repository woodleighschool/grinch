import type { components } from "@/api/openapi";

type ExecutableSource = components["schemas"]["ExecutableSource"];

export const EXECUTABLE_SOURCE_CHOICES = [
  { id: "event", name: "Event" },
  { id: "process", name: "Process" },
] satisfies { id: ExecutableSource; name: string }[];
