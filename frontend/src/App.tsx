import { useEffect, useState, Suspense, lazy } from "react";
import { Link, NavLink, Route, Routes } from "react-router-dom";
import { Toaster } from "react-hot-toast";
import { Login } from "./auth";
import { ApiUser } from "./api";
import { useCurrentUser } from "./hooks/useQueries";
import { showErrorToast } from "./utils/toast";
import { Icons } from "./components/Icons";
import { Button } from "./components/Button";

const Dashboard = lazy(() => import("./pages/Dashboard"));
const Applications = lazy(() => import("./pages/Applications"));
const Users = lazy(() => import("./pages/Users"));
const Devices = lazy(() => import("./pages/Devices"));
const Settings = lazy(() => import("./pages/Settings"));
const ApplicationDetails = lazy(() => import("./pages/ApplicationDetails"));
const UserDetails = lazy(() => import("./pages/UserDetails"));

export default function App() {
  const { data: user, isLoading, error } = useCurrentUser();

  useEffect(() => {
    if (error) {
      console.error("Failed to load user:", error);
      showErrorToast("Failed to load user information");
    }
  }, [error]);

  if (isLoading) {
    return (
      <div className="center-page">
        <p className="muted-text">Loading…</p>
      </div>
    );
  }

  if (!user) {
    return <Login onLogin={() => window.location.reload()} />;
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
            <Icons.Brand />
            <span className="brand-text">Grinch</span>
          </Link>
        </div>
        <div className="nav-links">
          <NavLink to="/" end className="nav-link">
            <Icons.Dashboard />
            <span>Dashboard</span>
          </NavLink>
          <NavLink to="/applications" className="nav-link">
            <Icons.Applications />
            <span>Applications</span>
          </NavLink>
          <NavLink to="/users" className="nav-link">
            <Icons.Users />
            <span>Users</span>
          </NavLink>
          <NavLink to="/devices" className="nav-link">
            <Icons.Devices />
            <span>Devices</span>
          </NavLink>
        </div>
        <div className="navbar-actions">
          <NavLink to="/settings" className="settings-link" title="Settings">
            <Icons.Settings />
          </NavLink>
          <div className="user-info">
            <div className="user-avatar">{(user.display_name ?? user.principal_name)?.[0]?.toUpperCase() ?? "U"}</div>
            <span className="user-name">{user.display_name ?? user.principal_name}</span>
          </div>
          <Button variant="ghost" onClick={handleLogout}>
            <Icons.Logout />
            <span>Logout</span>
          </Button>
        </div>
      </nav>
      <main className="main-content">
        <Suspense
          fallback={
            <div className="center-page">
              <p className="muted-text">Loading page…</p>
            </div>
          }
        >
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/applications" element={<Applications />} />
            <Route path="/applications/:appId" element={<ApplicationDetails />} />
            <Route path="/users" element={<Users />} />
            <Route path="/users/:userId" element={<UserDetails />} />
            <Route path="/devices" element={<Devices />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </Suspense>
      </main>
      <Toaster />
    </div>
  );
}
