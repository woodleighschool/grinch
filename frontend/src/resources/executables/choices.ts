import type { ExecutableSource } from "@/api/types";

export const EXECUTABLE_SOURCE_CHOICES = [
  { id: "event", name: "Event" },
  { id: "process", name: "Process" },
] satisfies { id: ExecutableSource; name: string }[];
