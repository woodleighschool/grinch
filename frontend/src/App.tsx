import { useEffect, useMemo, useState, Suspense, lazy } from "react";
import { NavLink, Route, Routes, useLocation } from "react-router-dom";
import { useCurrentUser, useStatus } from "./hooks/useQueries";
import {
  AppBar,
  Toolbar,
  Typography,
  Box,
  Container,
  Button,
  Avatar,
  Stack,
  Tabs,
  Tab,
  Divider,
  Snackbar,
  Alert,
  LinearProgress,
  IconButton,
  Link,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import { Logo } from "./components";
import AppsIcon from "@mui/icons-material/Apps";
import DashboardIcon from "@mui/icons-material/Dashboard";
import LogoutIcon from "@mui/icons-material/Logout";
import GroupIcon from "@mui/icons-material/Group";
import DevicesIcon from "@mui/icons-material/Devices";
import EventNoteIcon from "@mui/icons-material/EventNote";
import SettingsIcon from "@mui/icons-material/Settings";
import GitHubIcon from "@mui/icons-material/GitHub";

const Login = lazy(() => import("./pages/Login"));
const Dashboard = lazy(() => import("./pages/Dashboard"));
const Applications = lazy(() => import("./pages/Applications"));
const Users = lazy(() => import("./pages/Users"));
const Devices = lazy(() => import("./pages/Devices"));
const Events = lazy(() => import("./pages/Events"));
const Settings = lazy(() => import("./pages/Settings"));
const ApplicationDetails = lazy(() => import("./pages/ApplicationDetails"));
const UserDetails = lazy(() => import("./pages/UserDetails"));

export default function App() {
  const { data: user, error } = useCurrentUser();
  const { data: status } = useStatus();
  const location = useLocation();
  const [toast, setToast] = useState<{ open: boolean; message: string; severity: "error" | "success" }>({
    open: false,
    message: "",
    severity: "error",
  });
  const [creditsOpen, setCreditsOpen] = useState(false);

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
	if (p.startsWith("/events")) return "/events";
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
  const versionLabel = status?.version?.version;

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
          <Button component={NavLink} to="/" color="inherit" startIcon={<Logo size={32} />}>
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
              <Tab icon={<EventNoteIcon fontSize="small" />} iconPosition="start" label="Events" component={NavLink} to="/events" value="/events" />
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
			<Route path="/events" element={<Events />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </Suspense>
      </Container>

      <AppBar component="footer" position="static" elevation={1}>
        <Container maxWidth="lg">
          <Toolbar disableGutters sx={{ py: 1.5 }}>
            <Stack
              direction={{ xs: "column", sm: "row" }}
              spacing={1}
              alignItems={{ xs: "flex-start", sm: "center" }}
              justifyContent="space-between"
              sx={{ width: "100%" }}
            >
              <Stack direction="row" spacing={1} alignItems="center">
                <IconButton
                  component="a"
                  href="https://github.com/woodleighschool/grinch"
                  target="_blank"
                  rel="noopener noreferrer"
                  size="small"
                  color="inherit"
                >
                  <GitHubIcon fontSize="inherit" />
                </IconButton>

                <Typography variant="caption" color="text.secondary">
                  {versionLabel}
                </Typography>
              </Stack>

              <Button variant="text" size="small" onClick={() => setCreditsOpen(true)}>
                Credits
              </Button>
            </Stack>
          </Toolbar>
        </Container>
      </AppBar>

      <Dialog open={creditsOpen} onClose={() => setCreditsOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Credits</DialogTitle>
        <DialogContent dividers>
          <Typography variant="body2" color="text.secondary">
            <Link href="https://www.flaticon.com/free-icons/grinch" title="grinch icons" target="_blank" rel="noopener noreferrer">
              Grinch icons created by Freepik - Flaticon
            </Link>
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreditsOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>

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
