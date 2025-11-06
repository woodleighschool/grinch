import { useEffect, useMemo, useState } from 'react';
import { Device, listDevices } from '../api';

export default function Devices() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    void loadDevices();
  }, []);

  async function loadDevices() {
    setLoading(true);
    setError(null);
    try {
      const result = await listDevices();
      setDevices(Array.isArray(result) ? result : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load devices');
    } finally {
      setLoading(false);
    }
  }

  const filteredDevices = useMemo(() => {
    return devices.filter((device) => {
      const term = searchTerm.trim().toLowerCase();
      if (!term) {
        return true;
      }
      const target = [
        device.hostname,
        device.serial_number,
        device.machine_id,
        device.primary_user_principal,
        device.primary_user_display_name
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();
      return target.includes(term);
    });
  }, [devices, searchTerm]);

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

  function getStatusCategory(lastSeen?: string | null): 'success' | 'warning' | 'danger' | 'muted' {
    if (!lastSeen) {
      return 'muted';
    }
    const lastSeenDate = new Date(lastSeen);
    const diffMs = Date.now() - lastSeenDate.getTime();
    const diffMinutes = diffMs / (1000 * 60);
    if (diffMinutes <= 15) return 'success'; // online
    if (diffMinutes <= 120) return 'warning'; // recently active
    return 'danger'; // offline
  }

  const statusCounts = useMemo(() => {
    let online = 0;
    let warning = 0;
    let offline = 0;
    devices.forEach((device) => {
      const category = getStatusCategory(device.last_seen);
      if (category === 'success') online += 1;
      else if (category === 'warning') warning += 1;
      else if (category === 'danger') offline += 1;
    });
    return { online, warning, offline, total: devices.length };
  }, [devices]);

  if (loading) {
    return (
      <div className="card">
        <h2>Devices</h2>
        <p>Loading devices…</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="card">
        <h2>Devices</h2>
        <div className="alert error">
          <p style={{ margin: 0 }}>{error}</p>
          <button className="secondary" style={{ marginTop: '12px' }} onClick={() => void loadDevices()}>Retry</button>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <h2>Devices</h2>
      <p>Monitor and manage Santa agents deployed across your organization's devices.</p>

      <div className="stat-bubble-row">
        <div className="stat-bubble success">
          <span className="stat-bubble-value">{statusCounts.online}</span>
          <span className="stat-bubble-label">Online</span>
        </div>
        <div className="stat-bubble danger">
          <span className="stat-bubble-value">{statusCounts.offline}</span>
          <span className="stat-bubble-label">Offline</span>
        </div>
        <div className="stat-bubble info">
          <span className="stat-bubble-value">{statusCounts.total}</span>
          <span className="stat-bubble-label">Total Devices</span>
        </div>
      </div>

      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '16px'
      }}>
        <input
          type="text"
          placeholder="Search devices..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          style={{
            flex: 1,
            maxWidth: '300px'
          }}
        />
        <button
          className="primary"
          style={{ marginLeft: '16px' }}
          onClick={() => void loadDevices()}
        >
          Refresh
        </button>
      </div>

      {filteredDevices.length === 0 ? (
        <div className="empty-state">
          {searchTerm ? (
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
        <div style={{ overflowX: 'auto' }}>
          <table className="table">
            <thead>
              <tr>
                <th>Hostname</th>
                <th>Serial</th>
                <th>Status</th>
                <th>Primary User</th>
                <th>OS Version</th>
                <th>Santa Version</th>
                <th>Last Seen</th>
              </tr>
            </thead>
            <tbody>
              {filteredDevices.map((device) => (
                <tr key={device.id}>
                  <td>
                    <span className="principal-box">{device.hostname}</span>
                  </td>
                  <td>
                    <code>{device.serial_number ?? device.machine_id}</code>
                  </td>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      {(() => {
                        const category = getStatusCategory(device.last_seen);
                        return (
                          <>
                            <span className={`status-dot ${category}`} />
                            <span className={`status-text ${category}`} style={{ fontWeight: 500 }}>
                              {category === 'success' ? 'Online' : category === 'warning' ? 'Inactive' : device.last_seen ? 'Offline' : 'Unknown'}
                            </span>
                          </>
                        );
                      })()}
                    </div>
                  </td>
                  <td>
                    {device.primary_user_display_name ?? device.primary_user_principal ?? '—'}
                  </td>
                  <td>
                    {device.os_version ? (
                      <span>{device.os_version}{device.os_build ? ` (${device.os_build})` : ''}</span>
                    ) : '—'}
                  </td>
                  <td>
                    {device.santa_version ? (
                      <span className="badge secondary">{device.santa_version}</span>
                    ) : '—'}
                  </td>
                  <td>
                    {device.last_seen ? formatLastSeen(device.last_seen) : 'Never'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
