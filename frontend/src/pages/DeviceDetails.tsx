import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { formatDateTime, formatCompactDateTime } from "../utils/dates";
import type { Device } from "../api";
import { useDeviceDetails } from "../hooks/useQueries";
import { UserSummary } from "./UserDetails";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Chip,
  Grid,
  LinearProgress,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import { PageSnackbar, type PageToast } from "../components";

function PreflightPanel({ preflight }: { preflight: Record<string, any> }) {
  return (
    <Grid container spacing={3}>
      <Grid size={{ xs: 12, md: 7 }}>
        <Stack spacing={2}>
          <Typography color="text.secondary">
            <pre>{JSON.stringify(preflight, null, "\t")}</pre>
          </Typography>
        </Stack>
      </Grid>
    </Grid>
  );
}

function DeviceSummary({ device, ...props }: { device: Device }) {
  return (
    <Stack {...(props as any)} spacing={1.5}>
      <Stack direction="row" spacing={1} flexWrap="wrap">
        {device.clientMode && <Chip variant="outlined" label={`${device.clientMode}`} />}
        {device.lastSeen && <Chip variant="outlined" label={`Last seen on ${formatDateTime(device.lastSeen)}`} />}
        {device.lastPreflightAt && <Chip variant="outlined" label={`Last preflight on ${formatDateTime(device.lastPreflightAt)}`} />}
        {device.lastPreflightPayload?.santa_version && <Chip variant="outlined" label={`Santa version: ${device.lastPreflightPayload.santa_version}`} />}
        {device.cleanSyncRequested && <Chip variant="outlined" label="Clean sync pending" />}
      </Stack>
    </Stack>
  );
}

export default function DeviceDetails() {
  const [expandedPreflightPanel, setExpandedPreflightPanel] = useState<string | null>(null);
  const { deviceId } = useParams<{ deviceId: string }>();
  const navigate = useNavigate();

  const { data, isLoading, error } = useDeviceDetails(deviceId ?? "");
  const details = data ?? null;
  const [toast, setToast] = useState<PageToast>({ open: false, message: "", severity: "error" });

  useEffect(() => {
    if (error) {
      console.error("Failed to load device details", error);
      setToast({ open: true, message: "Failed to load device details.", severity: "error" });
    }
  }, [error]);
  const handleToastClose = () => setToast((prev) => ({ ...prev, open: false }));

  if (!deviceId) {
    return (
      <Card elevation={1}>
        <CardHeader title="Device Details" />
        <CardContent>
          <Stack spacing={2}>
            <Alert severity="error">Missing device identifier.</Alert>
            <Button variant="contained" onClick={() => navigate(-1)} startIcon={<ArrowBackIcon />}>
              Back
            </Button>
          </Stack>
        </CardContent>
      </Card>
    );
  }

  if (!isLoading && !error && !details) {
    return (
      <Card elevation={1}>
        <CardHeader title="Device Details" />
        <CardContent>
          <Stack spacing={2}>
            <Typography color="text.secondary">Device not found.</Typography>
            <Button variant="contained" onClick={() => navigate(-1)} startIcon={<ArrowBackIcon />}>
              Back
            </Button>
          </Stack>
        </CardContent>
      </Card>
    );
  }

  const {
    machine,
    primary_user,
    recent_blocks: events = [],
    policies = [],
  } = details ?? {
    machine: undefined,
    primary_user: undefined,
    recent_blocks: [],
    policies: [],
  };

  return (
    <Stack spacing={3}>
      {isLoading && <LinearProgress />}

      <Stack direction="row" spacing={1} alignItems="center">
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate(-1)}>
          Back
        </Button>
        <Typography variant="h4">Device Details</Typography>
      </Stack>

      <Card elevation={1}>
        <CardHeader
          title={machine ? `${machine.hostname} - ${machine.serial}` : "Loading device..."}
          subheader="View device data, primary user, events, and applied policies."
        />
        <CardContent>{machine ? <DeviceSummary device={machine} /> : <Typography color="text.secondary">Loading device details…</Typography>}</CardContent>
      </Card>

      <Card elevation={1}>
        <CardHeader title="Primary User" />
        <CardContent>
          {primary_user ? (
            <UserSummary component="li" onClick={() => navigate(`/users/${primary_user.id}`)} sx={{ cursor: "pointer" }} user={primary_user} />
          ) : (
            <Typography color="text.secondary">Loading user details...</Typography>
          )}
        </CardContent>
      </Card>

      <Card elevation={1}>
        <CardHeader title="Recent Blocks" subheader="Latest Santa telemetry targeting this device." />
        <CardContent>
          {events.length === 0 ? (
            <Alert severity="info">No recent Santa blocks recorded for this user.</Alert>
          ) : (
            <Stack spacing={2}>
              {events.map((event) => {
                const occurredAt = event.occurredAt ? formatDateTime(event.occurredAt) : "—";
                const processPath = typeof event.payload?.file_name === "string" ? event.payload.file_name : event.kind;

                return (
                  <Card key={event.id} elevation={2}>
                    <CardContent>
                      <Stack spacing={1}>
                        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                          <Typography fontWeight={600}>{processPath}</Typography>
                          {event.kind && <Chip size="small" label={event.kind} color="error" />}
                        </Stack>
                        <Typography variant="body2" color="text.secondary">
                          Host: {event.hostname || "—"} · {occurredAt}
                        </Typography>
                      </Stack>
                    </CardContent>
                  </Card>
                );
              })}
            </Stack>
          )}
        </CardContent>
      </Card>

      <Card elevation={1}>
        <CardHeader title="Policies Applied" subheader="Rules that currently impact this device's primary user." />
        <CardContent>
          {policies.length === 0 ? (
            <Alert severity="info">No policies currently target this user.</Alert>
          ) : (
            <Stack spacing={2}>
              {policies.map((policy) => (
                <Card key={policy.scope_id} elevation={2}>
                  <CardContent>
                    <Stack spacing={1.5}>
                      <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                        <Typography fontWeight={600}>{policy.application_name}</Typography>
                        <Stack direction="row" spacing={1} flexWrap="wrap">
                          <Chip size="small" variant="outlined" label={policy.rule_type} />
                          <Chip size="small" color={policy.action?.toLowerCase() === "allow" ? "success" : "error"} label={policy.action.toUpperCase()} />
                          {policy.via_group && <Chip size="small" variant="outlined" label={`Via group: ${policy.target_name || policy.target_id}`} />}
                        </Stack>
                      </Stack>
                      <Typography variant="body2">
                        Identifier:{" "}
                        <Typography component="code" variant="body2">
                          {policy.identifier}
                        </Typography>
                      </Typography>
                      {policy.target_type === "user" && !policy.via_group && (
                        <Typography variant="body2" color="text.secondary">
                          Applied directly to this user.
                        </Typography>
                      )}
                    </Stack>
                  </CardContent>
                </Card>
              ))}
            </Stack>
          )}
        </CardContent>
      </Card>
      <Card elevation={1}>
        <CardHeader title="Preflight Data" subheader="The last preflight recorded by the device" />
        <CardContent>
          {machine?.lastPreflightPayload ? (
            <Accordion
              elevation={2}
              expanded={expandedPreflightPanel === "preflight"}
              onChange={(_, isExpanded) => setExpandedPreflightPanel(isExpanded ? "preflight" : null)}
            >
              <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="preflight-content" id="preflight-header">
                <Stack direction="row" spacing={2} alignItems="center">
                  <Stack>
                    <Typography fontWeight={600}>Data</Typography>
                  </Stack>
                </Stack>
              </AccordionSummary>
              <AccordionDetails>
                <Stack spacing={2}>
                  {isLoading && <LinearProgress />}
                  <PreflightPanel preflight={machine.lastPreflightPayload} />
                </Stack>
              </AccordionDetails>
            </Accordion>
          ) : (
            <Alert severity="info">No preflight recorded yet</Alert>
          )}
        </CardContent>
      </Card>
      <PageSnackbar toast={toast} onClose={handleToastClose} />
    </Stack>
  );
}
