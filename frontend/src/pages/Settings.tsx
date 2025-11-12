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
  Snackbar,
  LinearProgress,
} from "@mui/material";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import HelpIcon from "@mui/icons-material/Help";
import CakeIcon from "@mui/icons-material/Cake"; // TODO: find better icon

function SantaConfigPanel({ config }: { config: SantaConfig | null }) {
  if (!config) {
    return <Alert severity="warning">Unable to load configuration XML. Verify the backend is reachable and try again.</Alert>;
  }

  return (
    <Grid container spacing={3}>
      <Grid size={{ xs: 12, md: 7 }}>
        <Typography color="text.secondary" sx={{ mb: 2 }}>
          Deploy this XML via MDM to preconfigure Santa's sync URLs, telemetry, and ownership metadata.
        </Typography>

        <Stack direction="row" spacing={1.5} sx={{ mb: 2 }}>
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

        <TextField label="Santa Configuration XML" value={config.xml} multiline minRows={18} fullWidth InputProps={{ readOnly: true }} />
        <Typography variant="caption" color="text.secondary" sx={{ display: "block", mt: 1 }}>
          Paste this payload into your MDM profile. Curly-brace <code>{"{{ }}"}</code> placeholders should be expanded by your provider.
        </Typography>
      </Grid>

      <Grid size={{ xs: 12, md: 5 }}>
        <Card elevation={2} sx={{ mb: 3 }}>
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
      </Grid>
    </Grid>
  );
}

export default function Settings() {
  const [expandedPanel, setExpandedPanel] = useState<string | null>("santa");
  const { data: santaConfig, isLoading, error, refetch, isFetching } = useSantaConfig();
  const [toast, setToast] = useState<{ open: boolean; message: string }>({ open: false, message: "" });

  useEffect(() => {
    if (error) {
      console.error("Santa config load failed", error);
      setToast({ open: true, message: "Failed to load Santa configuration." });
    }
  }, [error]);

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
          {(isLoading || isFetching) && <LinearProgress sx={{ mb: 2 }} />}
          <SantaConfigPanel config={santaConfig ?? null} />
          {error && (
            <Stack direction="row" spacing={1} sx={{ mt: 2 }}>
              <Button size="small" variant="outlined" onClick={() => void refetch()}>
                Retry
              </Button>
            </Stack>
          )}
        </AccordionDetails>
      </Accordion>

      <Snackbar
        open={toast.open}
        autoHideDuration={4000}
        onClose={() => setToast((t) => ({ ...t, open: false }))}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert severity="error" onClose={() => setToast((t) => ({ ...t, open: false }))} variant="filled">
          {toast.message}
        </Alert>
      </Snackbar>
    </Stack>
  );
}
