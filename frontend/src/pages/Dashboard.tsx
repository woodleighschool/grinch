import { useEffect, useMemo } from "react";
import { format, isValid, parseISO } from "date-fns";
import { group } from "d3-array";
import { Card, CardContent, CardHeader, LinearProgress, Stack, Typography, useTheme } from "@mui/material";
import { BarChart } from "@mui/x-charts/BarChart";
import EventNoteIcon from "@mui/icons-material/EventNote";

import { EmptyState, PageHeader } from "../components";
import { useEventStats } from "../hooks/useQueries";
import { useToast } from "../hooks/useToast";

type ChartDay = {
  day: string;
  label: string;
  counts: Record<string, number>;
  total: number;
};

interface EventVolumeChartProps {
  dataset: ChartDay[];
  kinds: string[];
  totalEvents: number;
  colorForKind: (kind: string) => string;
}

function EventVolumeChart({ dataset, kinds, totalEvents, colorForKind }: EventVolumeChartProps) {
  if (dataset.length === 0 || kinds.length === 0) {
    return (
      <EmptyState
        title="No Data Available"
        description="Not enough event history yet to display trends."
        icon={<EventNoteIcon fontSize="inherit" />}
      />
    );
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
          legend: {
            direction: "horizontal",
            position: { vertical: "bottom", horizontal: "start" },
          },
        }}
      />

      <Typography
        variant="body2"
        color="text.secondary"
      >
        {totalEvents.toLocaleString()} total events (last 14 days)
      </Typography>
    </Stack>
  );
}

export default function Dashboard() {
  const theme = useTheme();
  const { stats, loading: statsLoading, error: statsError } = useEventStats(14);
  const { showToast } = useToast();

  useEffect(() => {
    if (statsError == null) return;

    console.error("Event stats load failed", statsError);
    const message = statsError || "Failed to load policy outcome trends.";
    showToast({
      message,
      severity: "error",
    });
  }, [statsError, showToast]);

  const chartData = useMemo(() => {
    if (stats.length === 0) {
      return {
        dataset: [] as ChartDay[],
        kinds: [] as string[],
        totalEvents: 0,
      };
    }

    const kindSet = new Set<string>();
    const grouped = group(stats, ({ bucket }) => {
      const parsed = parseISO(bucket);
      return isValid(parsed) ? format(parsed, "yyyy-MM-dd") : bucket;
    });

    const dataset: ChartDay[] = [];

    grouped.forEach((entries, dayKey) => {
      const entryArray = Array.from(entries);
      if (entryArray.length === 0) return;

      const firstEntry = entryArray[0];
      if (!firstEntry) return;
      const parsed = parseISO(firstEntry.bucket);
      const label = isValid(parsed) ? format(parsed, "MMM d") : firstEntry.bucket;
      const counts: Record<string, number> = {};
      let total = 0;

      entryArray.forEach(({ kind, total: entryTotal }) => {
        counts[kind] = (counts[kind] ?? 0) + entryTotal;
        total += entryTotal;
        kindSet.add(kind);
      });

      if (total === 0) return;

      dataset.push({
        day: dayKey,
        label,
        counts,
        total,
      });
    });

    dataset.sort((a, b) => a.day.localeCompare(b.day));

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

    if (kind.includes("ALLOW")) {
      return kind.includes("UNKNOWN") ? theme.palette.warning.main : theme.palette.success.main;
    }

    return theme.palette.info.main;
  };

  return (
    <Stack spacing={3}>
      <PageHeader
        title="Dashboard"
        subtitle="Overview of recent activity and policy outcomes."
      />

      <Card elevation={1}>
        <CardHeader
          title="Policy Outcomes"
          subheader="Stacked daily totals from the last 14 days."
        />
        {statsLoading && <LinearProgress />}

        <CardContent>
          {!statsLoading && (
            <EventVolumeChart
              dataset={chartData.dataset}
              kinds={chartData.kinds}
              totalEvents={chartData.totalEvents}
              colorForKind={colorForKind}
            />
          )}
        </CardContent>
      </Card>
    </Stack>
  );
}
