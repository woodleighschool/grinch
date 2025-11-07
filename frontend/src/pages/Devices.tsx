import { useEffect, useState, useMemo } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Badge, Table, Button } from "../components";
import { useSearch, searchConfigs } from "../hooks/useSearch";
import { Device, listDevices } from "../api";

export default function Devices() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [error, setError] = useState<string | null>(null);

  const { searchTerm, setSearchTerm, filteredItems: filteredDevices, clearSearch, isSearching } = useSearch(devices, searchConfigs.devices);

  useEffect(() => {
    void loadDevices();
  }, []);

  async function loadDevices() {
    setError(null);
    try {
      const result = await listDevices();
      setDevices(Array.isArray(result) ? result : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load devices");
    }
  }

  function formatLastSeen(isoString: string) {
    const date = new Date(isoString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffMins < 60) {
      return `${diffMins}m ago`;
    } else if (diffHours < 24) {
      return `${diffHours}h ago`;
    } else {
      return `${diffDays}d ago`;
    }
  }

  function getStatusCategory(lastSeen?: string | null): "success" | "warning" | "danger" | "muted" {
    if (!lastSeen) {
      return "muted";
    }
    const lastSeenDate = new Date(lastSeen);
    const diffMs = Date.now() - lastSeenDate.getTime();
    const diffMinutes = diffMs / (1000 * 60);
    if (diffMinutes <= 15) return "success"; // online
    if (diffMinutes <= 120) return "warning"; // recently active
    return "danger"; // offline
  }

  const statusCounts = useMemo(() => {
    let online = 0;
    let warning = 0;
    let offline = 0;
    devices.forEach((device) => {
      const category = getStatusCategory(device.last_seen);
      if (category === "success") online += 1;
      else if (category === "warning") warning += 1;
      else if (category === "danger") offline += 1;
    });
    return { online, warning, offline, total: devices.length };
  }, [devices]);

  const columns = useMemo<ColumnDef<Device>[]>(
    () => [
      {
        accessorKey: "hostname",
        header: "Hostname",
        cell: ({ getValue }) => <span className="principal-box">{getValue() as string}</span>,
      },
      {
        accessorKey: "serial_number",
        header: "Serial",
        cell: ({ row }) => <code>{row.original.serial_number ?? row.original.machine_id}</code>,
      },
      {
        accessorKey: "last_seen",
        header: "Status",
        cell: ({ getValue, row }) => {
          const lastSeen = getValue() as string | null;
          const category = getStatusCategory(lastSeen);
          return (
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "6px",
              }}
            >
              <span className={`status-dot ${category}`} />
              <span
                className={`status-text ${category}`}
                style={{
                  fontWeight: 500,
                }}
              >
                {category === "success" ? "Online" : category === "warning" ? "Inactive" : lastSeen ? "Offline" : "Unknown"}
              </span>
            </div>
          );
        },
      },
      {
        accessorKey: "primary_user_display_name",
        header: "Primary User",
        cell: ({ row }) => row.original.primary_user_display_name ?? row.original.primary_user_principal ?? "—",
      },
      {
        accessorKey: "os_version",
        header: "OS Version",
        cell: ({ row }) => {
          const { os_version, os_build } = row.original;
          return os_version ? (
            <span>
              {os_version}
              {os_build ? ` (${os_build})` : ""}
            </span>
          ) : (
            "—"
          );
        },
      },
      {
        accessorKey: "santa_version",
        header: "Santa Version",
        cell: ({ getValue }) => {
          const version = getValue() as string | null;
          return version ? <Badge variant="secondary">{version}</Badge> : "—";
        },
      },
      {
        accessorKey: "last_seen",
        header: "Last Seen",
        cell: ({ getValue }) => {
          const lastSeen = getValue() as string | null;
          return lastSeen ? formatLastSeen(lastSeen) : "Never";
        },
      },
    ],
    [],
  );

  if (error) {
    return (
      <div className="card">
        <h2>Devices</h2>
        <div className="alert error">
          <p style={{ margin: 0 }}>{error}</p>
          <Button variant="secondary" style={{ marginTop: "12px" }} onClick={() => void loadDevices()}>
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <h2>Devices</h2>
      <p>Monitor and manage Santa agents deployed across your organization's devices.</p>

      <div className="badge-row">
        <Badge size="lg" variant="success" value={statusCounts.online} label="Online" caps />
        <Badge size="lg" variant="danger" value={statusCounts.offline} label="Offline" caps />
        <Badge size="lg" variant="info" value={statusCounts.total} label="Total Devices" caps />
      </div>

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "16px",
        }}
      >
        <input
          type="text"
          placeholder="Search devices..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          style={{
            flex: 1,
            maxWidth: "300px",
          }}
        />
        <Button variant="primary" style={{ marginLeft: "16px" }} onClick={() => void loadDevices()}>
          Refresh
        </Button>
      </div>

      {filteredDevices.length === 0 ? (
        <div className="empty-state">
          {isSearching ? (
            <>
              <h3>No devices found</h3>
              <p>No devices match the search term "{searchTerm}"</p>
            </>
          ) : (
            <>
              <h3>No devices connected</h3>
              <p>Deploy Santa agents to see devices here.</p>
            </>
          )}
        </div>
      ) : (
        <Table
          data={filteredDevices}
          columns={columns}
          globalFilter={searchTerm}
          sorting={true}
          filtering={true}
          pagination={devices.length > 10}
          pageSize={100}
        />
      )}
    </div>
  );
}
