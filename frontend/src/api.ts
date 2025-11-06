export interface ApiUser {
    id: string;
    principal_name: string;
    display_name?: string;
    email?: string;
    is_admin: boolean;
}

export interface ApiError {
    error: string;
    message: string;
    existing_application?: {
        id: string;
        name: string;
    };
}

export class ApplicationDuplicateError extends Error {
    constructor(
        message: string,
        public existingApplication: { id: string; name: string },
    ) {
        super(message);
        this.name = "ApplicationDuplicateError";
    }
}

export interface BlockedEvent {
    id: number;
    process_path: string;
    process_hash?: string;
    signer?: string;
    blocked_reason?: string;
    occurred_at: string;
    ingested_at: string;
    application_id?: string;
}

export interface Application {
    id: string;
    name: string;
    rule_type: string;
    identifier: string;
    description?: string;
    enabled: boolean;
}

export interface ApplicationScope {
    id: string;
    application_id: string;
    target_type: "group" | "user";
    target_id: string;
    action: "allow" | "block";
    created_at: string;
}

export interface DirectoryGroup {
    id: string;
    external_id: string;
    display_name: string;
    description?: string;
}

export interface GroupMembership {
    group_id: string;
    user_id: string;
}

export interface DirectoryUser {
    id: string;
    external_id?: string;
    display_name?: string;
    principal_name: string;
    email?: string;
    user_type: "local" | "cloud";
    is_protected_local: boolean;
    is_admin: boolean;
    role_groups?: string[];
    synced_at?: string;
    created_at: string;
    updated_at: string;
}

export interface SAMLSettings {
    enabled: boolean;
    metadata_url?: string;
    entity_id?: string;
    acs_url?: string;
    sp_key_path?: string;
    sp_cert_path?: string;
    name_id_format?: string;
    object_id_attribute?: string;
    upn_attribute?: string;
    email_attribute?: string;
    display_name_attribute?: string;
}

export interface SantaConfig {
    xml: string;
}

export interface Device {
    id: string;
    hostname: string;
    serial_number?: string;
    machine_id: string;
    primary_user_id?: string;
    primary_user_principal?: string;
    primary_user_display_name?: string;
    last_seen?: string;
    os_version?: string;
    os_build?: string;
    model_identifier?: string;
    santa_version?: string;
    client_mode?: string;
    created_at: string;
    updated_at: string;
}

export interface UserEvent {
    id: number;
    host_id?: string;
    hostname?: string;
    application_id?: string;
    process_path: string;
    blocked_reason?: string;
    decision?: string;
    occurred_at: string;
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
    recent_events: UserEvent[];
    policies: UserPolicy[];
}

export interface AuthProviders {
    saml: boolean;
    local: boolean;
}

async function handleResponse<T>(res: Response): Promise<T> {
    if (!res.ok) {
        const text = await res.text();

        // Try to parse as JSON for structured error responses
        try {
            const errorData: ApiError = JSON.parse(text);

            // Handle duplicate identifier error specifically
            if (
                errorData.error === "DUPLICATE_IDENTIFIER" &&
                errorData.existing_application
            ) {
                throw new ApplicationDuplicateError(
                    errorData.message,
                    errorData.existing_application,
                );
            }

            // For other structured errors, throw with the message
            throw new Error(errorData.message || text || res.statusText);
        } catch (parseError) {
            // If it's not JSON or parsing fails, fall back to text
            if (parseError instanceof ApplicationDuplicateError) {
                throw parseError;
            }
            throw new Error(text || res.statusText);
        }
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

export async function listApplications(): Promise<Application[]> {
    const res = await fetch("/api/apps", { credentials: "include" });
    return handleResponse<Application[]>(res);
}

export async function checkApplicationExists(
    identifier: string,
): Promise<Application | null> {
    const res = await fetch(
        `/api/apps/check?identifier=${encodeURIComponent(identifier)}`,
        {
            credentials: "include",
        },
    );
    if (res.status === 404) {
        return null;
    }
    return handleResponse<Application>(res);
}

export async function createApplication(payload: {
    name: string;
    rule_type: string;
    identifier: string;
    description?: string;
}): Promise<Application> {
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

export async function updateApplication(
    appId: string,
    payload: { enabled: boolean },
): Promise<Application> {
    const res = await fetch(`/api/apps/${appId}`, {
        method: "PATCH",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
    });
    return handleResponse<Application>(res);
}

export async function listScopes(appId: string): Promise<ApplicationScope[]> {
    const res = await fetch(`/api/apps/${appId}/scopes`, {
        credentials: "include",
    });
    return handleResponse<ApplicationScope[]>(res);
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

export async function deleteScope(
    appId: string,
    scopeId: string,
): Promise<void> {
    const res = await fetch(`/api/apps/${appId}/scopes/${scopeId}`, {
        method: "DELETE",
        credentials: "include",
    });
    if (!res.ok && res.status !== 404) {
        throw new Error("Failed to delete scope");
    }
}

export async function listGroups(): Promise<DirectoryGroup[]> {
    const res = await fetch("/api/groups", { credentials: "include" });
    return handleResponse<DirectoryGroup[]>(res);
}

export async function listGroupMemberships(): Promise<GroupMembership[]> {
    const res = await fetch("/api/groups/memberships", {
        credentials: "include",
    });
    return handleResponse<GroupMembership[]>(res);
}

export async function listUsers(): Promise<DirectoryUser[]> {
    const res = await fetch("/api/users", { credentials: "include" });
    return handleResponse<DirectoryUser[]>(res);
}

export async function getUserDetails(
    userId: string,
): Promise<UserDetailResponse> {
    const res = await fetch(`/api/users/${userId}`, { credentials: "include" });
    return handleResponse<UserDetailResponse>(res);
}

export async function listDevices(): Promise<Device[]> {
    const res = await fetch("/api/devices", { credentials: "include" });
    return handleResponse<Device[]>(res);
}

export async function listBlocked(): Promise<BlockedEvent[]> {
    const res = await fetch("/api/events/blocked", { credentials: "include" });
    return handleResponse<BlockedEvent[]>(res);
}

export function subscribeBlockedEvents(
    onEvent: (event: BlockedEvent) => void,
): () => void {
    const source = new EventSource("/api/events/blocked/stream", {
        withCredentials: true,
    });
    source.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data) as BlockedEvent;
            onEvent(data);
        } catch (err) {
            console.error("failed to parse event", err);
        }
    };
    source.onerror = () => {
        source.close();
    };
    return () => source.close();
}

// Settings API functions

export async function getSAMLSettings(): Promise<SAMLSettings> {
    const res = await fetch("/api/settings/saml", {
        credentials: "include",
    });
    return handleResponse<SAMLSettings>(res);
}

export async function updateSAMLSettings(
    settings: SAMLSettings,
): Promise<SAMLSettings> {
    const res = await fetch("/api/settings/saml", {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(settings),
    });
    return handleResponse<SAMLSettings>(res);
}

export async function getSantaConfig(): Promise<SantaConfig> {
    const res = await fetch("/api/settings/santa-config", {
        credentials: "include",
    });
    return handleResponse<SantaConfig>(res);
}
