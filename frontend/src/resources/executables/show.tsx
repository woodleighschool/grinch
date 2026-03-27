import { ShowActions } from "@/resources/shared/actions";
import { ExecutableEntitlementsArrayField, SigningChainArrayField } from "@/resources/shared/executableFields";
import type { ReactElement } from "react";
import { DateField, Labeled, NumberField, Show, TabbedShowLayout, TextField } from "react-admin";

export const ExecutableShow = (): ReactElement => (
  <Show actions={<ShowActions />}>
    <TabbedShowLayout>
      <TabbedShowLayout.Tab label="Overview">
        <TextField source="file_name" label="File Name" />
        <TextField source="file_sha256" label="SHA-256" />
        <TextField source="cdhash" label="CDHash" />
        <TextField source="file_bundle_id" label="Bundle ID" />
        <TextField source="file_bundle_path" label="Bundle Path" />
        <TextField source="signing_id" label="Signing ID" />
        <TextField source="team_id" label="Team ID" />
        <Labeled label="Occurrences">
          <NumberField source="occurrences" />
        </Labeled>
        <DateField source="created_at" label="Created" showTime />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Signing">
        <SigningChainArrayField source="signing_chain" />
      </TabbedShowLayout.Tab>
      <TabbedShowLayout.Tab label="Entitlements">
        <ExecutableEntitlementsArrayField />
      </TabbedShowLayout.Tab>
    </TabbedShowLayout>
  </Show>
);
