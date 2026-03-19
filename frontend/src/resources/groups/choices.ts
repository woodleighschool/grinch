import type { components } from "@/api/openapi";

type MemberKind = components["schemas"]["MemberKind"];

export const MEMBER_KIND_CHOICES = [
  { id: "user", name: "User" },
  { id: "machine", name: "Machine" },
] satisfies { id: MemberKind; name: string }[];
