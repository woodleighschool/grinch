import type { ReactElement } from "react";
import { Create, TabbedForm } from "react-admin";
import {
  PolicyDetailsFields,
  PolicyRulesInput,
  PolicySettingsFields,
  PolicyTargetsInput,
} from "@/resources/policies/fields";
import { CLIENT_MODE, FILE_ACCESS_ACTION } from "@/api/constants";

const defaultPolicyValues = {
  enabled: true,
  set_client_mode: CLIENT_MODE.MONITOR,
  set_batch_size: 50,
  set_full_sync_interval_seconds: 600,
  set_push_notification_full_sync_interval_seconds: 14_400,
  set_push_notification_global_rule_sync_deadline_seconds: 600,
  set_enable_bundles: false,
  set_enable_transitive_rules: false,
  set_enable_all_event_upload: false,
  set_disable_unknown_event_upload: false,
  set_allowed_path_regex: "",
  set_blocked_path_regex: "",
  set_block_usb_mount: false,
  set_remount_usb_mode: [],
  set_override_file_access_action: FILE_ACCESS_ACTION.NO_OVERRIDE,
};

export const PolicyCreate = (): ReactElement => (
  <Create redirect="edit">
    <TabbedForm defaultValues={defaultPolicyValues}>
      <TabbedForm.Tab label="Details">
        <PolicyDetailsFields />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Settings">
        <PolicySettingsFields />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Rules">
        <PolicyRulesInput />
      </TabbedForm.Tab>
      <TabbedForm.Tab label="Targets">
        <PolicyTargetsInput />
      </TabbedForm.Tab>
    </TabbedForm>
  </Create>
);
