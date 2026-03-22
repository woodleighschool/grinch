import type { components } from "@/api/openapi";
import { ResourceLink } from "@/resources/shared/resourceLinks";
import { SourceField } from "@/resources/shared/sourceField";
import type { ReactElement } from "react";
import { useRecordContext } from "react-admin";

type MembershipListItem = components["schemas"]["MembershipListItem"];
type MembershipGroupRecord = Pick<MembershipListItem, "group">;

export const MembershipGroupLinkField = (): ReactElement | undefined => {
  const membership = useRecordContext<MembershipGroupRecord>();

  if (!membership) {
    return undefined;
  }

  return <ResourceLink resource="groups" id={membership.group.id} label={membership.group.name} />;
};

export const MembershipGroupSourceField = (): ReactElement | undefined => {
  const membership = useRecordContext<MembershipGroupRecord>();

  if (!membership) {
    return undefined;
  }

  return <SourceField source={membership.group.source} />;
};
