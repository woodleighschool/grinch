import { useCallback, useEffect, useState, type ReactNode } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Accordion, AccordionDetails, AccordionSummary, Alert, Box, Button, Chip, Grid, LinearProgress, Stack, Typography } from "@mui/material";
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
  const identityLabel = device.hostname || device.serial || "Unnamed device";

  return (
    <Stack spacing={2}>
      <Stack spacing={0.25}>
        <Typography
          variant="h6"
          fontWeight={600}
        >
          {identityLabel}
        </Typography>
        {device.serial && (
          <Typography
            variant="body2"
            color="text.secondary"
          >
            Serial {device.serial}
          </Typography>
        )}
      </Stack>

      {device.id && (
        <Stack spacing={0.25}>
          <Typography
            variant="caption"
            color="text.secondary"
          >
            Machine ID
          </Typography>
          <Typography
            variant="body2"
            fontFamily="monospace"
          >
            {device.id}
          </Typography>
        </Stack>
      )}

      <Stack
        direction="row"
        spacing={1}
        flexWrap="wrap"
        useFlexGap
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
  const deviceDetail = data?.device ?? null;
  const primaryUser = data?.primary_user ?? null;
  const recentBlocks = data?.recent_blocks ?? [];
  const handleSelectUser = useCallback(() => {
    if (!primaryUser) return;
    void navigate(`/users/${primaryUser.id}`);
  }, [navigate, primaryUser]);

  useEffect(() => {
    if (!error) return;

    console.error("Failed to load device details", error);
    showToast({
      message: error instanceof Error ? error.message : "Failed to load device details.",
      severity: "error",
    });
  }, [error, showToast]);

  const identityParts = deviceDetail ? [deviceDetail.hostname, deviceDetail.serial].filter(Boolean) : [];
  const pageTitle = "Device Details";
  const pageSubtitle = identityParts.length ? identityParts.join(" • ") : undefined;
  const breadcrumbs = [{ label: "Devices", to: "/devices" }, { label: deviceDetail?.hostname ?? deviceDetail?.serial ?? "Details" }];

  let content: ReactNode = null;

  if (!deviceId) {
    content = <Alert severity="error">Missing device identifier.</Alert>;
  } else if (isLoading) {
    content = <LinearProgress />;
  } else if (error) {
    content = <Alert severity="error">{error instanceof Error ? error.message : "Failed to load device details."}</Alert>;
  } else if (!deviceDetail) {
    content = (
      <EmptyState
        title="Device not found"
        description={`No device found with ID ${deviceId}`}
      />
    );
  } else {
    content = (
      <Stack spacing={3}>
        <Grid
          container
          spacing={3}
        >
          <Grid size={{ xs: 12, md: 7 }}>
            <SectionCard
              title="Device overview"
              subheader="Reported by Santa from the latest preflight."
            >
              <DeviceSummary device={deviceDetail} />
            </SectionCard>
          </Grid>
          {primaryUser && (
            <Grid size={{ xs: 12, md: 5 }}>
              <SectionCard
                title="Primary user"
                subheader="Determined from latest telemetry, not the logged-in user."
              >
                <Stack spacing={1.5}>
                  <UserSummary
                    component="div"
                    user={primaryUser}
                    showMetadata={false}
                  />
                  <Button
                    variant="outlined"
                    onClick={handleSelectUser}
                    sx={{ alignSelf: "flex-start" }}
                  >
                    View full profile
                  </Button>
                </Stack>
              </SectionCard>
            </Grid>
          )}

          <Grid size={{ xs: 12 }}>
            <SectionCard
              title="Recent blocks"
              subheader="Latest Santa telemetry targeting this device."
            >
              <RecentBlocksList
                events={recentBlocks}
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
                    <Box
                      component="pre"
                      sx={{
                        p: 2,
                        borderRadius: 2,
                        border: (theme) => `1px solid ${theme.palette.divider}`,
                        bgcolor: "background.paper",
                        overflowX: "auto",
                        fontFamily: "Consolas, Menlo, monospace",
                        fontSize: 13,
                      }}
                    >
                      {JSON.stringify(deviceDetail.lastPreflightPayload, null, 2)}
                    </Box>
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
