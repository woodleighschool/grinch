import { useEffect, useState, type ReactNode } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Card,
  CardContent,
  Chip,
  Grid,
  LinearProgress,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";

import { formatDateTime } from "../utils/dates";
import type { Device } from "../api";
import { useDeviceDetails } from "../hooks/useQueries";
import { EmptyState, PageHeader, RecentBlocksList, SectionCard, UserSummary } from "../components";
import { useToast } from "../hooks/useToast";

interface DeviceSummaryProps {
  device: Device;
}

function DeviceSummary({ device }: DeviceSummaryProps) {
  const rawSantaVersion = device.lastPreflightPayload?.santa_version;
  const santaVersion = typeof rawSantaVersion === "string" || typeof rawSantaVersion === "number" ? String(rawSantaVersion) : undefined;

  return (
    <Stack spacing={1.5}>
      <Stack spacing={0.25}>
        <Typography
          variant="h5"
          fontWeight={600}
        >
          {device.hostname || "Unnamed device"}
        </Typography>
        <Typography
          variant="body2"
          color="text.secondary"
        >
          Serial:{" "}
          <Typography
            component="span"
            variant="body2"
            fontFamily="monospace"
          >
            {device.serial || "—"}
          </Typography>
        </Typography>
        {device.id && (
          <Typography
            variant="body2"
            color="text.secondary"
          >
            Machine ID:{" "}
            <Typography
              component="span"
              variant="body2"
              fontFamily="monospace"
            >
              {device.id}
            </Typography>
          </Typography>
        )}
      </Stack>

      <Stack
        direction="row"
        spacing={1}
        flexWrap="wrap"
      >
        {device.clientMode && (
          <Chip
            size="small"
            variant="outlined"
            label={device.clientMode}
          />
        )}
        {device.lastSeen && (
          <Chip
            size="small"
            variant="outlined"
            label={`Last seen ${formatDateTime(device.lastSeen)}`}
          />
        )}
        {device.lastPreflightAt && (
          <Chip
            size="small"
            variant="outlined"
            label={`Last preflight ${formatDateTime(device.lastPreflightAt)}`}
          />
        )}
        {santaVersion && (
          <Chip
            size="small"
            variant="outlined"
            label={`Santa ${santaVersion}`}
          />
        )}
        {device.cleanSyncRequested && (
          <Chip
            size="small"
            variant="outlined"
            label="Clean sync pending"
          />
        )}
      </Stack>
    </Stack>
  );
}

export default function DeviceDetails() {
  const [isPreflightExpanded, setIsPreflightExpanded] = useState(false);

  const { deviceId } = useParams<{ deviceId: string }>();
  const navigate = useNavigate();

  const { showToast } = useToast();
  const { data, isLoading, error } = useDeviceDetails(deviceId ?? "");
  const device = data?.device;

  useEffect(() => {
    if (!error) return;

    console.error("Failed to load device details", error);
    showToast({
      message: error instanceof Error ? error.message : "Failed to load device details.",
      severity: "error",
    });
  }, [error, showToast]);

  const pageTitle = device?.hostname ?? "Device Details";
  const pageSubtitle = device?.serial;
  const breadcrumbs = [{ label: "Devices", to: "/devices" }, { label: pageTitle }];

  let content: ReactNode = null;

  if (!deviceId) {
    content = <Alert severity="error">Missing device identifier.</Alert>;
  } else if (isLoading) {
    content = <LinearProgress />;
  } else if (!data) {
    content = (
      <EmptyState
        title="Device not found"
        description={`No device found with ID ${deviceId}`}
      />
    );
  } else {
    const { primary_user, recent_blocks: events = [] } = data;
    const deviceDetail = data.device;

    const handleSelectUser = () => {
      if (!primary_user) return;
      void navigate(`/users/${primary_user.id}`);
    };

    content = (
      <Stack spacing={3}>
        <Grid
          container
          spacing={3}
        >
          <Grid size={{ xs: 12, md: 5, lg: 4 }}>
            <SectionCard
              title="Device overview"
              subheader="Reported by Santa from the latest preflight."
            >
              <DeviceSummary device={deviceDetail} />
            </SectionCard>
          </Grid>
          {primary_user && (
            <Grid size={{ xs: 12, md: 7, lg: 4 }}>
              <SectionCard
                title="Primary user"
                subheader="Determined from latest telemetry, not the logged-in user."
              >
                <UserSummary
                  component="div"
                  user={primary_user}
                  onClick={handleSelectUser}
                  sx={{
                    "cursor": "pointer",
                    "&:hover": { bgcolor: "action.hover" },
                    "borderRadius": 1,
                    "p": 1,
                  }}
                />
              </SectionCard>
            </Grid>
          )}

          <Grid size={{ xs: 12, md: 12, lg: 4 }}>
            <SectionCard
              title="Recent blocks"
              subheader="Latest Santa telemetry targeting this device."
            >
              <RecentBlocksList
                events={events}
                emptyMessage="No recent Santa blocks recorded for this device."
              />
            </SectionCard>
          </Grid>
        </Grid>

        {deviceDetail.lastPreflightPayload && (
          <Grid
            container
            spacing={3}
          >
            <Grid size={{ xs: 12 }}>
              <SectionCard
                title="Last preflight payload"
                subheader="Raw JSON payload reported by Santa."
              >
                <Accordion
                  expanded={isPreflightExpanded}
                  onChange={(_, expanded) => {
                    setIsPreflightExpanded(expanded);
                  }}
                  sx={{ boxShadow: "none", borderRadius: 1 }}
                >
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Typography fontWeight={600}>View JSON payload</Typography>
                  </AccordionSummary>
                  <AccordionDetails>
                    <Card variant="outlined">
                      <CardContent>
                        <TextField
                          fullWidth
                          multiline
                          minRows={8}
                          value={JSON.stringify(deviceDetail.lastPreflightPayload, null, 2)}
                          slotProps={{ input: { readOnly: true } }}
                        />
                      </CardContent>
                    </Card>
                  </AccordionDetails>
                </Accordion>
              </SectionCard>
            </Grid>
          </Grid>
        )}
      </Stack>
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
