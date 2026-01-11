import type { ReactElement } from "react";
import { Admin, Resource, radiantDarkTheme, radiantLightTheme } from "react-admin";
import { authProvider } from "@/providers/authProvider";
import { dataProvider } from "@/providers/dataProvider";
import { AdminLayout } from "@/admin/layout";
import { LoginPage } from "@/admin/login";
import events from "@/resources/events";
import groups from "@/resources/groups";
import machines from "@/resources/machines";
import memberships from "@/resources/memberships";
import policies from "@/resources/policies";
import rules from "@/resources/rules";
import users from "@/resources/users";

export const AdminApp = (): ReactElement => (
  <Admin
    dataProvider={dataProvider}
    authProvider={authProvider}
    loginPage={LoginPage}
    theme={radiantLightTheme}
    darkTheme={radiantDarkTheme}
    layout={AdminLayout}
    title="Grinch"
    requireAuth
  >
    <Resource name="rules" {...rules} />
    <Resource name="policies" {...policies} />
    <Resource name="machines" {...machines} />
    <Resource name="events" {...events} />
    <Resource name="users" {...users} />
    <Resource name="groups" {...groups} />
    <Resource name="memberships" {...memberships} />
  </Admin>
);
