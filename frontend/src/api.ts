export interface ApiUser {
  display_name: string;
}

export interface ApiErrorResponse {
  error?: string;
  message?: string;
  field_errors?: Record<string, string>;
  existing_application?: {
    id: string;
    name: string;
  };
}

export class ApiValidationError extends Error {
  constructor(
    message: string,
    public code: string,
    public fieldErrors: Record<string, string>,
    public status: number,
    public existingApplication?: { id: string; name: string },
  ) {
    super(message);
    this.name = "ApiValidationError";
  }
}

export interface EventRecord {
  id: string;
  occurredAt?: string;
  kind: string;
  payload: Record<string, unknown>;
  hostname: string;
  machineId: string;
  email?: string;
  userId?: string;
}

export interface EventStat {
  bucket: string;
  kind: string;
  total: number;
}

export interface Application {
  id: string;
  name: string;
  rule_type: string;
  identifier: string;
  description?: string;
  enabled: boolean;
  assignment_stats?: ApplicationAssignmentStats;
}

export interface ApplicationAssignmentStats {
  allow_scopes: number;
  block_scopes: number;
  total_scopes: number;
  allow_users: number;
  block_users: number;
  total_users: number;
}

export interface ApplicationFilters {
  search?: string;
  rule_type?: string;
  identifier?: string;
  enabled?: boolean;
}

export interface ApplicationScope {
  id: string;
  application_id: string;
  target_type: "group" | "user";
  target_id: string;
  action: "allow" | "block";
  created_at: string;
  target_display_name?: string;
  target_description?: string;
  target_upn?: string;
  effective_member_ids: string[];
  effective_member_count: number;
  effective_members?: DirectoryUser[];
}

export interface ApplicationDetailResponse {
  application: Application;
  scopes: ApplicationScope[];
}

export interface DirectoryGroup {
  id: string;
  displayName: string;
  description?: string;
}

export interface GroupQueryParams {
  search?: string;
}

export interface GroupEffectiveMembersResponse {
  group: DirectoryGroup;
  members: DirectoryUser[];
  member_ids: string[];
  count: number;
}

export interface DirectoryUser {
  id: string;
  upn: string;
  displayName: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface UserQueryParams {
  search?: string;
}

export interface UserEffectivePoliciesResponse {
  user: DirectoryUser;
  policies: UserPolicy[];
}

export interface SantaConfig {
  xml: string;
}

export interface AppStatusResponse {
  status: string;
  version: BuildInfo;
}

export interface BuildInfo {
	version: string;
	gitCommit: string;
	buildDate: string;
}

export interface Device {
  id: string;
  machineIdentifier: string;
  serial: string;
  hostname: string;
  primaryUser?: string;
  clientMode?: string;
  cleanSyncRequested?: boolean;
  lastSeen?: string;
  lastPreflightAt?: string;
  lastPostflightAt?: string;
  lastRulesReceived?: number;
  lastRulesProcessed?: number;
  ruleCursor?: string;
  syncCursor?: string;
}

// export interface DeviceDetailResponse {
//   device: Device;
// 	primaryUser: DirectoryUser;
//   recent_events: DeviceEvent[];
//   policies: DevicePolicy[];
// }

export interface DeviceQueryParams {
  search?: string;
  limit?: number;
  offset?: number;
}

export interface UserPolicy {
  scope_id: string;
  application_id: string;
  application_name: string;
  rule_type: string;
  identifier: string;
  action: string;
  target_type: "user" | "group";
  target_id: string;
  target_name?: string;
  via_group: boolean;
  created_at: string;
}

export interface UserDetailResponse {
  user: DirectoryUser;
  groups: DirectoryGroup[];
  devices: Device[];
  recent_blocks: EventRecord[];
  policies: UserPolicy[];
}

export interface AuthProviders {
  oauth: boolean;
  local: boolean;
}

export interface ValidationSuccess<T> {
  valid: true;
  normalised: T;
}

export interface ScopeValidationRequest {
  application_id: string;
  target_type: "group" | "user";
  target_id: string;
  action: "allow" | "block";
}

export interface ScopeValidationResponse {
  application_id: string;
  target_type: "group" | "user";
  target_id: string;
  action: "allow" | "block";
}

export interface ApplicationValidationPayload {
  name: string;
  rule_type: string;
  identifier: string;
  description?: string;
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text();
    try {
      const errorData: ApiErrorResponse = JSON.parse(text);
      if (errorData.field_errors) {
        throw new ApiValidationError(
          errorData.message || "Validation failed",
          errorData.error || "VALIDATION_FAILED",
          errorData.field_errors,
          res.status,
          errorData.existing_application,
        );
      }
      throw new Error(errorData.message || errorData.error || text || res.statusText);
    } catch (parseError) {
      if (parseError instanceof ApiValidationError) {
        throw parseError;
      }
      throw new Error(text || res.statusText);
    }
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return res.json() as Promise<T>;
}

export async function getCurrentUser(): Promise<ApiUser | null> {
  const res = await fetch("/api/auth/me", {
    credentials: "include",
  });
  if (res.status === 401) {
    return null;
  }
  return handleResponse<ApiUser>(res);
}

export async function getAuthProviders(): Promise<AuthProviders> {
  const res = await fetch("/api/auth/providers", {
    credentials: "include",
  });
  return handleResponse<AuthProviders>(res);
}

export async function listApplications(filters: ApplicationFilters = {}): Promise<Application[]> {
  const params = new URLSearchParams();
  if (filters.search?.trim()) params.set("search", filters.search.trim());
  if (filters.rule_type?.trim()) params.set("rule_type", filters.rule_type.trim());
  if (filters.identifier?.trim()) params.set("identifier", filters.identifier.trim());
  if (filters.enabled !== undefined) params.set("enabled", String(filters.enabled));
  const query = params.toString();
  const res = await fetch(`/api/apps${query ? `?${query}` : ""}`, { credentials: "include" });
  return handleResponse<Application[]>(res);
}

export async function validateApplication(payload: ApplicationValidationPayload): Promise<ValidationSuccess<ApplicationValidationPayload>> {
  const res = await fetch("/api/apps/validate", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return handleResponse<ValidationSuccess<ApplicationValidationPayload>>(res);
}

export async function createApplication(payload: { name: string; rule_type: string; identifier: string; description?: string }): Promise<Application> {
  const res = await fetch("/api/apps", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return handleResponse<Application>(res);
}

export async function deleteApplication(appId: string): Promise<void> {
  const res = await fetch(`/api/apps/${appId}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok && res.status !== 404) {
    throw new Error("Failed to delete application");
  }
}

export async function updateApplication(appId: string, payload: { enabled: boolean }): Promise<Application> {
  const res = await fetch(`/api/apps/${appId}`, {
    method: "PATCH",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return handleResponse<Application>(res);
}

export async function listScopes(appId: string, options?: { includeMembers?: boolean }): Promise<ApplicationScope[]> {
  const params = new URLSearchParams();
  if (options?.includeMembers) params.set("include_members", "true");
  const query = params.toString();
  const res = await fetch(`/api/apps/${appId}/scopes${query ? `?${query}` : ""}`, {
    credentials: "include",
  });
  return handleResponse<ApplicationScope[]>(res);
}

export async function getApplicationDetail(appId: string, options?: { includeMembers?: boolean }): Promise<ApplicationDetailResponse> {
  const params = new URLSearchParams();
  if (options?.includeMembers) params.set("include_members", "true");
  const query = params.toString();
  const res = await fetch(`/api/apps/${appId}${query ? `?${query}` : ""}`, {
    credentials: "include",
  });
  return handleResponse<ApplicationDetailResponse>(res);
}

export async function createScope(
  appId: string,
  payload: {
    target_type: "group" | "user";
    target_id: string;
    action: "allow" | "block";
  },
): Promise<ApplicationScope> {
  const res = await fetch(`/api/apps/${appId}/scopes`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return handleResponse<ApplicationScope>(res);
}

export async function validateScope(payload: ScopeValidationRequest): Promise<ValidationSuccess<ScopeValidationResponse>> {
  const res = await fetch("/api/scopes/validate", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return handleResponse<ValidationSuccess<ScopeValidationResponse>>(res);
}

export async function deleteScope(appId: string, scopeId: string): Promise<void> {
  const res = await fetch(`/api/apps/${appId}/scopes/${scopeId}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok && res.status !== 404) {
    throw new Error("Failed to delete scope");
  }
}

export async function listGroups(params: GroupQueryParams = {}): Promise<DirectoryGroup[]> {
  const search = params.search?.trim();
  const url = search ? `/api/groups?search=${encodeURIComponent(search)}` : "/api/groups";
  const res = await fetch(url, { credentials: "include" });
  return handleResponse<DirectoryGroup[]>(res);
}

export async function getGroupEffectiveMembers(groupId: string): Promise<GroupEffectiveMembersResponse> {
  const res = await fetch(`/api/groups/${groupId}/effective-members`, {
    credentials: "include",
  });
  return handleResponse<GroupEffectiveMembersResponse>(res);
}

export async function listUsers(params: UserQueryParams = {}): Promise<DirectoryUser[]> {
  const search = params.search?.trim();
  const url = search ? `/api/users?search=${encodeURIComponent(search)}` : "/api/users";
  const res = await fetch(url, { credentials: "include" });
  return handleResponse<DirectoryUser[]>(res);
}

export async function getUserEffectivePolicies(userId: string): Promise<UserEffectivePoliciesResponse> {
  const res = await fetch(`/api/users/${userId}/effective-policies`, { credentials: "include" });
  return handleResponse<UserEffectivePoliciesResponse>(res);
}

export async function getUserDetails(userId: string): Promise<UserDetailResponse> {
  const res = await fetch(`/api/users/${userId}`, { credentials: "include" });
  return handleResponse<UserDetailResponse>(res);
}

export async function listDevices(params: DeviceQueryParams = {}): Promise<Device[]> {
  const query = new URLSearchParams();
  if (typeof params.limit === "number") query.set("limit", `${params.limit}`);
  if (typeof params.offset === "number") query.set("offset", `${params.offset}`);
  if (params.search?.trim()) query.set("search", params.search.trim());
  const qs = query.toString();
  const res = await fetch(`/api/machines${qs ? `?${qs}` : ""}`, { credentials: "include" });
  return handleResponse<Device[]>(res);
}

// export async function getDeviceDetails(deviceId: string): Promise<DeviceDetailResponse> {
// 	const res = await fetch(`/api/devices/${deviceId}`, {credentials: "include"});
// 	return handleResponse<DeviceDetailsResponse>(res);
// }

export async function listEvents(limit = 50, offset = 0): Promise<EventRecord[]> {
  const params = new URLSearchParams({ limit: `${limit}`, offset: `${offset}` });
  const res = await fetch(`/api/events?${params.toString()}`, { credentials: "include" });
  return handleResponse<EventRecord[]>(res);
}

export async function getEventStats(days = 14): Promise<EventStat[]> {
  const params = new URLSearchParams({ days: `${days}` });
  const res = await fetch(`/api/events/stats?${params.toString()}`, { credentials: "include" });
  return handleResponse<EventStat[]>(res);
}

export async function getStatus(): Promise<AppStatusResponse> {
  const res = await fetch("/api/status", {
    credentials: "include",
  });
  return handleResponse<AppStatusResponse>(res);
}

export async function getSantaConfig(): Promise<SantaConfig> {
  const res = await fetch("/api/settings/santa-config", {
    credentials: "include",
  });
  return handleResponse<SantaConfig>(res);
}
