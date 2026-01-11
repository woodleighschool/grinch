import type { AuthProvider, UserIdentity } from "react-admin";
import { getCurrentUser, isAuthError, loginLocal, logout } from "@/api/auth";

export const authProvider: AuthProvider = {
  login({ username, password }: { username: string; password: string }): Promise<void> {
    return loginLocal(username, password);
  },

  logout(): Promise<void> {
    return logout();
  },

  async checkAuth(): Promise<void> {
    const user = await getCurrentUser();
    if (!user) {
      throw new Error("Not authenticated");
    }
  },

  checkError(error: unknown): Promise<void> {
    if (isAuthError(error)) {
      return Promise.reject(new Error("Not authenticated"));
    }
    return Promise.resolve();
  },

  async getIdentity(): Promise<UserIdentity> {
    const user = await getCurrentUser();
    if (!user) {
      throw new Error("Not authenticated");
    }

    const identity: UserIdentity = {
      id: user.id,
      fullName: user.name ?? "Unknown User",
    };
    if (user.picture) {
      identity.avatar = user.picture;
    }
    return identity;
  },

  getPermissions(): Promise<unknown[]> {
    return Promise.resolve([]);
  },
};
