import { useEffect } from "react";
import { Alert, Button, Card, CardContent, CardHeader, Grid, LinearProgress, List, ListItem, ListItemText, Stack, TextField, Typography } from "@mui/material";
import HelpIcon from "@mui/icons-material/Help";

import { type SantaConfig } from "../api";
import { PageHeader } from "../components";
import { useSantaConfig } from "../hooks/useQueries";
import { useToast } from "../hooks/useToast";

// Props & types
interface SantaConfigPanelProps {
  config?: SantaConfig | null | undefined;
}

// Subcomponents
function SantaConfigPanel({ config }: SantaConfigPanelProps) {
  if (!config) {
    return <Alert severity="warning">Unable to load configuration XML. Verify the backend is reachable and try again.</Alert>;
  }

  return (
    <Grid
      container
      spacing={3}
    >
      <Grid size={{ xs: 12, md: 7 }}>
        <Stack spacing={2}>
          <Typography color="text.secondary">Deploy this XML via MDM to preconfigure Santa&apos;s sync URLs, telemetry, and ownership metadata.</Typography>

          <Button
            variant="outlined"
            startIcon={<HelpIcon />}
            component="a"
            href="https://northpole.dev/configuration/keys/"
            target="_blank"
            rel="noopener noreferrer"
          >
            Configuration Reference
          </Button>

          <Stack spacing={1}>
            <TextField
              label="Santa Configuration XML"
              value={config.xml}
              multiline
              minRows={18}
              fullWidth
              slotProps={{ input: { readOnly: true } }}
            />
            <Typography
              variant="caption"
              color="text.secondary"
            >
              Paste this payload into your MDM profile. Curly-brace <code>{"{{ }}"}</code> placeholders should be expanded by your provider.
            </Typography>
          </Stack>
        </Stack>
      </Grid>

      <Grid size={{ xs: 12, md: 5 }}>
        <Stack spacing={3}>
          <Card elevation={2}>
            <CardHeader title="Deployment Checklist" />
            <CardContent>
              <List dense>
                <ListItem>
                  <ListItemText
                    primary={
                      <span>
                        Deploy the payload as a profile targeting <code>com.northpolesec.santa</code>.
                      </span>
                    }
                  />
                </ListItem>
                <ListItem>
                  <ListItemText primary="Sync server URLs should point at this Grinch instance." />
                </ListItem>
                <ListItem>
                  <ListItemText primary="Defaults keep Santa in Monitor mode; raise enforcement when ready." />
                </ListItem>
              </List>
            </CardContent>
          </Card>

          <Card elevation={2}>
            <CardHeader title="Template Placeholders" />
            <CardContent>
              <List dense>
                <ListItem>
                  <ListItemText
                    primary={
                      <span>
                        Adjust <code>{"{{username}}"}</code> to match your MDM provider&apos;s placeholder syntax.
                      </span>
                    }
                  />
                </ListItem>
              </List>
            </CardContent>
          </Card>
        </Stack>
      </Grid>
    </Grid>
  );
}

// Page component
export default function Settings() {
  const { data: config, isLoading, error } = useSantaConfig();
  const { showToast } = useToast();

  // Effects
  useEffect(() => {
    if (!error) return;

    showToast({
      message: error instanceof Error ? error.message : "Failed to load settings.",
      severity: "error",
    });
  }, [error, showToast]);

  // Render
  return (
    <Stack spacing={3}>
      <PageHeader
        title="Settings"
        subtitle="Configure global settings and MDM profiles."
      />

      <Card elevation={1}>
        <CardHeader
          title="MDM Configuration"
          subheader="Generate a configuration profile for your fleet."
        />
        <CardContent>{isLoading ? <LinearProgress /> : <SantaConfigPanel config={config} />}</CardContent>
      </Card>
    </Stack>
  );
}
