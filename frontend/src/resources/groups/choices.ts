export type MemberKind = "user" | "machine";

export const MEMBER_KIND_CHOICES = [
  { id: "user", name: "User" },
  { id: "machine", name: "Machine" },
] satisfies { id: MemberKind; name: string }[];
