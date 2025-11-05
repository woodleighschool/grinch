import { useEffect, useState } from 'react';
import { Link, NavLink, Route, Routes } from 'react-router-dom';
import Dashboard from './components/Dashboard';
import ApplicationManager from './components/ApplicationManager';
import Users from './components/Users';
import Devices from './components/Devices';
import Settings from './components/Settings';
import Login from './components/Login';
import { ApiUser, getCurrentUser } from './api';

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
    return <div className="app-container"><p style={{ padding: '48px' }}>Loading…</p></div>;
  }

  if (!user) {
    return <Login onLogin={setUser} />;
  }

  async function handleLogout() {
    await fetch('/api/auth/logout', { method: 'POST', credentials: 'include' });
    window.location.href = '/';
  }

  if (user && !user.is_admin) {
    return (
      <div className="app-container" style={{ alignItems: 'center', justifyContent: 'center' }}>
        <div className="card" style={{ maxWidth: '480px' }}>
          <h1>Access Restricted</h1>
          <p>Your account does not have administrator access to the Santa control plane.</p>
          <button className="primary" onClick={handleLogout} style={{ marginTop: '24px' }}>
            Sign out
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="app-container">
      <nav className="navbar">
        <div>
          <Link to="/">🎄 Grinch</Link>
        </div>
        <div className="nav-links">
          <NavLink to="/" end>Dashboard</NavLink>
          <NavLink to="/applications">Applications</NavLink>
          <NavLink to="/users">Users</NavLink>
          <NavLink to="/devices">Devices</NavLink>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <NavLink to="/settings" title="Settings">
            ⚙️
          </NavLink>
          {user.is_admin && <span className="badge success">Admin</span>}
          <span>{user.display_name ?? user.principal_name}</span>
          <button className="secondary" onClick={handleLogout}>Logout</button>
        </div>
      </nav>
      <main className="main-content">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/applications" element={<ApplicationManager />} />
          <Route path="/users" element={<Users />} />
          <Route path="/devices" element={<Devices />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </main>
    </div>
  );
}
