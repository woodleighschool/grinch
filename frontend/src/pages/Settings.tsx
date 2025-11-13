import { useEffect, useState } from "react";
import { type SantaConfig } from "../api";
import { useSantaConfig } from "../hooks/useQueries";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Button,
  Card,
  CardContent,
  CardHeader,
  Grid,
  List,
  ListItem,
  ListItemText,
  Stack,
  TextField,
  Typography,
  LinearProgress,
} from "@mui/material";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import HelpIcon from "@mui/icons-material/Help";
import CakeIcon from "@mui/icons-material/Cake"; // TODO: find better icon
import { PageSnackbar, type PageToast } from "../components";

function SantaConfigPanel({ config }: { config: SantaConfig | null }) {
  if (!config) {
    return <Alert severity="warning">Unable to load configuration XML. Verify the backend is reachable and try again.</Alert>;
  }

  return (
    <Grid container spacing={3}>
      <Grid size={{ xs: 12, md: 7 }}>
        <Stack spacing={2}>
          <Typography color="text.secondary">Deploy this XML via MDM to preconfigure Santa's sync URLs, telemetry, and ownership metadata.</Typography>

          <Stack direction="row" spacing={1.5}>
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
          </Stack>

          <Stack spacing={1}>
            <TextField label="Santa Configuration XML" value={config.xml} multiline minRows={18} fullWidth InputProps={{ readOnly: true }} />
            <Typography variant="caption" color="text.secondary">
              Paste this payload into your MDM profile. Curly-brace <code>{"{{ }}"}</code> placeholders should be expanded by your provider.
            </Typography>
          </Stack>
        </Stack>
      </Grid>

      <Grid size={{ xs: 12, md: 5 }}>
        <Stack spacing={3}>
          <Card elevation={2}>
            <CardHeader title="Deployment checklist" />
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
            <CardHeader title="Template placeholders" />
            <CardContent>
              <List dense>
                <ListItem>
                  <ListItemText
                    primary={
                      <span>
                        Adjust <code>{"{{username}}"}</code> to match your MDM provider's placeholder syntax.
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

export default function Settings() {
  const [expandedPanel, setExpandedPanel] = useState<string | null>("santa");
  const { data: santaConfig, isLoading, error, refetch, isFetching } = useSantaConfig();
  const [toast, setToast] = useState<PageToast>({ open: false, message: "", severity: "error" });

  useEffect(() => {
    if (error) {
      console.error("Santa config load failed", error);
      setToast({ open: true, message: "Failed to load Santa configuration.", severity: "error" });
    }
  }, [error]);
  const handleToastClose = () => setToast((prev) => ({ ...prev, open: false }));

  return (
    <Stack spacing={3}>
      <Card elevation={1}>
        <CardHeader title="Settings" subheader="Configure system defaults and deployment helpers for Santa." />
        <CardContent />
      </Card>

      <Accordion expanded={expandedPanel === "santa"} onChange={(_, isExpanded) => setExpandedPanel(isExpanded ? "santa" : null)}>
        <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="santa-content" id="santa-header">
          <Stack direction="row" spacing={2} alignItems="center">
            <CakeIcon />
            <Stack>
              <Typography fontWeight={600}>Santa Client Configuration</Typography>
              <Typography variant="body2" color="text.secondary">
                Generate configuration XML for Santa clients to deploy via MDM.
              </Typography>
            </Stack>
          </Stack>
        </AccordionSummary>
        <AccordionDetails>
          <Stack spacing={2}>
            {(isLoading || isFetching) && <LinearProgress />}
            <SantaConfigPanel config={santaConfig ?? null} />
            {error && (
              <Stack direction="row" spacing={1}>
                <Button size="small" variant="outlined" onClick={() => void refetch()}>
                  Retry
                </Button>
              </Stack>
            )}
          </Stack>
        </AccordionDetails>
      </Accordion>

      <PageSnackbar toast={toast} onClose={handleToastClose} />
    </Stack>
  );
}
