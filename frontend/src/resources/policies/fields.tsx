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
    <TextInput source="name" label="Name" validate={[required()]} helperText="Unique name for this policy." />
    <TextInput source="description" label="Description" multiline minRows={2} />
    <BooleanInput source="enabled" label="Enabled" />
    <NumberInput
      source="priority"
      label="Priority"
      min={0}
      step={1}
      validate={[required()]}
      helperText="Higher numbers take precedence per machine."
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

    <BooleanInput source="set_enable_bundles" label="Enable Bundle Analysis" />
    <BooleanInput source="set_enable_transitive_rules" label="Enable Transitive Rules" />
    <BooleanInput source="set_enable_all_event_upload" label="Upload All Events" />
    <BooleanInput source="set_disable_unknown_event_upload" label="Skip Unknown Events" />

    <NumberInput source="set_batch_size" label="Batch Size" min={1} step={1} validate={[required()]} />
    <NumberInput
      source="set_full_sync_interval_seconds"
      label="Full Sync Interval (seconds)"
      min={60}
      step={1}
      validate={[required()]}
    />
    <NumberInput
      source="set_push_notification_full_sync_interval_seconds"
      label="Push Sync Interval (seconds)"
      min={60}
      step={1}
      validate={[required()]}
    />
    <NumberInput
      source="set_push_notification_global_rule_sync_deadline_seconds"
      label="Global Sync Deadline (seconds)"
      min={0}
      step={1}
      validate={[required()]}
    />

    <TextInput source="set_allowed_path_regex" label="Allowed Paths" placeholder="^/Applications/.*" />
    <TextInput source="set_blocked_path_regex" label="Blocked Paths" placeholder="^/Volumes/.*" />

    <FormDataConsumer<Partial<Policy>>>
      {({ formData }): ReactElement => (
        <SelectInput
          source="set_override_file_access_action"
          label="File Access Override"
          choices={FILE_ACCESS_ACTION_CHOICES}
          validate={[required()]}
          helperText={enumDescription(FILE_ACCESS_ACTION, formData.set_override_file_access_action)}
        />
      )}
    </FormDataConsumer>

    <BooleanInput source="set_block_usb_mount" label="Block USB Mounts" />
    <SelectArrayInput
      source="set_remount_usb_mode"
      label="USB Mount Restrictions"
      choices={[
        { id: "rdonly", name: "Read-only (rdonly)" },
        { id: "noexec", name: "Disallow executables (noexec)" },
        { id: "nosuid", name: "Ignore suid bits (nosuid)" },
        { id: "nobrowse", name: "Hide in Finder (nobrowse)" },
        { id: "noowners", name: "Ignore ownership (noowners)" },
        { id: "nodev", name: "Ignore devices (nodev)" },
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
    return "Select a rule to see details.";
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
            label="Policy"
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
        label="Type"
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
