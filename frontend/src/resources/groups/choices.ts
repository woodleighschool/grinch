import type { MemberKind } from "@/api/types";

export const MEMBER_KIND_CHOICES = [
  { id: "user", name: "User" },
  { id: "machine", name: "Machine" },
] satisfies { id: MemberKind; name: string }[];
