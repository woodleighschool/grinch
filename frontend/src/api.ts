export interface ApiUser {
  display_name: string;
}

export interface ExistingApplicationSummary {
  id: string;
  name: string;
}

export interface ApiErrorResponse {
  error?: string;
  message?: string;
  field_errors?: Record<string, string>;
  existing_application?: ExistingApplicationSummary;
}

export class ApiValidationError extends Error {
  constructor(
    message: string,
    public code: string,
    public fieldErrors: Record<string, string>,
    public status: number,
    public existingApplication?: ExistingApplicationSummary,
  ) {
    super(message);
    this.name = "ApiValidationError";
  }
}

export interface EventRecord {
  id: string;
  occurredAt?: string;
  kind: string;
  payload?: Record<string, unknown>;
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
  block_message?: string;
  cel_enabled: boolean;
  cel_expression?: string;
  enabled: boolean;
  assignment_stats?: ApplicationAssignmentStats;
}

export interface ApplicationAssignmentStats {
  allow_scopes: number;
  block_scopes: number;
  cel_scopes: number;
  total_scopes: number;
  allow_users: number;
  block_users: number;
  cel_users: number;
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
  action: "allow" | "block" | "cel";
  created_at: string;
  target_display_name?: string;
  target_description?: string;
  target_upn?: string;
  effective_member_ids: string[];
  effective_member_count?: number;
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
  lastPreflightPayload?: Record<string, unknown>;
  ruleCursor?: string;
  syncCursor?: string;
}

export interface DeviceDetailResponse {
  device: Device;
  primary_user?: DirectoryUser | null;
  recent_blocks: EventRecord[];
  policies: UserPolicy[];
}

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

// Application payloads

export interface ApplicationPayload {
  name: string;
  rule_type: string;
  identifier: string;
  description?: string;
  block_message?: string;
  cel_enabled?: boolean;
  cel_expression?: string;
}

export type ApplicationValidationPayload = ApplicationPayload;
export type ApplicationCreatePayload = ApplicationPayload;

export interface ApplicationUpdatePayload extends Partial<ApplicationPayload> {
  enabled?: boolean;
}

// Scope payloads

export interface ScopePayload {
  target_type: "group" | "user";
  target_id: string;
  action: "allow" | "block" | "cel";
}

export type ScopeValidationRequest = ScopePayload;
export interface ScopeValidationResponse extends ScopePayload {
  application_id: string;
}

const API_BASE = "/api/v1";

function isApiErrorResponse(value: unknown): value is ApiErrorResponse {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  const hasMessage = typeof candidate.message === "string";
  const hasError = typeof candidate.error === "string";
  const hasFieldErrors = candidate.field_errors !== undefined;

  return hasMessage || hasError || hasFieldErrors;
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text();

    try {
      const parsed: unknown = JSON.parse(text);

      if (isApiErrorResponse(parsed)) {
        const errorData = parsed;

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
      }

      throw new Error(text || res.statusText);
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

async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${path}`;
  const res = await fetch(url, { credentials: "include", ...options });
  return handleResponse<T>(res);
}

// Auth

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

// Applications

export async function listApplications(filters: ApplicationFilters = {}): Promise<Application[]> {
  const params = new URLSearchParams();

  if (filters.search?.trim()) params.set("search", filters.search.trim());
  if (filters.rule_type?.trim()) params.set("rule_type", filters.rule_type.trim());
  if (filters.identifier?.trim()) params.set("identifier", filters.identifier.trim());
  if (filters.enabled !== undefined) params.set("enabled", String(filters.enabled));

  const query = params.toString();
  return apiRequest<Application[]>(`/applications${query ? `?${query}` : ""}`);
}

export async function validateApplication(payload: ApplicationValidationPayload): Promise<ValidationSuccess<ApplicationValidationPayload>> {
  return apiRequest<ValidationSuccess<ApplicationValidationPayload>>("/applications/validate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export async function createApplication(payload: ApplicationCreatePayload): Promise<Application> {
  return apiRequest<Application>("/applications", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export async function deleteApplication(appId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/applications/${appId}`, {
    method: "DELETE",
    credentials: "include",
  });

  if (!res.ok && res.status !== 404) {
    throw new Error("Failed to delete application");
  }
}

export async function updateApplication(appId: string, payload: ApplicationUpdatePayload): Promise<Application> {
  return apiRequest<Application>(`/applications/${appId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

// Scopes

export async function listScopes(appId: string, options?: { includeMembers?: boolean }): Promise<ApplicationScope[]> {
  const params = new URLSearchParams();

  if (options?.includeMembers) params.set("include_members", "true");

  const query = params.toString();
  return apiRequest<ApplicationScope[]>(`/applications/${appId}/scopes${query ? `?${query}` : ""}`);
}

export async function getApplicationDetail(appId: string, options?: { includeMembers?: boolean }): Promise<ApplicationDetailResponse> {
  const params = new URLSearchParams();

  if (options?.includeMembers) params.set("include_members", "true");

  const query = params.toString();
  return apiRequest<ApplicationDetailResponse>(`/applications/${appId}${query ? `?${query}` : ""}`);
}

export async function createScope(appId: string, payload: ScopePayload): Promise<ApplicationScope> {
  return apiRequest<ApplicationScope>(`/applications/${appId}/scopes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export async function validateScope(appId: string, payload: ScopeValidationRequest): Promise<ValidationSuccess<ScopeValidationResponse>> {
  return apiRequest<ValidationSuccess<ScopeValidationResponse>>(`/applications/${appId}/scopes/validate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export async function deleteScope(appId: string, scopeId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/applications/${appId}/scopes/${scopeId}`, {
    method: "DELETE",
    credentials: "include",
  });

  if (!res.ok && res.status !== 404) {
    throw new Error("Failed to delete scope");
  }
}

// Groups

export async function listGroups(params: GroupQueryParams = {}): Promise<DirectoryGroup[]> {
  const search = params.search?.trim();
  const url = search ? `/groups?search=${encodeURIComponent(search)}` : "/groups";
  return apiRequest<DirectoryGroup[]>(url);
}

export async function getGroupEffectiveMembers(groupId: string): Promise<GroupEffectiveMembersResponse> {
  return apiRequest<GroupEffectiveMembersResponse>(`/groups/${groupId}/members`);
}

// Users

export async function listUsers(params: UserQueryParams = {}): Promise<DirectoryUser[]> {
  const search = params.search?.trim();
  const url = search ? `/users?search=${encodeURIComponent(search)}` : "/users";
  return apiRequest<DirectoryUser[]>(url);
}

export async function getUserEffectivePolicies(userId: string): Promise<UserEffectivePoliciesResponse> {
  return apiRequest<UserEffectivePoliciesResponse>(`/users/${userId}/policies`);
}

export async function getUserDetails(userId: string): Promise<UserDetailResponse> {
  return apiRequest<UserDetailResponse>(`/users/${userId}`);
}

// Devices

export async function listDevices(params: DeviceQueryParams = {}): Promise<Device[]> {
  const query = new URLSearchParams();

  if (typeof params.limit === "number") query.set("limit", String(params.limit));
  if (typeof params.offset === "number") query.set("offset", String(params.offset));
  if (params.search?.trim()) query.set("search", params.search.trim());

  const qs = query.toString();
  return apiRequest<Device[]>(`/devices${qs ? `?${qs}` : ""}`);
}

export async function getDeviceDetails(deviceId: string): Promise<DeviceDetailResponse> {
  return apiRequest<DeviceDetailResponse>(`/devices/${deviceId}`);
}

// Events

export async function listEvents(limit = 50, offset = 0): Promise<EventRecord[]> {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });

  return apiRequest<EventRecord[]>(`/events?${params.toString()}`);
}

export async function getEventStats(days = 14): Promise<EventStat[]> {
  const params = new URLSearchParams({ days: String(days) });
  return apiRequest<EventStat[]>(`/events/stats?${params.toString()}`);
}

// Status / settings

export async function getStatus(): Promise<AppStatusResponse> {
  return apiRequest<AppStatusResponse>("/status");
}

export async function getSantaConfig(): Promise<SantaConfig> {
  return apiRequest<SantaConfig>("/settings/santa-config");
}
