import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
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

function formatDateTime(value?: string): string {
    if (!value) {
        return "—";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }
    return date.toLocaleString();
}

function formatEventTime(value: string): string {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }
    return date.toLocaleString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    });
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
                {user.email && (
                    <div
                        style={{
                            fontSize: "14px",
                            color: "var(--text-primary)",
                            marginBottom: "8px",
                        }}
                    >
                        <strong>Email:</strong> {user.email}
                    </div>
                )}
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
                <span className="badge secondary">
                    {capitalize(user.user_type)}
                </span>
                {user.user_type === "local" && user.is_protected_local && (
                    <span className="badge secondary">Protected</span>
                )}
                {user.role_groups && user.role_groups.length > 0 && (
                    <span className="badge secondary">
                        Roles: {user.role_groups.join(", ")}
                    </span>
                )}
                <span className="summary-pill neutral" title="Account created">
                    <span className="summary-pill-label">Created</span>
                    <span className="summary-pill-value">
                        {formatDateTime(user.created_at)}
                    </span>
                </span>
                {user.synced_at && (
                    <span className="summary-pill neutral" title="Last sync">
                        <span className="summary-pill-label">Last Synced</span>
                        <span className="summary-pill-value">
                            {formatDateTime(user.synced_at)}
                        </span>
                    </span>
                )}
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
                        <p style={{ margin: 0 }}>
                            This user is not assigned to any groups.
                        </p>
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
                            <span key={group.id} className="badge secondary">
                                {group.display_name}
                            </span>
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
                        <p style={{ margin: 0 }}>
                            No devices have reported this user yet.
                        </p>
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
                                            <span className="principal-box">
                                                {device.machine_id}
                                            </span>
                                        </td>
                                        <td>
                                            {device.last_seen
                                                ? formatDateTime(
                                                    device.last_seen,
                                                )
                                                : "—"}
                                        </td>
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
                        <p style={{ margin: 0 }}>
                            No recent Santa events recorded for this user.
                        </p>
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
                                    {event.hostname
                                        ? `Host: ${event.hostname}`
                                        : "Host: —"}{" "}
                                    · {formatEventTime(event.occurred_at)}
                                </div>
                                <div
                                    style={{
                                        fontSize: "13px",
                                        color: "var(--text-primary)",
                                    }}
                                >
                                    Decision: {event.decision || "Unknown"}
                                    {event.blocked_reason
                                        ? ` · Reason: ${event.blocked_reason}`
                                        : ""}
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
                        <p style={{ margin: 0 }}>
                            No policies currently target this user.
                        </p>
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
                                        <span className="badge secondary">
                                            {policy.rule_type}
                                        </span>
                                        <span className="badge secondary">
                                            {policy.action.toUpperCase()}
                                        </span>
                                        {policy.via_group && (
                                            <span className="badge secondary">
                                                Via group:{" "}
                                                {policy.target_name ||
                                                    policy.target_id}
                                            </span>
                                        )}
                                    </div>
                                </div>
                                <div
                                    style={{
                                        fontSize: "13px",
                                        color: "var(--text-primary)",
                                        marginBottom: "4px",
                                    }}
                                >
                                    Identifier:{" "}
                                    <span className="principal-box">
                                        {policy.identifier}
                                    </span>
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
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!userId) {
            setError("Missing user identifier.");
            setLoading(false);
            return;
        }

        let cancelled = false;
        setLoading(true);
        (async () => {
            try {
                const result = await getUserDetails(userId);
                if (!cancelled) {
                    setDetails(result);
                }
            } catch (err) {
                if (!cancelled) {
                    setError(
                        err instanceof Error
                            ? err.message
                            : "Failed to load user.",
                    );
                }
            } finally {
                if (!cancelled) {
                    setLoading(false);
                }
            }
        })();

        return () => {
            cancelled = true;
        };
    }, [userId]);

    if (loading) {
        return (
            <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
                <div className="card">
                    <h2>User Details</h2>
                    <p>Loading user information…</p>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
                <div className="card">
                    <h2>User Details</h2>
                    <p className="error-text">{error}</p>
                    <Link
                        to="/users"
                        className="primary"
                        style={{ textDecoration: "none", marginTop: "12px" }}
                    >
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
                    <Link
                        to="/users"
                        className="primary"
                        style={{ textDecoration: "none", marginTop: "12px" }}
                    >
                        Back to users
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
            <div style={{ marginBottom: "24px" }}>
                <button
                    type="button"
                    className="secondary"
                    onClick={() => navigate(-1)}
                >
                    ← Back
                </button>
                <h2 style={{ marginTop: "16px", marginBottom: "8px" }}>
                    User Details
                </h2>
                <p className="muted-text" style={{ marginBottom: "24px" }}>
                    View user information, group assignments, devices, and
                    recent activity.
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
