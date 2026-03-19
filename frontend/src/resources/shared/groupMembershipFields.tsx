import type { components } from "@/api/openapi";
import { ResourceLink } from "@/resources/shared/resourceLinks";
import { SourceField } from "@/resources/shared/sourceField";
import type { ReactElement } from "react";
import { useRecordContext } from "react-admin";

type GroupMembershipListItem = components["schemas"]["GroupMembershipListItem"];
type GroupMembershipGroupRecord = Pick<GroupMembershipListItem, "group">;

export const GroupMembershipGroupLinkField = (): ReactElement | undefined => {
  const membership = useRecordContext<GroupMembershipGroupRecord>();

  if (!membership) {
    return undefined;
  }

  return <ResourceLink resource="groups" id={membership.group.id} label={membership.group.name} />;
};

export const GroupMembershipGroupSourceField = (): ReactElement | undefined => {
  const membership = useRecordContext<GroupMembershipGroupRecord>();

  if (!membership) {
    return undefined;
  }

  return <SourceField source={membership.group.source} />;
};
