import { useEffect, useState, useMemo } from "react";
import { Link } from "react-router-dom";
import { ColumnDef } from "@tanstack/react-table";
import { Icons, Badge, Table, Button } from "../components";
import { useSearch, searchConfigs } from "../hooks/useSearch";
import type { DirectoryUser } from "../api";
import { listUsers } from "../api";

function getDisplayName(user: DirectoryUser): string {
  return user.display_name || user.principal_name;
}

export default function Users() {
  const [users, setUsers] = useState<DirectoryUser[]>([]);
  const [error, setError] = useState<string | null>(null);

  const {
    searchTerm,
    setSearchTerm,
    filteredItems: filteredUsers,
    clearSearch,
    isSearching,
    hasResults,
  } = useSearch(users, searchConfigs.users);

  const columns = useMemo<ColumnDef<DirectoryUser>[]>(
    () => [
      {
        accessorKey: "display_name",
        header: "Name",
        cell: ({ row }) => {
          const displayName = getDisplayName(row.original);
          return (
            <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <Icons.User />
              <span style={{ fontWeight: 500 }}>{displayName}</span>
            </div>
          );
        },
      },
      {
        accessorKey: "principal_name",
        header: "Principal Name",
        cell: ({ getValue }) => <code style={{ fontSize: "13px" }}>{getValue() as string}</code>,
      },
      {
        accessorKey: "user_type",
        header: "Type",
        cell: ({ getValue }) => {
          const userType = getValue() as string;
          return <Badge variant={userType === "cloud" ? "primary" : "secondary"}>{userType === "cloud" ? "CLOUD" : "LOCAL"}</Badge>;
        },
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => (
          <Link to={`/users/${row.original.id}`} style={{ textDecoration: "none" }}>
            <Button variant="secondary" size="sm">
              View Details
            </Button>
          </Link>
        ),
      },
    ],
    [],
  );

  useEffect(() => {
    (async () => {
      try {
        const usersData = await listUsers();
        setUsers(Array.isArray(usersData) ? usersData : []);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("Failed to load users");
        }
      }
    })();
  }, []);

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
        <div style={{ marginBottom: "16px" }}>
          <input
            type="search"
            placeholder="Search users..."
            value={searchTerm}
            onChange={(event) => setSearchTerm(event.target.value)}
            style={{
              maxWidth: "300px",
              marginBottom: "8px",
            }}
            aria-label="Search users"
          />
          {searchTerm && (
            <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
              <span style={{ fontSize: "14px", color: "var(--text-muted)" }}>
                Showing {filteredUsers.length} of {users.length} users
              </span>
              <Button variant="secondary" size="sm" onClick={clearSearch}>
                Clear
              </Button>
            </div>
          )}
        </div>
      )}

      {filteredUsers.length === 0 ? (
        <div className="empty-state">
          {isSearching ? (
            <>
              <h3>No users found</h3>
              <p>
                No users match the search term &quot;
                {searchTerm}&quot;
              </p>
            </>
          ) : (
            <>
              <h3>No users found</h3>
              <p>No users are available in the directory.</p>
            </>
          )}
        </div>
      ) : (
        <Table
          data={filteredUsers}
          columns={columns}
          globalFilter={searchTerm}
          onGlobalFilterChange={setSearchTerm}
          sorting={true}
          filtering={true}
          pagination={users.length > 10}
          pageSize={100}
        />
      )}
    </div>
  );
}
