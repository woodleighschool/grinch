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
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import { PageSnackbar, type PageToast } from "../components";

function UserSummary({ user }: { user: DirectoryUser }) {
  return (
    <Stack spacing={1.5}>
      <Typography color="text.secondary">{user.upn}</Typography>
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
  const [toast, setToast] = useState<PageToast>({ open: false, message: "", severity: "error" });

  useEffect(() => {
    if (error) {
      console.error("Failed to load user details", error);
      setToast({ open: true, message: "Failed to load user details.", severity: "error" });
    }
  }, [error]);
  const handleToastClose = () => setToast((prev) => ({ ...prev, open: false }));

  if (!userId) {
    return (
      <Card elevation={1}>
        <CardHeader title="User Details" />
        <CardContent>
          <Stack spacing={2}>
            <Alert severity="error">Missing user identifier.</Alert>
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
        <CardHeader title="User Details" />
        <CardContent>
          <Stack spacing={2}>
            <Typography color="text.secondary">User not found.</Typography>
            <Button variant="contained" onClick={() => navigate(-1)} startIcon={<ArrowBackIcon />}>
              Back
            </Button>
          </Stack>
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

      <Stack direction="row" spacing={1} alignItems="center">
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate(-1)}>
          Back
        </Button>
        <Typography variant="h4">User Details</Typography>
      </Stack>

      <Card elevation={1}>
        <CardHeader title={user?.displayName ?? "Loading user..."} subheader="View directory data, devices, events, and applied policies." />
        <CardContent>{user ? <UserSummary user={user} /> : <Typography color="text.secondary">Loading user details…</Typography>}</CardContent>
      </Card>

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
                <Card key={event.id} elevation={2}>
                  <CardContent>
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
                  </CardContent>
                </Card>
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

      <PageSnackbar toast={toast} onClose={handleToastClose} />
    </Stack>
  );
}
