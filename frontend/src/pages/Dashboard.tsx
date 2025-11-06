import { useEffect, useMemo, useState } from "react";
import { useBlockedEvents } from "../hooks/useBlockedEvents";
import { useSearch, searchConfigs } from "../hooks/useSearch";
import { formatDateTime } from "../utils/dates";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import type { BlockedEvent, DirectoryGroup, DirectoryUser, ApplicationScope } from "../api";
import { listGroups, listUsers, listApplications, listScopes } from "../api";

function renderRow(event: BlockedEvent) {
  return (
    <tr key={event.id}>
      <td>{formatDateTime(event.occurred_at)}</td>
      <td>{event.process_path}</td>
      <td>{event.signer || "Unsigned"}</td>
      <td>
        <Badge variant="danger">{event.blocked_reason || "Blocked"}</Badge>
      </td>
    </tr>
  );
}

interface DirectoryStatsProps {
  groups: DirectoryGroup[];
  users: DirectoryUser[];
  totalScopes: number;
}

function DirectoryStats({ groups, users, totalScopes }: DirectoryStatsProps) {
  const [activeTab, setActiveTab] = useState<"overview" | "groups" | "users">("overview");

  const {
    searchTerm: groupSearchTerm,
    setSearchTerm: setGroupSearchTerm,
    filteredItems: filteredGroups,
    clearSearch: clearGroupSearch,
    isSearching: isSearchingGroups,
  } = useSearch(groups.slice(0, 20), searchConfigs.groups);

  const {
    searchTerm: userSearchTerm,
    setSearchTerm: setUserSearchTerm,
    filteredItems: filteredUsers,
    clearSearch: clearUserSearch,
    isSearching: isSearchingUsers,
  } = useSearch(users.slice(0, 20), searchConfigs.users);

  // Use the appropriate search state based on active tab
  const searchTerm = activeTab === "groups" ? groupSearchTerm : userSearchTerm;
  const setSearchTerm = activeTab === "groups" ? setGroupSearchTerm : setUserSearchTerm;
  const isSearching = activeTab === "groups" ? isSearchingGroups : isSearchingUsers;

  return (
    <div className="card">
      <h2>Directory Overview</h2>
      <p>Manage access to applications through Entra ID groups and users.</p>

      <div className="stats-row">
        <Badge size="lg" variant="info" value={groups.length} label="Groups" caps />
        <Badge size="lg" variant="success" value={users.length} label="Users" caps />
        <Badge size="lg" variant="warning" value={totalScopes} label="Active Rules" caps />
      </div>

      <div className="toolbar" style={{ marginBottom: "16px" }}>
        {(["overview", "groups", "users"] as const).map((tab) => (
          <Button
            key={tab}
            variant={activeTab === tab ? "primary" : "secondary"}
            onClick={() => setActiveTab(tab)}
            style={{ padding: "8px 16px", fontSize: "14px" }}
          >
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </Button>
        ))}
      </div>

      {activeTab === "overview" && (
        <div className="empty-state">
          <p>Use the Applications page to assign applications to specific groups or users.</p>
          <p>
            <strong>Tip:</strong> Start with groups for broader access control, then use individual users for exceptions.
          </p>
        </div>
      )}

      {(activeTab === "groups" || activeTab === "users") && (
        <>
          <input
            type="text"
            placeholder={`Search ${activeTab}...`}
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            style={{ width: "100%", marginBottom: "16px" }}
          />
          <div style={{ maxHeight: "200px", overflowY: "auto" }}>
            {activeTab === "groups" ? (
              <ul
                style={{
                  margin: 0,
                  padding: 0,
                  listStyle: "none",
                }}
              >
                {filteredGroups.map((group) => (
                  <li
                    key={group.id}
                    style={{
                      padding: "8px 0",
                      borderBottom: "1px solid var(--border-secondary)",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <span>{group.display_name}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <ul
                style={{
                  margin: 0,
                  padding: 0,
                  listStyle: "none",
                }}
              >
                {filteredUsers.map((user) => (
                  <li
                    key={user.id}
                    style={{
                      padding: "8px 0",
                      borderBottom: "1px solid var(--border-secondary)",
                    }}
                  >
                    <div>{user.display_name || user.principal_name}</div>
                    {user.display_name && user.display_name !== user.principal_name && (
                      <div className="muted-text" style={{ fontSize: "12px" }}>
                        {user.principal_name}
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            )}
            {isSearching && (activeTab === "groups" ? filteredGroups : filteredUsers).length === 0 && (
              <div className="empty-state">
                No {activeTab} found matching "{searchTerm}"
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}

export default function Dashboard() {
  const { events, loading, error } = useBlockedEvents();
  const [groups, setGroups] = useState<DirectoryGroup[]>([]);
  const [users, setUsers] = useState<DirectoryUser[]>([]);
  const [totalScopes, setTotalScopes] = useState(0);

  useEffect(() => {
    (async () => {
      try {
        const [groupsData, usersData, appsData] = await Promise.all([listGroups(), listUsers(), listApplications()]);

        setGroups(Array.isArray(groupsData) ? groupsData : []);
        setUsers(Array.isArray(usersData) ? usersData : []);

        // Calculate total scopes across all applications
        const safeApps = Array.isArray(appsData) ? appsData : [];
        const scopeCounts = await Promise.all(
          safeApps.map(async (app) => {
            const data = await listScopes(app.id);
            return Array.isArray(data) ? data.length : 0;
          }),
        );

        setTotalScopes(scopeCounts.reduce((sum, count) => sum + count, 0));
      } catch (err) {
        console.error("Failed to load directory data:", err);
      }
    })();
  }, []);

  return (
    <div style={{ display: "grid", gap: "24px", gridTemplateColumns: "1fr" }}>
      <div>
        <DirectoryStats groups={groups} users={users} totalScopes={totalScopes} />
      </div>

      <div className="card">
        <h2>Real-time Blocked Launches</h2>
        <p>Incoming telemetry from Santa agents appears here instantly.</p>
        {loading && <p>Loading events…</p>}
        {error && <p className="error-text">Failed to load: {error}</p>}
        {!loading && events.length === 0 && <p>No blocked launches recorded yet.</p>}
        {events.length > 0 && (
          <table className="table">
            <thead>
              <tr>
                <th>Occurred</th>
                <th>Process</th>
                <th>Signer</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>{events.map(renderRow)}</tbody>
          </table>
        )}
      </div>
    </div>
  );
}
