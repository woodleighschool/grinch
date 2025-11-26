import { useCallback, useEffect, type ReactNode } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Alert,
  Card,
  CardActionArea,
  CardContent,
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
  Typography,
} from "@mui/material";

import { formatDateTime } from "../utils/dates";
import type { Device, DirectoryGroup, UserPolicy } from "../api";
import { useUserDetails } from "../hooks/useQueries";
import { EmptyState, PageHeader, RecentBlocksList, SectionCard, UserSummary } from "../components";
import { useToast } from "../hooks/useToast";

interface GroupAssignmentChipsProps {
  groups: DirectoryGroup[];
}

function GroupAssignmentChips({ groups }: GroupAssignmentChipsProps) {
  if (groups.length === 0) {
    return <Alert severity="info">This user is not assigned to any groups.</Alert>;
  }

  const sortedGroups = [...groups].sort((a, b) => a.displayName.localeCompare(b.displayName));

  return (
    <Stack
      direction="row"
      flexWrap="wrap"
      gap={1}
    >
      {sortedGroups.map((group) => (
        <Chip
          key={group.id}
          label={group.displayName}
          variant="outlined"
          size="small"
        />
      ))}
    </Stack>
  );
}

interface DevicesTableProps {
  devices: Device[];
  onSelectDevice: (deviceId: string) => void;
}

function DevicesTable({ devices, onSelectDevice }: DevicesTableProps) {
  if (devices.length === 0) {
    return <Alert severity="info">No devices have reported this user yet.</Alert>;
  }

  return (
    <TableContainer
      component={Paper}
      elevation={0}
      variant="outlined"
    >
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
            <TableRow
              key={device.id}
              hover
              sx={{ cursor: "pointer" }}
              onClick={() => {
                onSelectDevice(device.id);
              }}
            >
              <TableCell>{device.hostname || "—"}</TableCell>
              <TableCell>{device.serial || "—"}</TableCell>
              <TableCell>
                <Typography
                  component="code"
                  variant="body2"
                >
                  {device.id}
                </Typography>
              </TableCell>
              <TableCell>{device.lastSeen ? formatDateTime(device.lastSeen) : "—"}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

interface PoliciesListProps {
  policies: UserPolicy[];
  onSelectPolicy?: (applicationId: string) => void;
}

function PoliciesList({ policies, onSelectPolicy }: PoliciesListProps) {
  if (policies.length === 0) {
    return <Alert severity="info">No policies currently target this user.</Alert>;
  }

  return (
    <Stack spacing={1.5}>
      {policies.map((policy) => {
        const action = policy.action;
        const actionLower = action.toLowerCase();
        const actionColor = actionLower === "allow" ? "success" : actionLower === "block" ? "error" : "info";
        const canNavigate = Boolean(policy.application_id && onSelectPolicy);

        return (
          <Card
            key={policy.scope_id}
            elevation={0}
            variant="outlined"
          >
            <CardActionArea
              onClick={() => {
                if (policy.application_id && onSelectPolicy) {
                  onSelectPolicy(policy.application_id);
                }
              }}
              disabled={!canNavigate}
              sx={{ textAlign: "left" }}
            >
              <CardContent>
                <Stack spacing={1.25}>
                  <Stack
                    direction="row"
                    spacing={1}
                    alignItems="center"
                    flexWrap="wrap"
                  >
                    <Typography fontWeight={600}>{policy.application_name}</Typography>

                    <Stack
                      direction="row"
                      spacing={1}
                      flexWrap="wrap"
                    >
                      <Chip
                        size="small"
                        variant="outlined"
                        label={policy.rule_type}
                      />
                      <Chip
                        size="small"
                        color={actionColor}
                        label={action.toUpperCase()}
                      />
                      {policy.via_group && (
                        <Chip
                          size="small"
                          variant="outlined"
                          label={`Via group: ${policy.target_name || policy.target_id}`}
                        />
                      )}
                    </Stack>
                  </Stack>

                  <Typography variant="body2">
                    Identifier:{" "}
                    <Typography
                      component="code"
                      variant="body2"
                    >
                      {policy.identifier}
                    </Typography>
                  </Typography>

                  {policy.target_type === "user" && !policy.via_group && (
                    <Typography
                      variant="body2"
                      color="text.secondary"
                    >
                      Applied directly to this user.
                    </Typography>
                  )}
                </Stack>
              </CardContent>
            </CardActionArea>
          </Card>
        );
      })}
    </Stack>
  );
}

export default function UserDetails() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const { showToast } = useToast();

  const { data, isLoading, error } = useUserDetails(userId ?? "");
  const user = data?.user;
  const handleSelectDevice = useCallback(
    (deviceId: string) => {
      void navigate(`/devices/${deviceId}`);
    },
    [navigate],
  );
  const handleSelectPolicy = useCallback(
    (applicationId: string) => {
      void navigate(`/applications/${applicationId}`);
    },
    [navigate],
  );

  useEffect(() => {
    if (!error) return;

    console.error("Failed to load user details", error);
    showToast({
      message: error instanceof Error ? error.message : "Failed to load user details.",
      severity: "error",
    });
  }, [error, showToast]);

  const subtitleParts = user ? [user.displayName, user.upn].filter(Boolean) : [];
  const pageTitle = "User Details";
  const pageSubtitle = subtitleParts.length ? subtitleParts.join(" • ") : undefined;
  const breadcrumbs = [{ label: "Users", to: "/users" }, { label: user?.displayName ?? "Details" }];

  let content: ReactNode = null;

  if (!userId) {
    content = <Alert severity="error">Missing user identifier.</Alert>;
  } else if (isLoading) {
    content = <LinearProgress />;
  } else if (error) {
    content = <Alert severity="error">{error instanceof Error ? error.message : "Failed to load user details."}</Alert>;
  } else if (!data) {
    content = (
      <EmptyState
        title="User not found"
        description={`No user found with ID ${userId}`}
      />
    );
  } else {
    const { groups = [], devices = [], recent_blocks: events = [], policies = [] } = data;

    content = (
      <Grid
        container
        spacing={3}
      >
        <Grid size={{ xs: 12 }}>
          <SectionCard
            title="Profile"
            subheader="Directory details and group membership."
          >
            <Stack spacing={2}>
              <UserSummary user={data.user} />
              <Stack spacing={1}>
                <Typography
                  variant="subtitle2"
                  color="text.secondary"
                >
                  Group assignments
                </Typography>
                <GroupAssignmentChips groups={groups} />
              </Stack>
            </Stack>
          </SectionCard>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <SectionCard
            title="Devices"
            subheader="Santa agents associated with this user."
          >
            <DevicesTable
              devices={devices}
              onSelectDevice={handleSelectDevice}
            />
          </SectionCard>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <SectionCard
            title="Recent blocks"
            subheader="Latest Santa telemetry targeting this user."
          >
            <RecentBlocksList
              events={events}
              emptyMessage="No recent Santa blocks recorded for this user."
            />
          </SectionCard>
        </Grid>

        <Grid size={{ xs: 12 }}>
          <SectionCard
            title="Policies applied"
            subheader="Rules that currently impact this user."
          >
            <PoliciesList
              policies={policies}
              onSelectPolicy={handleSelectPolicy}
            />
          </SectionCard>
        </Grid>
      </Grid>
    );
  }

  return (
    <Stack spacing={3}>
      <PageHeader
        title={pageTitle}
        subtitle={pageSubtitle}
        breadcrumbs={breadcrumbs}
      />
      {content}
    </Stack>
  );
}
