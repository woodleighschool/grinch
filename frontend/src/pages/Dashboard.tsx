import { useMemo, useEffect, useState } from "react";
import { format, isValid, parseISO } from "date-fns";
import { useTheme } from "@mui/material/styles";
import { useBlockedEvents, useEventStats } from "../hooks/useQueries";
import { formatDateTime } from "../utils/dates";
import type { EventRecord } from "../api";
import {
  Alert,
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
import { BarChart } from "@mui/x-charts/BarChart";

type ChartDay = {
  day: string;
  label: string;
  counts: Record<string, number>;
  total: number;
};

type EventVolumeChartProps = {
  dataset: ChartDay[];
  kinds: string[];
  totalEvents: number;
  colorForKind: (kind: string) => string;
};

function renderRow(event: EventRecord) {
  const occurredAt = event.occurredAt ? formatDateTime(event.occurredAt) : "—";
  const processPath = typeof event.payload?.process_path === "string" ? (event.payload.process_path as string) : event.kind;
  const reason = typeof event.payload?.decision === "string" ? (event.payload.decision as string) : event.kind;

  return (
    <TableRow key={event.id} hover>
      <TableCell>{occurredAt}</TableCell>
      <TableCell>
        <Typography noWrap title={processPath} variant="body2">
          {processPath}
        </Typography>
      </TableCell>
      <TableCell>{event.machineId}</TableCell>
      <TableCell>
        <Chip label={reason} color="error" size="small" />
      </TableCell>
    </TableRow>
  );
}

function EventVolumeChart({ dataset, kinds, totalEvents, colorForKind }: EventVolumeChartProps) {
  if (dataset.length === 0 || kinds.length === 0) {
    return <Typography color="text.secondary">Not enough event history yet.</Typography>;
  }

  const chartDataset = dataset.map((day) => {
    const base: Record<string, number | string> = { label: day.label };
    kinds.forEach((kind) => {
      base[kind] = day.counts[kind] ?? 0;
    });
    return base;
  });

  const series = kinds.map((kind) => ({
    dataKey: kind,
    label: kind.replace(/_/g, " "),
    stack: "total",
    color: colorForKind(kind),
  }));

  return (
    <Stack spacing={2}>
      <BarChart
        dataset={chartDataset}
        height={320}
        xAxis={[{ scaleType: "band", dataKey: "label" }]}
        series={series}
        margin={{ top: 16, bottom: 80, left: 48, right: 24 }}
        slotProps={{
          legend: { direction: "horizontal", position: { vertical: "bottom", horizontal: "start" } },
        }}
      />
      <Typography variant="body2" color="text.secondary">
        {totalEvents.toLocaleString()} total events (last 14 days)
      </Typography>
    </Stack>
  );
}

export default function Dashboard() {
  const { events, loading, error } = useBlockedEvents();
  const { stats, loading: statsLoading, error: statsError } = useEventStats(14);
  const theme = useTheme();

  const [toast, setToast] = useState<{ open: boolean; message: string }>({ open: false, message: "" });

  useEffect(() => {
    if (error) {
      console.error("Recent events load failed", error);
      setToast({ open: true, message: "Failed to load recent events." });
    }
  }, [error]);

  useEffect(() => {
    if (statsError) {
      console.error("Event stats load failed", statsError);
      setToast({ open: true, message: "Failed to load policy outcome trends." });
    }
  }, [statsError]);

  const chartData = useMemo(() => {
    if (!stats || stats.length === 0) {
      return { dataset: [] as ChartDay[], kinds: [] as string[], totalEvents: 0 };
    }

    const dayMap = new Map<string, ChartDay>();
    const kindSet = new Set<string>();

    stats.forEach(({ bucket, kind, total }) => {
      const parsed = parseISO(bucket);
      const dayKey = isValid(parsed) ? format(parsed, "yyyy-MM-dd") : bucket;
      const label = isValid(parsed) ? format(parsed, "MMM d") : bucket;
      const existing = dayMap.get(dayKey) ?? { day: dayKey, label, counts: {}, total: 0 };
      existing.counts[kind] = (existing.counts[kind] ?? 0) + total;
      existing.total += total;
      dayMap.set(dayKey, existing);
      kindSet.add(kind);
    });

    const dataset = Array.from(dayMap.values()).sort((a, b) => a.day.localeCompare(b.day));
    const priority: Record<string, number> = {
      BLOCKLIST: 0,
      BLOCK_BINARY: 0,
      BLOCK_CERTIFICATE: 1,
      SILENT_BLOCKLIST: 2,
      ALLOWLIST: 3,
      ALLOW_BINARY: 3,
      ALLOW_CERTIFICATE: 3,
    };
    const kinds = Array.from(kindSet).sort((a, b) => (priority[a] ?? 99) - (priority[b] ?? 99) || a.localeCompare(b));
    const totalEvents = dataset.reduce((sum, day) => sum + day.total, 0);

    return { dataset, kinds, totalEvents };
  }, [stats]);

  const colorForKind = (kind: string) => {
    if (kind.includes("BLOCK")) {
      return kind.includes("SILENT") ? theme.palette.warning.main : theme.palette.error.main;
    }
    if (kind.includes("ALLOW")) return theme.palette.success.main;
    return theme.palette.info.main;
  };

  return (
    <Stack spacing={3}>
      <Card elevation={1}>
        <CardHeader title="Policy Outcomes" subheader="Stacked daily totals from the last 14 days" />
        {statsLoading && <LinearProgress />}
        <CardContent>
          {!statsLoading && (
            <EventVolumeChart dataset={chartData.dataset} kinds={chartData.kinds} totalEvents={chartData.totalEvents} colorForKind={colorForKind} />
          )}
        </CardContent>
      </Card>

      <Card elevation={1}>
        <CardHeader title="Recent Santa Events" subheader="Incoming telemetry from Santa agents appears as it is ingested." />
        {loading && <LinearProgress />}
        <CardContent>
          {!loading && events.length === 0 && <Typography color="text.secondary">No events recorded yet.</Typography>}
          {events.length > 0 && (
            <TableContainer component={Paper}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Occurred</TableCell>
                    <TableCell>Details</TableCell>
                    <TableCell>Machine</TableCell>
                    <TableCell>Kind</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>{events.map(renderRow)}</TableBody>
              </Table>
            </TableContainer>
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
