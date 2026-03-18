import { AdminLayout } from "@/admin/layout";
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
import type { ReactElement } from "react";
import { Admin, Resource, type RaThemeOptions } from "react-admin";

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
    <Resource name="group-memberships" />
  </Admin>
);
