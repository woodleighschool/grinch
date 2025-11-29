import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import {
  getCurrentUser,
  listApplications,
  getUserDetails,
  listUsers,
  listDevices,
  getDeviceDetails,
  listGroups,
  getSantaConfig,
  listEvents,
  getEventStats,
  listScopes,
  getApplicationDetail,
  createApplication,
  updateApplication,
  deleteApplication,
  createScope,
  deleteScope,
  getStatus,
  type ApiUser,
  type Application,
  type DirectoryUser,
  type DirectoryGroup,
  type Device,
  type EventRecord,
  type EventStat,
  type AppStatusResponse,
  type ApplicationUpdatePayload,
  type ScopePayload,
} from "../api";

// Query Keys
export const queryKeys = {
  users: ["users"] as const,
  user: (id: string) => ["user", id] as const,
  applications: ["applications"] as const,
  application: (id: string) => ["application", id] as const,
  applicationScopes: (appId: string) => ["application", appId, "scopes"] as const,
  applicationDetail: (id: string, includeMembers?: boolean) => ["applicationDetail", id, includeMembers ? "withMembers" : "basic"] as const,
  devices: ["devices"] as const,
  device: (id: string) => ["device", id] as const,
  groups: ["groups"] as const,
  santaConfig: ["santaConfig"] as const,
  currentUser: ["currentUser"] as const,
  events: (params?: { limit?: number; offset?: number }) => ["events", params?.limit ?? 50, params?.offset ?? 0] as const,
  eventStats: (days: number) => ["eventStats", days] as const,
  status: ["status"] as const,
} as const;

// Current User Hook
export function useCurrentUser() {
  return useQuery<ApiUser | null>({
    queryKey: queryKeys.currentUser,
    queryFn: getCurrentUser,
  });
}

export function useStatus() {
  return useQuery<AppStatusResponse>({
    queryKey: queryKeys.status,
    queryFn: () => getStatus(),
    staleTime: 60 * 1000,
  });
}

// Applications Hooks
export function useApplications() {
  return useQuery<Application[]>({
    queryKey: queryKeys.applications,
    queryFn: () => listApplications(),
    placeholderData: keepPreviousData,
  });
}

export function useApplicationScopes(appId: string) {
  return useQuery({
    queryKey: queryKeys.applicationScopes(appId),
    queryFn: () => listScopes(appId),
    enabled: !!appId,
  });
}

export function useApplicationDetail(appId?: string, options?: { includeMembers?: boolean }) {
  const includeMembers = options?.includeMembers;
  return useQuery({
    queryKey: queryKeys.applicationDetail(appId ?? "unknown", includeMembers),
    queryFn: () => {
      if (!appId) throw new Error("Missing application identifier.");
      return getApplicationDetail(appId, options);
    },
    enabled: Boolean(appId),
  });
}

export function useCreateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createApplication,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["applications"] });
    },
  });
}

export function useUpdateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, payload }: { appId: string; payload: ApplicationUpdatePayload }) => updateApplication(appId, payload),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: ["applications"] });
      void queryClient.invalidateQueries({
        predicate: (query) => {
          const key = query.queryKey;
          return Array.isArray(key) && key[0] === "applicationDetail" && key[1] === variables.appId;
        },
      });
    },
  });
}

export function useDeleteApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteApplication,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["applications"] });
    },
  });
}

export function useCreateScope() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, payload }: { appId: string; payload: ScopePayload }) => createScope(appId, payload),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.applicationScopes(variables.appId) });
      void queryClient.invalidateQueries({
        predicate: (query) => {
          const key = query.queryKey;
          return Array.isArray(key) && key[0] === "applicationDetail" && key[1] === variables.appId;
        },
      });
    },
  });
}

export function useDeleteScope() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, scopeId }: { appId: string; scopeId: string }) => deleteScope(appId, scopeId),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.applicationScopes(variables.appId) });
      void queryClient.invalidateQueries({
        predicate: (query) => {
          const key = query.queryKey;
          return Array.isArray(key) && key[0] === "applicationDetail" && key[1] === variables.appId;
        },
      });
    },
  });
}

// Users Hooks
export function useUsers() {
  return useQuery<DirectoryUser[]>({
    queryKey: queryKeys.users,
    queryFn: () => listUsers(),
    placeholderData: keepPreviousData,
  });
}

export function useUserDetails(userId: string) {
  return useQuery({
    queryKey: queryKeys.user(userId),
    queryFn: () => getUserDetails(userId),
    enabled: !!userId,
  });
}

// Devices Hook
export function useDevices() {
  return useQuery<Device[]>({
    queryKey: queryKeys.devices,
    queryFn: () => listDevices(),
    placeholderData: keepPreviousData,
  });
}

export function useDeviceDetails(deviceId: string) {
  return useQuery({
    queryKey: queryKeys.device(deviceId),
    queryFn: () => getDeviceDetails(deviceId),
    enabled: !!deviceId,
  });
}

// Groups Hook
export function useGroups() {
  return useQuery<DirectoryGroup[]>({
    queryKey: queryKeys.groups,
    queryFn: () => listGroups(),
    placeholderData: keepPreviousData,
  });
}

// Event Hooks
export function useBlockedEvents(limit = 50, offset = 0) {
  const query = useQuery<EventRecord[]>({
    queryKey: queryKeys.events({ limit, offset }),
    queryFn: () => listEvents(limit, offset),
    select: (result) => (Array.isArray(result) ? result : []),
  });

  return {
    events: query.data ?? [],
    loading: query.isLoading,
    error: query.error ? (query.error instanceof Error ? query.error.message : String(query.error)) : null,
    refetch: query.refetch,
  };
}

export function useEventStats(days = 14) {
  const query = useQuery<EventStat[]>({
    queryKey: queryKeys.eventStats(days),
    queryFn: () => getEventStats(days),
    select: (result) => (Array.isArray(result) ? result : []),
  });

  return {
    stats: query.data ?? [],
    loading: query.isLoading,
    error: query.error ? (query.error instanceof Error ? query.error.message : String(query.error)) : null,
  };
}

// Santa Config Hook
export function useSantaConfig() {
  return useQuery({
    queryKey: queryKeys.santaConfig,
    queryFn: getSantaConfig,
  });
}
