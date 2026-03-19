import type { GroupMembershipListItem } from "@/api/types";
import { ResourceLink } from "@/resources/shared/resourceLinks";
import { SourceField } from "@/resources/shared/sourceField";
import type { ReactElement } from "react";
import { RecordContextProvider, useRecordContext } from "react-admin";

type GroupMembershipGroupRecord = Pick<GroupMembershipListItem, "group">;

export const GroupMembershipGroupLinkField = (): ReactElement => {
  const membership = useRecordContext<GroupMembershipGroupRecord>();

  if (!membership) {
    return <></>;
  }

  return <ResourceLink resource="groups" id={membership.group.id} label={membership.group.name} />;
};

export const GroupMembershipGroupSourceField = (): ReactElement => {
  const membership = useRecordContext<GroupMembershipGroupRecord>();

  if (!membership) {
    return <></>;
  }

  return (
    <RecordContextProvider value={{ source: membership.group.source }}>
      <SourceField />
    </RecordContextProvider>
  );
};
