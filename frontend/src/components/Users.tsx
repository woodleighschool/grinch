import { useEffect, useMemo, useState } from 'react';
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

  const selectedUserSummary = useMemo(() => {
    if (!selectedUserId) {
      return null;
    }
    return users.find((user) => user.id === selectedUserId) ?? null;
  }, [users, selectedUserId]);

  const fallbackDisplayUser = useMemo(() => {
    if (!selectedUserId) {
      return null;
    }
    if (selectedDetails?.user) {
      return selectedDetails.user;
    }
    if (selectedUserSummary) {
      return selectedUserSummary;
    }
    return users.find((user) => user.id === selectedUserId) ?? null;
  }, [selectedDetails, selectedUserId, selectedUserSummary, users]);

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

      <div className="stat-toolbar" style={{ marginBottom: '24px' }}>
        <div className="stat-bubble success">
          <span className="stat-bubble-value">{users.length}</span>
          <span className="stat-bubble-label">Total Users</span>
        </div>

        <input
          type="text"
          placeholder="Search users..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          style={{
            flex: 1,
            maxWidth: '300px',
            marginLeft: '16px',
          }}
        />
      </div>

      {filteredUsers.length === 0 ? (
        <div
          style={{
            textAlign: 'center',
            padding: '40px',
            color: '#6b7280',
            backgroundColor: '#f9fafb',
            borderRadius: '8px',
          }}
        >
          {searchTerm ? (
            <>
              <h3 style={{ margin: '0 0 8px 0' }}>No users found</h3>
              <p style={{ margin: 0 }}>No users match the search term "{searchTerm}"</p>
            </>
          ) : (
            <>
              <h3 style={{ margin: '0 0 8px 0' }}>No users found</h3>
              <p style={{ margin: 0 }}>No users are available in the directory.</p>
            </>
          )}
        </div>
      ) : (
        <div style={{ maxHeight: '600px', overflowY: 'auto' }}>
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Principal Name</th>
                <th>Type</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredUsers.map((user) => {
                const isSelected = user.id === selectedUserId;
                return (
                  <tr
                    key={user.id}
                    onClick={() => setSelectedUserId((current) => (current === user.id ? null : user.id))}
                    style={{
                      cursor: 'pointer',
                      backgroundColor: isSelected ? '#ecfdf5' : undefined,
                    }}
                  >
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <span>👤</span>
                        <div>
                          <div>{getDisplayName(user)}</div>
                          {user.display_name && user.display_name !== user.principal_name && (
                            <div style={{ fontSize: '12px', color: '#6b7280' }}>{user.principal_name}</div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td>
                      <span className="principal-box">{user.principal_name}</span>
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                        <span className="badge secondary">{capitalize(user.user_type)}</span>
                        {user.is_admin && <span className="badge primary">Admin</span>}
                        {user.user_type === 'local' && user.is_protected_local && (
                          <span className="badge secondary">Protected</span>
                        )}
                      </div>
                    </td>
                    <td>
                      {user.user_type === 'local' ? (
                        <button
                          className="secondary"
                          style={{ fontSize: '12px', padding: '4px 8px' }}
                          onClick={(event) => {
                            event.stopPropagation();
                            alert('User management actions coming soon!');
                          }}
                        >
                          Manage
                        </button>
                      ) : (
                        <span style={{ fontSize: '12px', color: '#6b7280' }}>Cloud managed</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {selectedUserId && (
        <div
          style={{
            marginTop: '24px',
            paddingTop: '24px',
            borderTop: '1px solid #e5e7eb',
          }}
        >
          {detailsLoading && (
            <p style={{ margin: 0, color: '#6b7280' }}>
              Loading details for {fallbackDisplayUser ? getDisplayName(fallbackDisplayUser) : 'selected user'}...
            </p>
          )}
          {detailsError && <p className="error-text">Failed to load user details: {detailsError}</p>}
          {!detailsLoading && !detailsError && selectedDetails && (
            <UserDetailsPanel details={selectedDetails} fallbackUser={fallbackDisplayUser} />
          )}
        </div>
      )}
    </div>
  );
}

interface UserDetailsPanelProps {
  details: UserDetailResponse;
  fallbackUser: DirectoryUser | null;
}

function UserDetailsPanel({ details, fallbackUser }: UserDetailsPanelProps) {
  const user = details.user ?? fallbackUser;
  if (!user) {
    return null;
  }

  const groups = details.groups ?? [];
  const devices = details.devices ?? [];
  const events = details.recent_events ?? [];
  const policies = details.policies ?? [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: '16px', flexWrap: 'wrap' }}>
        <div>
          <h3 style={{ margin: '0 0 4px 0' }}>{getDisplayName(user)}</h3>
          <div style={{ color: '#6b7280', fontSize: '14px' }}>{user.principal_name}</div>
          {user.email && (
            <div style={{ marginTop: '4px', fontSize: '14px', color: '#374151' }}>
              <strong>Email:</strong> {user.email}
            </div>
          )}
          <div style={{ marginTop: '8px', fontSize: '13px', color: '#6b7280' }}>
            <strong>Created:</strong> {formatDateTime(user.created_at)}
            {user.synced_at && (
              <>
                {' · '}
                <strong>Last synced:</strong> {formatDateTime(user.synced_at)}
              </>
            )}
          </div>
        </div>
        <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap', alignItems: 'flex-start' }}>
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

      <section>
        <h4 style={{ marginBottom: '8px' }}>Group Assignments</h4>
        {groups.length === 0 ? (
          <p style={{ margin: 0, color: '#6b7280' }}>This user is not assigned to any groups.</p>
        ) : (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
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
          <p style={{ margin: 0, color: '#6b7280' }}>No devices have reported this user yet.</p>
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
                      <span
                        style={{
                          fontFamily: 'monospace',
                          fontSize: '13px',
                          backgroundColor: '#f3f4f6',
                          padding: '2px 6px',
                          borderRadius: '4px',
                        }}
                      >
                        {device.machine_id}
                      </span>
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
          <p style={{ margin: 0, color: '#6b7280' }}>No recent Santa events recorded for this user.</p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {events.map((event) => (
              <div
                key={event.id}
                style={{
                  border: '1px solid #e5e7eb',
                  borderRadius: '8px',
                  padding: '12px',
                  backgroundColor: '#f9fafb',
                }}
              >
                <div style={{ fontWeight: 600 }}>{event.process_path}</div>
                <div style={{ fontSize: '13px', color: '#6b7280' }}>
                  {event.hostname ? `Host: ${event.hostname}` : 'Host: —'} · {formatEventTime(event.occurred_at)}
                </div>
                <div style={{ fontSize: '13px', color: '#374151', marginTop: '4px' }}>
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
          <p style={{ margin: 0, color: '#6b7280' }}>No policies currently target this user.</p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {policies.map((policy) => (
              <div
                key={policy.scope_id}
                style={{
                  border: '1px solid #e5e7eb',
                  borderRadius: '8px',
                  padding: '12px',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', gap: '12px', flexWrap: 'wrap' }}>
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
                <div style={{ fontSize: '13px', color: '#374151', marginTop: '4px' }}>
                  Identifier:{' '}
                  <span
                    style={{
                      fontFamily: 'monospace',
                      backgroundColor: '#f3f4f6',
                      padding: '2px 6px',
                      borderRadius: '4px',
                    }}
                  >
                    {policy.identifier}
                  </span>
                </div>
                {!policy.via_group && (
                  <div style={{ fontSize: '13px', color: '#6b7280', marginTop: '4px' }}>
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
