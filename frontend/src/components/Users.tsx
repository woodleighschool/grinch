import { useEffect, useMemo, useRef, useState } from 'react';
import type { DirectoryUser, UserDetailResponse } from '../api';
import { getUserDetails, listUsers } from '../api';

function getDisplayName(user: DirectoryUser): string {
  return user.display_name || user.principal_name;
}

function capitalize(value: string): string {
  if (!value) {
    return '';
  }
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function formatDateTime(value?: string): string {
  if (!value) {
    return '—';
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
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function Users() {
  const [users, setUsers] = useState<DirectoryUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [selectedDetails, setSelectedDetails] = useState<UserDetailResponse | null>(null);
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [detailsError, setDetailsError] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const usersData = await listUsers();
        setUsers(Array.isArray(usersData) ? usersData : []);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError('Failed to load users');
        }
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  useEffect(() => {
    if (!selectedUserId) {
      setSelectedDetails(null);
      setDetailsError(null);
      setDetailsLoading(false);
      return;
    }

    let cancelled = false;
    setDetailsError(null);
    setDetailsLoading(true);

    (async () => {
      try {
        const details = await getUserDetails(selectedUserId);
        if (!cancelled) {
          setSelectedDetails(details);
        }
      } catch (err) {
        if (!cancelled) {
          if (err instanceof Error) {
            setDetailsError(err.message);
          } else {
            setDetailsError('Failed to load user details');
          }
        }
      } finally {
        if (!cancelled) {
          setDetailsLoading(false);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [selectedUserId]);

  const filteredUsers = useMemo(
    () =>
      users.filter((user) => {
        const display = getDisplayName(user).toLowerCase();
        const principal = user.principal_name.toLowerCase();
        const term = searchTerm.toLowerCase();
        return display.includes(term) || principal.includes(term);
      }),
    [users, searchTerm],
  );

  if (loading) {
    return (
      <div className="card">
        <h2>Users</h2>
        <p>Loading users...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="card">
        <h2>Users</h2>
        <p className="error-text">Failed to load users: {error}</p>
      </div>
    );
  }

  return (
    <div className="card">
      <h2>Users</h2>
      <p>Manage individual user access and permissions from your Entra ID directory.</p>

      {users.length > 0 && (
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: '12px',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: '16px',
          }}
        >
          <div
            style={{
              display: 'flex',
              gap: '8px',
              flex: '1 1 260px',
              minWidth: '220px',
            }}
          >
            <input
              type="search"
              placeholder="Search users..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              style={{ flex: 1 }}
              aria-label="Search users"
            />
            {searchTerm && (
              <button
                type="button"
                className="secondary"
                onClick={() => setSearchTerm('')}
                title="Clear search"
                style={{ whiteSpace: 'nowrap' }}
              >
                Clear
              </button>
            )}
          </div>
          <div
            style={{
              color: 'var(--text-muted)',
              fontSize: '14px',
              marginLeft: 'auto',
              textAlign: 'right',
            }}
          >
            Showing {filteredUsers.length} of {users.length} user{users.length !== 1 ? 's' : ''}
          </div>
        </div>
      )}

      {filteredUsers.length === 0 ? (
        <div className="empty-state">
          {searchTerm ? (
            <>
              <h3>No users found</h3>
              <p>No users match the search term "{searchTerm}"</p>
            </>
          ) : (
            <>
              <h3>No users found</h3>
              <p>No users are available in the directory.</p>
            </>
          )}
        </div>
      ) : (
        filteredUsers.map((user) => (
          <UserAssignmentCard
            key={user.id}
            user={user}
            isExpanded={selectedUserId === user.id}
            onToggle={() => {
              setSelectedUserId((current) => (current === user.id ? null : user.id));
            }}
            details={selectedUserId === user.id ? selectedDetails : null}
            loading={selectedUserId === user.id && detailsLoading}
            error={selectedUserId === user.id ? detailsError : null}
          />
        ))
      )}
    </div>
  );
}

interface UserAssignmentCardProps {
  user: DirectoryUser;
  isExpanded: boolean;
  onToggle: () => void;
  details: UserDetailResponse | null;
  loading: boolean;
  error: string | null;
}

function UserAssignmentCard({
  user,
  isExpanded,
  onToggle,
  details,
  loading,
  error,
}: UserAssignmentCardProps) {
  const cardRef = useRef<HTMLElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [contentHeight, setContentHeight] = useState(0);
  const detailsId = `user-card-details-${user.id}`;
  const displayUser = details?.user ?? user;

  useEffect(() => {
    if (isExpanded && contentRef.current) {
      setContentHeight(contentRef.current.scrollHeight);
    }
  }, [isExpanded, details, loading, error]);

  const handleToggle = () => {
    if (!isExpanded && cardRef.current) {
      const rect = cardRef.current.getBoundingClientRect();
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
      const cardTop = rect.top + scrollTop;

      onToggle();

      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          window.scrollTo({
            top: cardTop - 20,
            behavior: 'smooth',
          });
        });
      });
    } else {
      onToggle();
    }
  };

  const groupCount = details?.groups?.length;
  const deviceCount = details?.devices?.length;

  return (
    <article className={`assignment-card${isExpanded ? ' expanded' : ''}`} ref={cardRef}>
      <header
        className="assignment-card-header"
        onClick={handleToggle}
        role="button"
        tabIndex={0}
        aria-expanded={isExpanded}
        aria-controls={detailsId}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            handleToggle();
          }
        }}
      >
        <div className="assignment-card-summary">
          <span className="assignment-card-chevron" aria-hidden="true">
            ›
          </span>
          <span className="assignment-card-icon" aria-hidden="true">👤</span>
          <div className="assignment-card-summary-main">
            <div className="assignment-card-summary-title">
              <h3 className="assignment-card-title">{getDisplayName(displayUser)}</h3>
            </div>
            <div className="assignment-card-summary-meta">
              <code className="assignment-card-summary-identifier" title={displayUser.principal_name}>
                {displayUser.principal_name}
              </code>
              <div className="assignment-card-summary-stats">
                <span className="summary-pill neutral" title="User type">
                  <span className="summary-pill-label">Type</span>
                  <span className="summary-pill-value">{capitalize(displayUser.user_type)}</span>
                </span>
                <span
                  className={`summary-pill ${displayUser.is_admin ? 'success' : 'neutral'}`}
                  title="Administrator status"
                >
                  <span className="summary-pill-label">Admin</span>
                  <span className="summary-pill-value">{displayUser.is_admin ? 'Yes' : 'No'}</span>
                </span>
                <span className="summary-pill neutral" title="Groups loaded for this user">
                  <span className="summary-pill-label">Groups</span>
                  <span className="summary-pill-value">
                    {groupCount !== undefined ? groupCount : '—'}
                  </span>
                </span>
                <span className="summary-pill neutral" title="Devices associated with this user">
                  <span className="summary-pill-label">Devices</span>
                  <span className="summary-pill-value">
                    {deviceCount !== undefined ? deviceCount : '—'}
                  </span>
                </span>
              </div>
            </div>
          </div>
        </div>
        <div className="assignment-card-actions">
          {user.user_type === 'local' ? (
            <button
              type="button"
              className="secondary"
              style={{ fontSize: '12px', padding: '4px 12px' }}
            >
              Manage
            </button>
          ) : (
            <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>Cloud managed</span>
          )}
        </div>
      </header>

      <div
        className={`assignment-card-expanded-wrapper${isExpanded ? ' expanded' : ''}`}
        style={{ maxHeight: isExpanded ? `${contentHeight}px` : '0px' }}
      >
        <section
          className="assignment-card-expanded-content"
          id={detailsId}
          ref={contentRef}
          aria-hidden={!isExpanded}
        >
          <div className="assignment-card-expanded-details">
            <UserOverview user={displayUser} />
          </div>
          <div className="assignment-card-body">
            {loading ? (
              <p style={{ margin: 0, color: 'var(--text-muted)', fontSize: '14px' }}>
                Loading details for {getDisplayName(displayUser)}...
              </p>
            ) : error ? (
              <p className="error-text">Failed to load user details: {error}</p>
            ) : details ? (
              <UserDetailsPanel details={details} />
            ) : (
              <p style={{ margin: 0, color: 'var(--text-muted)', fontSize: '14px' }}>
                Select this user to load more details.
              </p>
            )}
          </div>
        </section>
      </div>
    </article>
  );
}

function UserOverview({ user }: { user: DirectoryUser }) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        gap: '16px',
        flexWrap: 'wrap',
        alignItems: 'flex-start',
      }}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', minWidth: '200px' }}>
        <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 600, color: 'var(--text-primary)' }}>
          {getDisplayName(user)}
        </h3>
        <div style={{ fontSize: '14px', color: 'var(--text-muted)' }}>{user.principal_name}</div>
        {user.email && (
          <div style={{ fontSize: '14px', color: 'var(--text-primary)' }}>
            <strong>Email:</strong> {user.email}
          </div>
        )}
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', fontSize: '13px', color: 'var(--text-muted)' }}>
          <span style={{ display: 'inline-flex', gap: '4px' }}>
            <strong>Created:</strong> {formatDateTime(user.created_at)}
          </span>
          {user.synced_at && (
            <span style={{ display: 'inline-flex', gap: '4px', alignItems: 'center' }}>
              <span aria-hidden="true">·</span>
              <strong>Last synced:</strong> {formatDateTime(user.synced_at)}
            </span>
          )}
        </div>
      </div>
      <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap', alignItems: 'center' }}>
        <span className="badge secondary">{capitalize(user.user_type)}</span>
        {user.is_admin && <span className="badge primary">Admin</span>}
        {user.user_type === 'local' && user.is_protected_local && (
          <span className="badge secondary">Protected</span>
        )}
        {user.role_groups && user.role_groups.length > 0 && (
          <span className="badge secondary">Roles: {user.role_groups.join(', ')}</span>
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
    <>
      <section>
        <h4 style={{ marginBottom: '8px' }}>Group Assignments</h4>
        {groups.length === 0 ? (
          <p style={{ margin: 0, color: 'var(--text-muted)', fontSize: '14px' }}>
            This user is not assigned to any groups.
          </p>
        ) : (
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            {groups.map((group) => (
              <span key={group.id} className="badge secondary">
                {group.display_name}
              </span>
            ))}
          </div>
        )}
      </section>

      <section>
        <h4 style={{ marginBottom: '8px' }}>Devices</h4>
        {devices.length === 0 ? (
          <p style={{ margin: 0, color: 'var(--text-muted)', fontSize: '14px' }}>
            No devices have reported this user yet.
          </p>
        ) : (
          <div style={{ overflowX: 'auto' }}>
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
                    <td>{device.serial_number || '—'}</td>
                    <td>
                      <span className="principal-box">{device.machine_id}</span>
                    </td>
                    <td>{device.last_seen ? formatDateTime(device.last_seen) : '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section>
        <h4 style={{ marginBottom: '8px' }}>Recent Events</h4>
        {events.length === 0 ? (
          <p style={{ margin: 0, color: 'var(--text-muted)', fontSize: '14px' }}>
            No recent Santa events recorded for this user.
          </p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {events.map((event) => (
              <div
                key={event.id}
                style={{
                  border: '1px solid var(--border-primary)',
                  borderRadius: '8px',
                  padding: '12px',
                  background: 'var(--bg-secondary)',
                }}
              >
                <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{event.process_path}</div>
                <div style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
                  {event.hostname ? `Host: ${event.hostname}` : 'Host: —'} · {formatEventTime(event.occurred_at)}
                </div>
                <div style={{ fontSize: '13px', color: 'var(--text-primary)', marginTop: '4px' }}>
                  Decision: {event.decision || 'Unknown'}
                  {event.blocked_reason ? ` · Reason: ${event.blocked_reason}` : ''}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      <section>
        <h4 style={{ marginBottom: '8px' }}>Policies Applied</h4>
        {policies.length === 0 ? (
          <p style={{ margin: 0, color: 'var(--text-muted)', fontSize: '14px' }}>
            No policies currently target this user.
          </p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {policies.map((policy) => (
              <div
                key={policy.scope_id}
                style={{
                  border: '1px solid var(--border-primary)',
                  borderRadius: '8px',
                  padding: '12px',
                  background: 'var(--bg-secondary)',
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    gap: '12px',
                    flexWrap: 'wrap',
                    alignItems: 'center',
                  }}
                >
                  <strong>{policy.application_name}</strong>
                  <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                    <span className="badge secondary">{policy.rule_type}</span>
                    <span className="badge secondary">{policy.action.toUpperCase()}</span>
                    {policy.via_group && (
                      <span className="badge secondary">
                        Via group: {policy.target_name || policy.target_id}
                      </span>
                    )}
                  </div>
                </div>
                <div style={{ fontSize: '13px', color: 'var(--text-primary)', marginTop: '4px' }}>
                  Identifier: <span className="principal-box">{policy.identifier}</span>
                </div>
                {!policy.via_group && (
                  <div style={{ fontSize: '13px', color: 'var(--text-muted)', marginTop: '4px' }}>
                    Applied directly to this user.
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </section>
    </>
  );
}
