import type { ReactElement } from "react";
import {
  ArrayInput,
  AutocompleteInput,
  BooleanInput,
  FormDataConsumer,
  NumberInput,
  ReferenceInput,
  SelectArrayInput,
  SelectInput,
  SimpleFormIterator,
  TextInput,
  required,
  useChoicesContext,
} from "react-admin";

import {
  CLIENT_MODE,
  FILE_ACCESS_ACTION,
  FILE_ACCESS_ACTION_CHOICES,
  POLICY,
  RULE_TYPE,
  enumDescription,
  enumName,
} from "@/api/constants";
import type { Policy, PolicyAttachment, PolicyTarget, Rule } from "@/api/types";

export const PolicyDetailsFields = (): ReactElement => (
  <>
    <TextInput source="name" label="Name" validate={[required()]} helperText="Unique policy name." />
    <TextInput source="description" label="Description" multiline minRows={2} />
    <BooleanInput source="enabled" label="Enabled" />
    <NumberInput
      source="priority"
      label="Priority"
      min={0}
      step={1}
      validate={[required()]}
      helperText="Higher numbers take precedence on each machine."
    />
  </>
);

export const PolicySettingsFields = (): ReactElement => (
  <>
    <FormDataConsumer<Partial<Policy>>>
      {({ formData }): ReactElement => (
        <SelectInput
          source="set_client_mode"
          label="Client Mode"
          choices={CLIENT_MODE.choices("MONITOR", "LOCKDOWN", "STANDALONE")}
          validate={[required()]}
          helperText={enumDescription(CLIENT_MODE, formData.set_client_mode)}
        />
      )}
    </FormDataConsumer>

    <BooleanInput
      source="set_enable_bundles"
      label="Bundle Rules and Hashing"
      helperText="Allow bundle rules and bundle hashing on clients."
    />

    <BooleanInput
      source="set_enable_transitive_rules"
      label="Compiler (Transitive) Rules"
      helperText="Treat compiler rules as transitive allow rules. When off, they act as standard allow rules."
    />

    <BooleanInput
      source="set_enable_all_event_upload"
      label="Upload All Execution Events"
      helperText="Upload all execution events, not only would-be blocked ones."
    />

    <BooleanInput
      source="set_disable_unknown_event_upload"
      label="Disable Unknown Event Uploads"
      helperText="Skip uploads for events that would be blocked in Lockdown while in Monitor mode."
    />

    <NumberInput
      source="set_batch_size"
      label="Event Upload Batch Size"
      min={1}
      step={1}
      validate={[required()]}
      helperText="Events per upload request. Default 50."
      placeholder="50"
    />

    <NumberInput
      source="set_full_sync_interval_seconds"
      label="Full Sync Interval"
      min={60}
      step={1}
      validate={[required()]}
      helperText="Seconds between full syncs. Default 600 (10 minutes); minimum 60."
      placeholder="600"
    />

    <NumberInput
      source="set_push_notification_full_sync_interval_seconds"
      label="Push Notification Full Sync Fallback Interval"
      min={60}
      step={1}
      validate={[required()]}
      helperText="Used when push notifications are enabled. Default 14400 (6 hours)."
      placeholder="14400"
    />

    <NumberInput
      source="set_push_notification_global_rule_sync_deadline_seconds"
      label="Push Notification Rule Sync Jitter Window"
      min={0}
      step={1}
      validate={[required()]}
      helperText="After a global rule sync notification, clients wait a random delay up to this many seconds. Default 600."
      placeholder="600"
    />

    <TextInput
      source="set_allowed_path_regex"
      label="Allowed Path Regex"
      helperText="Regular expression for allowed paths."
      placeholder="^/Applications/.*"
    />

    <TextInput
      source="set_blocked_path_regex"
      label="Blocked Path Regex"
      helperText="Regular expression for blocked paths."
      placeholder="^/Volumes/.*"
    />

    <FormDataConsumer<Partial<Policy>>>
      {({ formData }): ReactElement => (
        <SelectInput
          source="set_override_file_access_action"
          label="File Access Override Action"
          choices={FILE_ACCESS_ACTION_CHOICES}
          validate={[required()]}
          helperText={enumDescription(FILE_ACCESS_ACTION, formData.set_override_file_access_action)}
        />
      )}
    </FormDataConsumer>

    <BooleanInput
      source="set_block_usb_mount"
      label="Block USB and SD Mounts"
      helperText="Block USB and SD card mounts."
    />

    <SelectArrayInput
      source="set_remount_usb_mode"
      label="USB Mount Enforced Flags"
      helperText="When USB blocking is on, mounts with these flags are allowed; others are denied, then remounted with these flags."
      choices={[
        { id: "rdonly", name: "Read-Only (rdonly)" },
        { id: "noexec", name: "Disallow Executables (noexec)" },
        { id: "nosuid", name: "Ignore Setuid/Setgid (nosuid)" },
        { id: "nodev", name: "Ignore Device Files (nodev)" },
        { id: "nobrowse", name: "Hide in Finder (nobrowse)" },
        { id: "noowners", name: "Ignore Ownership (noowners)" },
        { id: "async", name: "Async I/O (async)" },
        { id: "-j", name: "Journaled (-j)" },
      ]}
    />
  </>
);

const RuleAssignmentHelperText = (): string => {
  const { selectedChoices = [] } = useChoicesContext<Rule>();
  const rule = selectedChoices[0];
  if (!rule) {
    return "Select a rule to view details.";
  }
  const ruleTypeName = enumName(RULE_TYPE, rule.rule_type) ?? "Unknown rule";
  return `${rule.identifier} - ${ruleTypeName}`;
};

export const PolicyRulesInput = (): ReactElement => (
  <ArrayInput source="attachments">
    <SimpleFormIterator inline reOrderButtons={false}>
      <ReferenceInput source="rule_id" reference="rules" label="Rule">
        <AutocompleteInput
          optionText="name"
          label="Rule"
          fullWidth
          validate={[required()]}
          helperText={<RuleAssignmentHelperText />}
        />
      </ReferenceInput>

      <FormDataConsumer<PolicyAttachment>>
        {({ scopedFormData }): ReactElement => (
          <SelectInput
            source="action"
            label="Rule Action"
            choices={POLICY.choices("ALLOW", "ALLOW_COMPILER", "BLOCK", "BLOCK_SILENTLY", "EVALUATE_EXPRESSION")}
            validate={[required()]}
            helperText={enumDescription(POLICY, scopedFormData?.action)}
          />
        )}
      </FormDataConsumer>

      <FormDataConsumer<PolicyAttachment>>
        {({ scopedFormData }): ReactElement => {
          const isExpression = scopedFormData?.action === POLICY.EVALUATE_EXPRESSION;
          return (
            <TextInput
              source="cel_expr"
              label="CEL Expression"
              multiline
              minRows={3}
              validate={isExpression ? [required()] : []}
              disabled={!isExpression}
            />
          );
        }}
      </FormDataConsumer>
    </SimpleFormIterator>
  </ArrayInput>
);

export const PolicyTargetsInput = (): ReactElement => (
  <ArrayInput source="targets">
    <SimpleFormIterator inline reOrderButtons={false}>
      <SelectInput
        source="kind"
        label="Target Type"
        choices={[
          { id: "all", name: "All Machines" },
          { id: "user", name: "User" },
          { id: "group", name: "Group" },
          { id: "machine", name: "Machine" },
        ]}
        optionText="name"
        optionValue="id"
        validate={[required()]}
      />

      <FormDataConsumer<PolicyTarget>>
        {({ scopedFormData }): ReactElement | undefined => {
          const kind = scopedFormData?.kind;

          if (kind === "user") {
            return (
              <ReferenceInput source="ref_id" reference="users">
                <AutocompleteInput label="User" optionText="display_name" fullWidth validate={[required()]} />
              </ReferenceInput>
            );
          }

          if (kind === "group") {
            return (
              <ReferenceInput source="ref_id" reference="groups">
                <AutocompleteInput label="Group" optionText="display_name" fullWidth validate={[required()]} />
              </ReferenceInput>
            );
          }

          if (kind === "machine") {
            return (
              <ReferenceInput source="ref_id" reference="machines">
                <AutocompleteInput label="Machine" fullWidth validate={[required()]} />
              </ReferenceInput>
            );
          }

          return undefined;
        }}
      </FormDataConsumer>
    </SimpleFormIterator>
  </ArrayInput>
);
