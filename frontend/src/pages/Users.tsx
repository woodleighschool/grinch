import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import type { DirectoryUser } from "../api";
import { listUsers } from "../api";

function getDisplayName(user: DirectoryUser): string {
    return user.display_name || user.principal_name;
}

export default function Users() {
    const [users, setUsers] = useState<DirectoryUser[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchTerm, setSearchTerm] = useState("");

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
                    setError("Failed to load users");
                }
            } finally {
                setLoading(false);
            }
        })();
    }, []);

    const filteredUsers = useMemo(() => {
        const term = searchTerm.trim().toLowerCase();
        if (!term) {
            return users;
        }
        return users.filter((user) => {
            const display = getDisplayName(user).toLowerCase();
            const principal = user.principal_name.toLowerCase();
            return display.includes(term) || principal.includes(term);
        });
    }, [users, searchTerm]);

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
            <p>
                Manage individual user access and permissions from your Entra ID
                directory.
            </p>

            {users.length > 0 && (
                <div
                    style={{
                        display: "flex",
                        flexWrap: "wrap",
                        gap: "12px",
                        alignItems: "center",
                        justifyContent: "space-between",
                        marginBottom: "16px",
                    }}
                >
                    <div
                        style={{
                            display: "flex",
                            gap: "8px",
                            flex: "1 1 260px",
                            minWidth: "220px",
                        }}
                    >
                        <input
                            type="search"
                            placeholder="Search users..."
                            value={searchTerm}
                            onChange={(event) =>
                                setSearchTerm(event.target.value)
                            }
                            style={{ flex: 1 }}
                            aria-label="Search users"
                        />
                        {searchTerm && (
                            <button
                                type="button"
                                className="secondary"
                                onClick={() => setSearchTerm("")}
                                title="Clear search"
                                style={{ whiteSpace: "nowrap" }}
                            >
                                Clear
                            </button>
                        )}
                    </div>
                    <div
                        style={{
                            color: "var(--text-muted)",
                            fontSize: "14px",
                            marginLeft: "auto",
                            textAlign: "right",
                        }}
                    >
                        Showing {filteredUsers.length} of {users.length} user
                        {users.length !== 1 ? "s" : ""}
                    </div>
                </div>
            )}

            {filteredUsers.length === 0 ? (
                <div className="empty-state">
                    {searchTerm ? (
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
                filteredUsers.map((user) => {
                    const displayName = getDisplayName(user);
                    return (
                        <Link
                            key={user.id}
                            to={`/users/${user.id}`}
                            style={{ textDecoration: "none", color: "inherit" }}
                        >
                            <article
                                className="assignment-card"
                                style={{ cursor: "pointer" }}
                            >
                                <div
                                    className="assignment-card-header"
                                    style={{ alignItems: "flex-start" }}
                                >
                                    <div
                                        className="assignment-card-summary"
                                        style={{ cursor: "pointer" }}
                                    >
                                        <span
                                            className="assignment-card-icon"
                                            aria-hidden="true"
                                        >
                                            👤
                                        </span>
                                        <div className="assignment-card-summary-main">
                                            <div className="assignment-card-summary-title">
                                                <h3 className="assignment-card-title">
                                                    {displayName}
                                                </h3>
                                                <span
                                                    className={`rule-chip rule-chip-${user.user_type === "cloud" ? "cloud" : "local"}`}
                                                    title={`${user.user_type === "cloud" ? "Cloud" : "Local"} user`}
                                                >
                                                    {user.user_type === "cloud"
                                                        ? "CLOUD"
                                                        : "LOCAL"}
                                                </span>
                                            </div>
                                            <div className="assignment-card-summary-meta">
                                                <code
                                                    className="assignment-card-summary-identifier"
                                                    title={user.principal_name}
                                                >
                                                    {user.principal_name}
                                                </code>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </article>
                        </Link>
                    );
                })
            )}
        </div>
    );
}
