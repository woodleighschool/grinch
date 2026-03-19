import { getCurrentUser, isAuthError, logout } from "@/api/auth";
import { authApi } from "@/api/authClient";
import type { AuthProvider, UserIdentity } from "react-admin";

type CurrentUser = Awaited<ReturnType<typeof getCurrentUser>>;

let currentUserPromise: Promise<CurrentUser> | undefined;

const clearCurrentUserCache = (): void => {
  currentUserPromise = undefined;
};

const getCachedCurrentUser = (): Promise<CurrentUser> => {
  currentUserPromise ??= getCurrentUser();
  return currentUserPromise;
};

export const authProvider: AuthProvider = {
  async login({ username, password }: { username: string; password: string }): Promise<void> {
    await authApi.loginLocal({ user: username, passwd: password, aud: globalThis.location.origin });
    clearCurrentUserCache();
  },

  async logout(): Promise<void> {
    clearCurrentUserCache();
    await logout();
  },

  async checkAuth(): Promise<void> {
    const user = await getCachedCurrentUser();
    if (!user) {
      throw new Error("Not authenticated");
    }
  },

  checkError(error: unknown): Promise<void> {
    if (isAuthError(error)) {
      clearCurrentUserCache();
      return Promise.reject(new Error("Not authenticated"));
    }
    return Promise.resolve();
  },

  async getIdentity(): Promise<UserIdentity> {
    const user = await getCachedCurrentUser();
    if (!user) {
      throw new Error("Not authenticated");
    }

    return {
      id: user.id,
      fullName: user.name ?? "Unknown User",
      ...(user.picture ? { avatar: user.picture } : {}),
    };
  },

  getPermissions(): Promise<unknown[]> {
    return Promise.resolve([]);
  },
};
