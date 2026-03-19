import { LoginPage } from "@/admin/login";
import { darkTheme, lightTheme } from "@/admin/theme";
import { authProvider } from "@/providers/authProvider";
import { dataProvider } from "@/providers/dataProvider";
import executables from "@/resources/executables";
import executionEvents from "@/resources/executionEvents";
import fileAccessEvents from "@/resources/fileAccessEvents";
import groups from "@/resources/groups";
import machines from "@/resources/machines";
import rules from "@/resources/rules";
import users from "@/resources/users";
import GitHubIcon from "@mui/icons-material/GitHub";
import { Box, IconButton, Typography } from "@mui/material";
import type { ComponentProps, ReactElement } from "react";
import { Admin, AppBar, Layout, Resource, TitlePortal, type RaThemeOptions } from "react-admin";

const repoURL = "https://github.com/woodleighschool/grinch";

const AppToolbar = (): ReactElement => (
  <Box sx={{ display: "flex", alignItems: "center", width: "100%", gap: 2 }}>
    <Typography variant="h5">Grinch 🎄</Typography>
    <TitlePortal />
    <Box sx={{ flex: 1 }} />
    <IconButton color="inherit" component="a" href={repoURL} target="_blank" rel="noreferrer">
      <GitHubIcon />
    </IconButton>
  </Box>
);

const AdminAppBar = (): ReactElement => (
  <AppBar toolbar={<AppToolbar />} sx={{ "& .RaUserMenu-userButton": { whiteSpace: "nowrap" } }} />
);

type LayoutProperties = ComponentProps<typeof Layout>;

const AdminLayout = ({ children, ...properties }: LayoutProperties): ReactElement => (
  <Layout {...properties} appBar={AdminAppBar}>
    {children}
  </Layout>
);

export const App = (): ReactElement => (
  <Admin
    dataProvider={dataProvider}
    authProvider={authProvider}
    loginPage={LoginPage}
    theme={lightTheme as RaThemeOptions}
    darkTheme={darkTheme as RaThemeOptions}
    layout={AdminLayout}
    title="Grinch"
    requireAuth
  >
    <Resource name="rules" {...rules} />
    <Resource name="machines" {...machines} />
    <Resource name="executables" {...executables} />
    <Resource name="execution-events" {...executionEvents} />
    <Resource name="file-access-events" {...fileAccessEvents} />
    <Resource name="users" {...users} />
    <Resource name="groups" {...groups} />
    {/* Registered for reference and mutation hooks used from other resource screens. */}
    <Resource name="group-memberships" />
  </Admin>
);
