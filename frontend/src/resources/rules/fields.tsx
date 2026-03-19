import type { RulePolicy, RuleTargetSubjectKind, RuleType } from "@/api/types";
import { SANTA_CEL_PLAYGROUND_URL } from "@/resources/shared/externalLinks";
import { searchFilterToQuery } from "@/resources/shared/search";
import { ruleIdentifierValidator, trimmedRequired } from "@/resources/shared/validation";
import CodeIcon from "@mui/icons-material/Code";
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormHelperText,
  IconButton,
  Link as MuiLink,
  Typography,
} from "@mui/material";
import { useState, type ReactElement } from "react";
import {
  ArrayInput,
  AutocompleteInput,
  BooleanInput,
  FormDataConsumer,
  ReferenceInput,
  SelectInput,
  SimpleFormIterator,
  TextInput,
  required,
  useSimpleFormIteratorItem,
} from "react-admin";
import { useWatch } from "react-hook-form";

export const RULE_TYPE_CHOICES = [
  { id: "binary", name: "Binary" },
  { id: "certificate", name: "Certificate" },
  { id: "team_id", name: "Team ID" },
  { id: "signing_id", name: "Signing ID" },
  { id: "cd_hash", name: "CD Hash" },
] as { id: RuleType; name: string }[];

export const RULE_POLICY_CHOICES = [
  { id: "allowlist", name: "Allowlist" },
  { id: "blocklist", name: "Blocklist" },
  { id: "silent_blocklist", name: "Silent Blocklist" },
  { id: "cel", name: "CEL" },
] as { id: RulePolicy; name: string }[];

const RULE_TARGET_SUBJECT_KIND_CHOICES = [
  { id: "group", name: "Group" },
  { id: "all_devices", name: "All Devices" },
  { id: "all_users", name: "All Users" },
] as { id: RuleTargetSubjectKind; name: string }[];

const RULE_TYPE_DESCRIPTION: Record<RuleType, string> = {
  binary: "SHA-256 hash of the exact binary.",
  certificate: "SHA-256 hash of the signing certificate.",
  team_id: "10-character Apple Team ID.",
  signing_id: "Signing identifier with team or platform prefix.",
  cd_hash: "Code directory hash of the binary.",
};

const IDENTIFIER_PLACEHOLDER: Record<RuleType, string> = {
  binary: "fc6679da622c3ff38933220b8e73c7322ecdc94b4570c50ecab0da311b292682",
  certificate: "7ae80b9ab38af0c63a9a81765f434d9a7cd8f720eb6037ef303de39d779bc258",
  team_id: "EQHXZ8M8AV",
  signing_id: "UBF8T346G9:com.microsoft.VSCode",
  cd_hash: "dbe8c39801f93e05fc7bc53a02af5b4d3cfc670a",
};

export const RuleDetailsFields = (): ReactElement => (
  <>
    <TextInput
      source="name"
      label="Name"
      placeholder="Visual Studio Code"
      validate={[trimmedRequired("Name")]}
      fullWidth
    />
    <TextInput
      source="description"
      label="Description"
      placeholder="VSCode has been used to bypass terminal restrictions."
      multiline
      minRows={2}
      fullWidth
    />
    <BooleanInput source="enabled" label="Enabled" helperText="Disabled rules are not sent to machines." />
    <Typography variant="body2" color="text">
      These fields are part of the payload sent to machines and are included in rule hash comparison.
    </Typography>
    <FormDataConsumer>
      {({ formData }): ReactElement => {
        const values = formData as { rule_type?: RuleType };
        return (
          <SelectInput
            source="rule_type"
            label="Rule Type"
            choices={RULE_TYPE_CHOICES}
            helperText={values.rule_type ? RULE_TYPE_DESCRIPTION[values.rule_type] : "Choose a rule type."}
            validate={[required()]}
            fullWidth
          />
        );
      }}
    </FormDataConsumer>
    <FormDataConsumer>
      {({ formData }): ReactElement => {
        const values = formData as { rule_type?: RuleType };
        return (
          <TextInput
            source="identifier"
            label="Identifier"
            helperText="Identifier for the selected rule type."
            placeholder={values.rule_type ? IDENTIFIER_PLACEHOLDER[values.rule_type] : ""}
            validate={[trimmedRequired("Identifier"), ruleIdentifierValidator]}
            fullWidth
          />
        );
      }}
    </FormDataConsumer>
    <TextInput
      source="custom_message"
      label="Block Message"
      multiline
      minRows={2}
      helperText="Shown when this rule blocks execution."
      placeholder="This app is not approved. Contact IT."
      fullWidth
    />
    <TextInput
      source="custom_url"
      label="Block Help URL"
      helperText="Help link shown when this rule blocks execution."
      placeholder="https://helpdesk.example.com/software?app=%bundle_or_file_identifier%"
      fullWidth
    />
  </>
);

const CelExpressionInput = ({ source, showRequired }: { source: string; showRequired: boolean }): ReactElement => {
  const [open, setOpen] = useState(false);

  return (
    <Box>
      <IconButton
        size="small"
        color={showRequired ? "error" : "default"}
        onClick={(): void => {
          setOpen(true);
        }}
        aria-label="Edit CEL expression"
      >
        <CodeIcon fontSize="small" />
      </IconButton>
      {showRequired ? <FormHelperText error>CEL required</FormHelperText> : undefined}
      <Dialog
        open={open}
        onClose={(): void => {
          setOpen(false);
        }}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>CEL Expression</DialogTitle>
        <DialogContent>
          <Box mb={1}>
            <MuiLink href={SANTA_CEL_PLAYGROUND_URL} target="_blank" rel="noreferrer" variant="body2">
              Open CEL playground
            </MuiLink>
          </Box>
          <TextInput
            source={source}
            label="CEL expression"
            helperText="Not validated — malformed expressions will be sent to clients."
            validate={[required()]}
            fullWidth
            multiline
            minRows={6}
          />
        </DialogContent>
        <DialogActions>
          <Button
            onClick={(): void => {
              setOpen(false);
            }}
          >
            Done
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

const IncludeTargetCelField = (): ReactElement | undefined => {
  const { index } = useSimpleFormIteratorItem();
  const indexPath = String(index);
  const policy = useWatch({ name: `targets.include.${indexPath}.policy` }) as RulePolicy | undefined;
  const celExpression = useWatch({ name: `targets.include.${indexPath}.cel_expression` }) as string | undefined;
  if (policy !== "cel") return undefined;
  return (
    <CelExpressionInput
      source="cel_expression"
      showRequired={celExpression?.trim() === "" || celExpression === undefined}
    />
  );
};

const SubjectField = (): ReactElement => (
  <FormDataConsumer>
    {({ scopedFormData }): ReactElement => {
      const subjectKind = scopedFormData?.subject_kind as RuleTargetSubjectKind | undefined;
      const isGroup = subjectKind === "group" || subjectKind === undefined;
      return (
        <ReferenceInput reference="groups" source="subject_id">
          <AutocompleteInput
            label="Group"
            optionText="name"
            optionValue="id"
            filterToQuery={searchFilterToQuery}
            validate={isGroup ? [required()] : []}
            disabled={!isGroup}
          />
        </ReferenceInput>
      );
    }}
  </FormDataConsumer>
);

export const RuleTargetsFields = (): ReactElement => (
  <>
    <Typography variant="h6">Include Targets</Typography>
    <Typography variant="body2" color="text.secondary">
      Include target order determines rule priority.
    </Typography>
    <ArrayInput source="targets.include" label={false}>
      <SimpleFormIterator inline>
        <SelectInput
          source="subject_kind"
          label="Target"
          choices={RULE_TARGET_SUBJECT_KIND_CHOICES}
          validate={[required()]}
        />
        <SubjectField />
        <SelectInput source="policy" label="Policy" choices={RULE_POLICY_CHOICES} validate={[required()]} />
        <IncludeTargetCelField />
      </SimpleFormIterator>
    </ArrayInput>
    <Typography variant="h6" sx={{ mt: 2 }}>
      Excluded Groups
    </Typography>
    <Typography variant="body2" color="text.secondary">
      Machines in any excluded group are skipped regardless of include targets.
    </Typography>
    <ArrayInput source="targets.exclude" label={false}>
      <SimpleFormIterator inline>
        <ReferenceInput reference="groups" source="group_id">
          <AutocompleteInput
            label="Group"
            optionText="name"
            optionValue="id"
            filterToQuery={searchFilterToQuery}
            validate={[required()]}
          />
        </ReferenceInput>
      </SimpleFormIterator>
    </ArrayInput>
  </>
);
