import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import {
  getCurrentUser,
  listApplications,
  getUserDetails,
  listUsers,
  listDevices,
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
  type Application,
  type ApplicationFilters,
  type DirectoryUser,
  type DirectoryGroup,
  type UserQueryParams,
  type Device,
  type DeviceQueryParams,
  type GroupQueryParams,
  type EventRecord,
  type EventStat,
  type AppStatusResponse,
} from "../api";

// Query Keys
export const queryKeys = {
  users: (filters?: UserQueryParams) => ["users", filters?.search ?? ""] as const,
  user: (id: string) => ["user", id] as const,
  applications: (filters?: ApplicationFilters) =>
    [
      "applications",
      filters?.search ?? "",
      filters?.rule_type ?? "",
      filters?.identifier ?? "",
      filters?.enabled === undefined ? "all" : String(filters.enabled),
    ] as const,
  application: (id: string) => ["application", id] as const,
  applicationScopes: (appId: string) => ["application", appId, "scopes"] as const,
  applicationDetail: (id: string, includeMembers?: boolean) => ["applicationDetail", id, includeMembers ? "withMembers" : "basic"] as const,
  devices: (filters?: DeviceQueryParams) =>
    [
      "devices",
      filters?.search ?? "",
      typeof filters?.limit === "number" ? filters.limit : "default",
      typeof filters?.offset === "number" ? filters.offset : 0,
    ] as const,
  groups: (filters?: GroupQueryParams) => ["groups", filters?.search ?? ""] as const,
  santaConfig: ["santaConfig"] as const,
  currentUser: ["currentUser"] as const,
  events: (params?: { limit?: number; offset?: number }) => ["events", params?.limit ?? 50, params?.offset ?? 0] as const,
  eventStats: (days: number) => ["eventStats", days] as const,
  status: ["status"] as const,
} as const;

// Current User Hook
export function useCurrentUser() {
  return useQuery({
    queryKey: queryKeys.currentUser,
    queryFn: getCurrentUser,
  });
}

export function useStatus() {
  return useQuery<AppStatusResponse>({
    queryKey: queryKeys.status,
    queryFn: getStatus,
    staleTime: 60 * 1000,
  });
}

// Applications Hooks
export function useApplications(filters: ApplicationFilters = {}) {
  return useQuery<Application[]>({
    queryKey: queryKeys.applications(filters),
    queryFn: () => listApplications(filters),
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
  return useQuery({
    queryKey: appId
      ? queryKeys.applicationDetail(appId, options?.includeMembers)
      : ["applicationDetail", "unknown", options?.includeMembers ? "withMembers" : "basic"],
    queryFn: () => {
      if (!appId) throw new Error("Missing application identifier.");
      return getApplicationDetail(appId, options);
    },
    enabled: !!appId,
  });
}

export function useCreateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createApplication,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["applications"] });
    },
  });
}

export function useUpdateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, payload }: { appId: string; payload: { enabled: boolean } }) => updateApplication(appId, payload),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["applications"] });
      if (variables?.appId) {
        queryClient.invalidateQueries({
          predicate: (query) => {
            const key = query.queryKey;
            return Array.isArray(key) && key[0] === "applicationDetail" && key[1] === variables.appId;
          },
        });
      }
    },
  });
}

export function useDeleteApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteApplication,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["applications"] });
    },
  });
}

export function useCreateScope() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, payload }: { appId: string; payload: { target_type: "group" | "user"; target_id: string; action: "allow" | "block" } }) =>
      createScope(appId, payload),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applicationScopes(variables.appId) });
      queryClient.invalidateQueries({
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
      queryClient.invalidateQueries({ queryKey: queryKeys.applicationScopes(variables.appId) });
      queryClient.invalidateQueries({
        predicate: (query) => {
          const key = query.queryKey;
          return Array.isArray(key) && key[0] === "applicationDetail" && key[1] === variables.appId;
        },
      });
    },
  });
}

// Users Hooks
export function useUsers(filters: UserQueryParams = {}) {
  return useQuery<DirectoryUser[]>({
    queryKey: queryKeys.users(filters),
    queryFn: () => listUsers(filters),
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
export function useDevices(filters: DeviceQueryParams = {}) {
  return useQuery<Device[]>({
    queryKey: queryKeys.devices(filters),
    queryFn: () => listDevices(filters),
    placeholderData: keepPreviousData,
  });
}

// Groups Hook
export function useGroups(filters: GroupQueryParams = {}) {
  return useQuery<DirectoryGroup[]>({
    queryKey: queryKeys.groups(filters),
    queryFn: () => listGroups(filters),
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
