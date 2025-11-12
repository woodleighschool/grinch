import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { formatDateTime, formatCompactDateTime } from "../utils/dates";
import type { DirectoryUser } from "../api";
import { useUserDetails } from "../hooks/useQueries";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Chip,
  LinearProgress,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Snackbar,
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";

function getDisplayName(user: DirectoryUser): string {
  return user.displayName || user.upn;
}

function UserSummary({ user }: { user: DirectoryUser }) {
  return (
    <Stack spacing={2}>
      <Box>
        <Typography variant="h4" sx={{ fontWeight: 600 }}>
          {getDisplayName(user)}
        </Typography>
        <Typography color="text.secondary">{user.upn}</Typography>
      </Box>
      <Stack direction="row" spacing={1} flexWrap="wrap">
        {user.createdAt && <Chip variant="outlined" label={`Created ${formatDateTime(user.createdAt)}`} />}
        {user.updatedAt && <Chip variant="outlined" label={`Updated ${formatDateTime(user.updatedAt)}`} />}
      </Stack>
    </Stack>
  );
}

export default function UserDetails() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();

  const { data, isLoading, error } = useUserDetails(userId ?? "");
  const details = data ?? null;
  const [toast, setToast] = useState<{ open: boolean; message: string }>({ open: false, message: "" });

  useEffect(() => {
    if (error) {
      console.error("Failed to load user details", error);
      setToast({ open: true, message: "Failed to load user details." });
    }
  }, [error]);

  if (!userId) {
    return (
      <Card elevation={1}>
        <CardHeader title="User Details" />
        <CardContent>
          <Alert severity="error">Missing user identifier.</Alert>
          <Button variant="contained" onClick={() => navigate(-1)} startIcon={<ArrowBackIcon />} sx={{ mt: 2 }}>
            Back
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (!isLoading && !error && !details) {
    return (
      <Card elevation={1}>
        <CardHeader title="User Details" />
        <CardContent>
          <Typography color="text.secondary">User not found.</Typography>
          <Button variant="contained" onClick={() => navigate(-1)} startIcon={<ArrowBackIcon />} sx={{ mt: 2 }}>
            Back
          </Button>
        </CardContent>
      </Card>
    );
  }

  const {
    user,
    groups = [],
    devices = [],
    recent_events: events = [],
    policies = [],
  } = details ?? {
    user: undefined,
    groups: [],
    devices: [],
    recent_events: [],
    policies: [],
  };

  return (
    <Stack spacing={3}>
      {isLoading && <LinearProgress />}

      <Button variant="text" startIcon={<ArrowBackIcon />} onClick={() => navigate(-1)}>
        Back to users
      </Button>

      {user && (
        <Card elevation={1}>
          <CardHeader title="User Overview" subheader="View directory data, devices, events, and applied policies." />
          <CardContent>
            <UserSummary user={user} />
          </CardContent>
        </Card>
      )}

      <Card elevation={1}>
        <CardHeader title="Group Assignments" />
        <CardContent>
          {groups.length === 0 ? (
            <Alert severity="info">This user is not assigned to any groups.</Alert>
          ) : (
            <Stack direction="row" flexWrap="wrap" gap={1}>
              {groups.map((group) => (
                <Chip key={group.id} label={group.displayName} variant="outlined" />
              ))}
            </Stack>
          )}
        </CardContent>
      </Card>

      <Card elevation={1}>
        <CardHeader title="Devices" subheader="Santa agents that have associated telemetry with this user." />
        <CardContent>
          {devices.length === 0 ? (
            <Alert severity="info">No devices have reported this user yet.</Alert>
          ) : (
            <TableContainer component={Paper} elevation={2}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Hostname</TableCell>
                    <TableCell>Serial</TableCell>
                    <TableCell>Machine ID</TableCell>
                    <TableCell>Last Seen</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {devices.map((device) => (
                    <TableRow key={device.id} hover>
                      <TableCell>{device.hostname || "—"}</TableCell>
                      <TableCell>{device.serial || "—"}</TableCell>
                      <TableCell>
                        <Typography component="code" variant="body2">
                          {device.id}
                        </Typography>
                      </TableCell>
                      <TableCell>{device.lastSeen ? formatDateTime(device.lastSeen) : "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </CardContent>
      </Card>

      <Card elevation={1}>
        <CardHeader title="Recent Events" subheader="Latest Santa telemetry targeting this user." />
        <CardContent>
          {events.length === 0 ? (
            <Alert severity="info">No recent Santa events recorded for this user.</Alert>
          ) : (
            <Stack spacing={2}>
              {events.map((event) => (
                <Paper key={event.id} elevation={2} sx={{ p: 2 }}>
                  <Stack spacing={1}>
                    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                      <Typography fontWeight={600}>{event.process_path}</Typography>
                      {event.decision && <Chip size="small" label={event.decision} color="error" />}
                    </Stack>
                    <Typography variant="body2" color="text.secondary">
                      Host: {event.hostname || "—"} · {formatCompactDateTime(event.occurred_at)}
                    </Typography>
                    <Typography variant="body2">
                      Decision: {event.decision || "Unknown"}
                      {event.blocked_reason ? ` · Reason: ${event.blocked_reason}` : ""}
                    </Typography>
                  </Stack>
                </Paper>
              ))}
            </Stack>
          )}
        </CardContent>
      </Card>

      <Card elevation={1}>
        <CardHeader title="Policies Applied" subheader="Rules that currently impact this user." />
        <CardContent>
          {policies.length === 0 ? (
            <Alert severity="info">No policies currently target this user.</Alert>
          ) : (
            <Stack spacing={2}>
              {policies.map((policy) => (
                <Paper key={policy.scope_id} elevation={2} sx={{ p: 2 }}>
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
                </Paper>
              ))}
            </Stack>
          )}
        </CardContent>
      </Card>

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
