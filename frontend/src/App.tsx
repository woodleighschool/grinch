import { useEffect, useState } from "react";
import { Link, NavLink, Route, Routes } from "react-router-dom";
import {
  Dashboard,
  Applications,
  Users,
  Devices,
  Settings,
  ApplicationDetails,
  UserDetails,
} from "./pages";
import { Login } from "./auth";
import { ApiUser, getCurrentUser } from "./api";

export default function App() {
  const [user, setUser] = useState<ApiUser | null>();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const me = await getCurrentUser();
        setUser(me);
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  if (loading) {
    return (
      <div className="center-page">
        <p className="muted-text">Loading…</p>
      </div>
    );
  }

  if (!user) {
    return <Login onLogin={setUser} />;
  }

  async function handleLogout() {
    await fetch("/api/auth/logout", {
      method: "POST",
      credentials: "include",
    });
    window.location.href = "/";
  }

  return (
    <div className="app-container">
      <nav className="navbar">
        <div>
          <Link to="/">🎄 Grinch</Link>
        </div>
        <div className="nav-links">
          <NavLink to="/" end>
            Dashboard
          </NavLink>
          <NavLink to="/applications">Applications</NavLink>
          <NavLink to="/users">Users</NavLink>
          <NavLink to="/devices">Devices</NavLink>
        </div>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "12px",
          }}
        >
          <NavLink to="/settings" title="Settings">
            ⚙️
          </NavLink>
          <span>{user.display_name ?? user.principal_name}</span>
          <button className="secondary" onClick={handleLogout}>
            Logout
          </button>
        </div>
      </nav>
      <main className="main-content">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/applications" element={<Applications />} />
          <Route path="/applications/:appId" element={<ApplicationDetails />} />
          <Route path="/users" element={<Users />} />
          <Route path="/users/:userId" element={<UserDetails />} />
          <Route path="/devices" element={<Devices />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </main>
    </div>
  );
}
