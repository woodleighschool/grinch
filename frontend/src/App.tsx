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
        <div className="navbar-brand">
          <Link to="/" className="brand-link">
            <span className="brand-icon">🎄</span>
            <span className="brand-text">Grinch</span>
          </Link>
        </div>
        <div className="nav-links">
          <NavLink to="/" end className="nav-link">
            <span className="nav-icon">📊</span>
            <span>Dashboard</span>
          </NavLink>
          <NavLink to="/applications" className="nav-link">
            <span className="nav-icon">📱</span>
            <span>Applications</span>
          </NavLink>
          <NavLink to="/users" className="nav-link">
            <span className="nav-icon">👥</span>
            <span>Users</span>
          </NavLink>
          <NavLink to="/devices" className="nav-link">
            <span className="nav-icon">💻</span>
            <span>Devices</span>
          </NavLink>
        </div>
        <div className="navbar-actions">
          <NavLink to="/settings" className="settings-link" title="Settings">
            <span className="settings-icon">⚙️</span>
          </NavLink>
          <div className="user-info">
            <div className="user-avatar">
              {(user.display_name ?? user.principal_name)?.[0]?.toUpperCase() ?? 'U'}
            </div>
            <span className="user-name">{user.display_name ?? user.principal_name}</span>
          </div>
          <button className="logout-btn" onClick={handleLogout}>
            <span className="logout-icon">🚪</span>
            <span>Logout</span>
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
