import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { formatDateTime, formatCompactDateTime } from "../utils/dates";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import type { DirectoryUser, UserDetailResponse } from "../api";
import { getUserDetails } from "../api";

function getDisplayName(user: DirectoryUser): string {
  return user.display_name || user.principal_name;
}

function capitalize(value: string): string {
  if (!value) {
    return "";
  }
  return value.charAt(0).toUpperCase() + value.slice(1);
}

interface SummaryProps {
  user: DirectoryUser;
}

function UserSummary({ user }: SummaryProps) {
  return (
    <div>
      <div style={{ marginBottom: "16px" }}>
        <div
          style={{
            fontSize: "24px",
            fontWeight: 600,
            marginBottom: "4px",
          }}
        >
          {getDisplayName(user)}
        </div>
        <div
          style={{
            fontSize: "16px",
            color: "var(--text-muted)",
            marginBottom: "8px",
          }}
        >
          {user.principal_name}
        </div>
      </div>

      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          gap: "16px",
          alignItems: "center",
          marginBottom: "16px",
        }}
      >
        <Badge variant="secondary" caps>{user.user_type}</Badge>
        {user.user_type === "local" && user.is_protected_local && <Badge variant="secondary" caps>Protected</Badge>}
        <Badge size="md" variant="neutral" label="Created" value={formatDateTime(user.created_at)} caps />
        {user.synced_at && <Badge size="md" variant="neutral" label="Last Synced" value={formatDateTime(user.synced_at)} caps />}
      </div>
    </div>
  );
}

interface UserDetailsPanelProps {
  details: UserDetailResponse;
}

function UserDetailsPanel({ details }: UserDetailsPanelProps) {
  const groups = details.groups ?? [];
  const devices = details.devices ?? [];
  const events = details.recent_events ?? [];
  const policies = details.policies ?? [];

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "32px" }}>
      <section>
        <h3
          style={{
            marginBottom: "16px",
            fontSize: "18px",
            fontWeight: 600,
          }}
        >
          Group Assignments
        </h3>
        {groups.length === 0 ? (
          <div className="empty-state">
            <p style={{ margin: 0 }}>This user is not assigned to any groups.</p>
          </div>
        ) : (
          <div
            style={{
              display: "flex",
              gap: "8px",
              flexWrap: "wrap",
            }}
          >
            {groups.map((group) => (
              <Badge key={group.id} variant="secondary">
                {group.display_name}
              </Badge>
            ))}
          </div>
        )}
      </section>

      <section>
        <h3
          style={{
            marginBottom: "16px",
            fontSize: "18px",
            fontWeight: 600,
          }}
        >
          Devices
        </h3>
        {devices.length === 0 ? (
          <div className="empty-state">
            <p style={{ margin: 0 }}>No devices have reported this user yet.</p>
          </div>
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table className="table">
              <thead>
                <tr>
                  <th>Hostname</th>
                  <th>Serial</th>
                  <th>Machine ID</th>
                  <th>Last Seen</th>
                </tr>
              </thead>
              <tbody>
                {devices.map((device) => (
                  <tr key={device.id}>
                    <td>{device.hostname}</td>
                    <td>{device.serial_number || "—"}</td>
                    <td>
                      <span className="principal-box">{device.machine_id}</span>
                    </td>
                    <td>{device.last_seen ? formatDateTime(device.last_seen) : "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section>
        <h3
          style={{
            marginBottom: "16px",
            fontSize: "18px",
            fontWeight: 600,
          }}
        >
          Recent Events
        </h3>
        {events.length === 0 ? (
          <div className="empty-state">
            <p style={{ margin: 0 }}>No recent Santa events recorded for this user.</p>
          </div>
        ) : (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "12px",
            }}
          >
            {events.map((event) => (
              <div
                key={event.id}
                style={{
                  border: "1px solid var(--border-primary)",
                  borderRadius: "8px",
                  padding: "16px",
                  background: "var(--bg-primary)",
                }}
              >
                <div
                  style={{
                    fontWeight: 600,
                    color: "var(--text-primary)",
                    marginBottom: "8px",
                  }}
                >
                  {event.process_path}
                </div>
                <div
                  style={{
                    fontSize: "13px",
                    color: "var(--text-muted)",
                    marginBottom: "4px",
                  }}
                >
                  {event.hostname ? `Host: ${event.hostname}` : "Host: —"} · {formatCompactDateTime(event.occurred_at)}
                </div>
                <div
                  style={{
                    fontSize: "13px",
                    color: "var(--text-primary)",
                  }}
                >
                  Decision: {event.decision || "Unknown"}
                  {event.blocked_reason ? ` · Reason: ${event.blocked_reason}` : ""}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      <section>
        <h3
          style={{
            marginBottom: "16px",
            fontSize: "18px",
            fontWeight: 600,
          }}
        >
          Policies Applied
        </h3>
        {policies.length === 0 ? (
          <div className="empty-state">
            <p style={{ margin: 0 }}>No policies currently target this user.</p>
          </div>
        ) : (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "12px",
            }}
          >
            {policies.map((policy) => (
              <div
                key={policy.scope_id}
                style={{
                  border: "1px solid var(--border-primary)",
                  borderRadius: "8px",
                  padding: "16px",
                  background: "var(--bg-primary)",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    gap: "12px",
                    flexWrap: "wrap",
                    alignItems: "center",
                    marginBottom: "8px",
                  }}
                >
                  <strong>{policy.application_name}</strong>
                  <div
                    style={{
                      display: "flex",
                      gap: "6px",
                      flexWrap: "wrap",
                    }}
                  >
                    <Badge variant="secondary">{policy.rule_type}</Badge>
                    <Badge variant="secondary">{policy.action.toUpperCase()}</Badge>
                    {policy.via_group && <Badge variant="secondary">Via group: {policy.target_name || policy.target_id}</Badge>}
                  </div>
                </div>
                <div
                  style={{
                    fontSize: "13px",
                    color: "var(--text-primary)",
                    marginBottom: "4px",
                  }}
                >
                  Identifier: <span className="principal-box">{policy.identifier}</span>
                </div>
                {!policy.via_group && (
                  <div
                    style={{
                      fontSize: "13px",
                      color: "var(--text-muted)",
                    }}
                  >
                    Applied directly to this user.
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

export default function UserDetails() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const [details, setDetails] = useState<UserDetailResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!userId) {
      setError("Missing user identifier.");
      return;
    }

    let cancelled = false;
    (async () => {
      try {
        const result = await getUserDetails(userId);
        if (!cancelled) {
          setDetails(result);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load user.");
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [userId]);

  if (error) {
    return (
      <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
        <div className="card">
          <h2>User Details</h2>
          <p className="error-text">{error}</p>
          <Link to="/users" className="primary" style={{ textDecoration: "none", marginTop: "12px" }}>
            Back to users
          </Link>
        </div>
      </div>
    );
  }

  if (!details) {
    return (
      <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
        <div className="card">
          <h2>User Details</h2>
          <p className="muted-text">User not found.</p>
          <Link to="/users" className="primary" style={{ textDecoration: "none", marginTop: "12px" }}>
            Back to users
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
      <div style={{ marginBottom: "24px" }}>
        <Button variant="secondary" onClick={() => navigate(-1)}>
          ← Back
        </Button>
        <h2 style={{ marginTop: "16px", marginBottom: "8px" }}>User Details</h2>
        <p className="muted-text" style={{ marginBottom: "24px" }}>
          View user information, group assignments, devices, and recent activity.
        </p>
      </div>

      <div className="card" style={{ marginBottom: "24px" }}>
        <UserSummary user={details.user} />
      </div>

      <div className="card">
        <UserDetailsPanel details={details} />
      </div>
    </div>
  );
}
