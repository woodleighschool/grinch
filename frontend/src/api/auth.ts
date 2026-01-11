import { fetchUtils } from "react-admin";

const authFetch = (url: string, options: fetchUtils.Options = {}): ReturnType<typeof fetchUtils.fetchJson> =>
  fetchUtils.fetchJson(url, {
    ...options,
    credentials: "include",
  });

export interface AuthUser {
  id: string;
  name?: string;
  email?: string;
  picture?: string;
  aud?: string;
  ip?: string;
  attrs?: Record<string, unknown>;
  role?: string;
}

export interface AuthProviders {
  microsoft: boolean;
  local: boolean;
}

export const isAuthError = (error: unknown): boolean => {
  if (typeof error !== "object" || error === null) {
    return false;
  }
  const status = (error as { status?: number }).status;
  return status === 401 || status === 403;
};

export async function getCurrentUser(): Promise<AuthUser | undefined> {
  try {
    const response = await authFetch("/auth/user");
    const json = response.json as unknown;
    if (json == undefined) {
      return undefined;
    }
    return json as AuthUser;
  } catch (error) {
    if (isAuthError(error)) {
      return undefined;
    }
    throw error;
  }
}

export async function loginLocal(username: string, password: string): Promise<void> {
  await authFetch("/auth/local/login?session=1", {
    method: "POST",
    headers: new Headers({ "Content-Type": "application/json" }),
    body: JSON.stringify({
      user: username,
      passwd: password,
      aud: globalThis.location.origin,
    }),
  });
}

export async function logout(): Promise<void> {
  try {
    await authFetch("/auth/logout", {
      method: "POST",
    });
  } catch (error) {
    if (isAuthError(error)) {
      return;
    }
    throw error;
  }
}

export async function listAuthProviders(): Promise<AuthProviders> {
  const response = await authFetch("/auth/list");
  const json = response.json as unknown;
  const providers = Array.isArray(json) ? json : [];
  const normalized = new Set(providers.map((entry): string => String(entry).trim().toLowerCase()));

  return {
    microsoft: normalized.has("microsoft"),
    local: normalized.has("local"),
  };
}
