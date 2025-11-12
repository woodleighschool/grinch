import { useEffect, useMemo, useState, Suspense, lazy } from "react";
import { NavLink, Route, Routes, useLocation } from "react-router-dom";
import { useCurrentUser } from "./hooks/useQueries";
import { AppBar, Toolbar, Typography, Box, Container, Button, Avatar, Stack, Tabs, Tab, Divider, Snackbar, Alert, LinearProgress } from "@mui/material";
import ParkIcon from "@mui/icons-material/Park";
import AppsIcon from "@mui/icons-material/Apps";
import DashboardIcon from "@mui/icons-material/Dashboard";
import LogoutIcon from "@mui/icons-material/Logout";
import GroupIcon from "@mui/icons-material/Group";
import DevicesIcon from "@mui/icons-material/Devices";
import SettingsIcon from "@mui/icons-material/Settings";

const Login = lazy(() => import("./pages/Login"));
const Dashboard = lazy(() => import("./pages/Dashboard"));
const Applications = lazy(() => import("./pages/Applications"));
const Users = lazy(() => import("./pages/Users"));
const Devices = lazy(() => import("./pages/Devices"));
const Settings = lazy(() => import("./pages/Settings"));
const ApplicationDetails = lazy(() => import("./pages/ApplicationDetails"));
const UserDetails = lazy(() => import("./pages/UserDetails"));

export default function App() {
  const { data: user, error } = useCurrentUser();
  const location = useLocation();
  const [toast, setToast] = useState<{ open: boolean; message: string; severity: "error" | "success" }>({
    open: false,
    message: "",
    severity: "error",
  });

  useEffect(() => {
    if (error) {
      console.error("Failed to load current user", error);
      setToast({ open: true, message: "Failed to load user. Some features may be unavailable.", severity: "error" });
    }
  }, [error]);

  const activeTab = useMemo(() => {
    const p = location.pathname;
    if (p === "/") return "/";
    if (p.startsWith("/applications")) return "/applications";
    if (p.startsWith("/users")) return "/users";
    if (p.startsWith("/devices")) return "/devices";
    return false;
  }, [location.pathname]);

  async function handleLogout() {
    try {
      const res = await fetch("/api/auth/logout", { method: "POST", credentials: "include" });
      if (!res.ok) throw new Error(`Logout failed: ${res.status}`);
      window.location.href = "/";
    } catch (e) {
      console.error("Logout error", e);
      setToast({ open: true, message: "Logout failed. Please try again.", severity: "error" });
    }
  }

  const userDisplay = user?.display_name ?? "Administrator";
  const userInitial = (userDisplay[0] ?? "U").toUpperCase();

  if (!user) {
    return (
      <Suspense
        fallback={
          <Box p={2}>
            <LinearProgress />
          </Box>
        }
      >
        <Login onLogin={() => window.location.reload()} />
      </Suspense>
    );
  }

  return (
    <Box minHeight="100dvh" display="flex" flexDirection="column">
      <AppBar position="sticky" color="primary" enableColorOnDark>
        <Toolbar>
          <Button component={NavLink} to="/" color="inherit" startIcon={<ParkIcon />}>
            <Typography variant="h6" component="span">
              Grinch
            </Typography>
          </Button>

          <Box flex={1} display="flex" justifyContent="center">
            <Tabs value={activeTab} textColor="inherit" indicatorColor="secondary" aria-label="main navigation">
              <Tab icon={<DashboardIcon fontSize="small" />} iconPosition="start" label="Dashboard" component={NavLink} to="/" value="/" />
              <Tab
                icon={<AppsIcon fontSize="small" />}
                iconPosition="start"
                label="Applications"
                component={NavLink}
                to="/applications"
                value="/applications"
              />
              <Tab icon={<GroupIcon fontSize="small" />} iconPosition="start" label="Users" component={NavLink} to="/users" value="/users" />
              <Tab icon={<DevicesIcon fontSize="small" />} iconPosition="start" label="Devices" component={NavLink} to="/devices" value="/devices" />
            </Tabs>
          </Box>

          <Stack direction="row" spacing={1.5} alignItems="center">
            <Button component={NavLink} to="/settings" color="inherit" startIcon={<SettingsIcon />}>
              Settings
            </Button>

            <Button
              color="inherit"
              variant="outlined"
              onClick={handleLogout}
              startIcon={<Avatar sx={{ width: 28, height: 28, fontSize: 13 }}>{userInitial}</Avatar>}
              endIcon={<LogoutIcon fontSize="small" />}
              aria-label={`Logout ${userDisplay}`}
            >
              <Typography variant="body2" noWrap>
                {userDisplay}
              </Typography>
            </Button>
          </Stack>
        </Toolbar>
        <Divider />
      </AppBar>

      <Container component="main" maxWidth={false} sx={{ py: 3, flex: 1 }}>
        <Suspense
          fallback={
            <Box p={2}>
              <LinearProgress />
            </Box>
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
      </Container>

      <Divider />
      <Box component="footer" sx={{ py: 2 }}>
        <Container maxWidth="lg">
          <Typography variant="caption" color="text.secondary">
            {/* TODO: Footer content */}
          </Typography>
        </Container>
      </Box>

      <Snackbar
        open={toast.open}
        autoHideDuration={4000}
        onClose={() => setToast((t) => ({ ...t, open: false }))}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert severity={toast.severity} onClose={() => setToast((t) => ({ ...t, open: false }))} variant="filled">
          {toast.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
