import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ApiUser,
  getCurrentUser,
  listApplications,
  Application,
  DirectoryUser,
  getUserDetails,
  UserDetailResponse,
  listUsers,
  Device,
  listDevices,
  BlockedEvent,
  listBlocked,
  DirectoryGroup,
  listGroups,
  SantaConfig,
  getSantaConfig,
  ApplicationScope,
  listScopes,
  createApplication,
  updateApplication,
  deleteApplication,
  createScope,
  deleteScope,
} from "../api";

// Query Keys
export const queryKeys = {
  users: ["users"] as const,
  user: (id: string) => ["user", id] as const,
  applications: ["applications"] as const,
  application: (id: string) => ["application", id] as const,
  applicationScopes: (appId: string) => ["application", appId, "scopes"] as const,
  devices: ["devices"] as const,
  blockedEvents: ["blockedEvents"] as const,
  groups: ["groups"] as const,
  santaConfig: ["santaConfig"] as const,
  currentUser: ["currentUser"] as const,
} as const;

// Current User Hook
export function useCurrentUser() {
  return useQuery({
    queryKey: queryKeys.currentUser,
    queryFn: getCurrentUser,
    retry: (failureCount, error) => {
      // Don't retry on 401 errors (user not authenticated)
      if (error instanceof Error && error.message.includes("401")) return false;
      return failureCount < 3;
    },
  });
}

// Applications Hooks
export function useApplications() {
  return useQuery({
    queryKey: queryKeys.applications,
    queryFn: listApplications,
  });
}

export function useApplicationScopes(appId: string) {
  return useQuery({
    queryKey: queryKeys.applicationScopes(appId),
    queryFn: () => listScopes(appId),
    enabled: !!appId,
  });
}

export function useCreateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createApplication,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications });
    },
  });
}

export function useUpdateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, payload }: { appId: string; payload: { enabled: boolean } }) => updateApplication(appId, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications });
    },
  });
}

export function useDeleteApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteApplication,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications });
    },
  });
}

export function useCreateScope() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      appId,
      payload,
    }: {
      appId: string;
      payload: { target_type: "group" | "user"; target_id: string; action: "allow" | "block" };
    }) => createScope(appId, payload),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applicationScopes(variables.appId) });
    },
  });
}

export function useDeleteScope() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, scopeId }: { appId: string; scopeId: string }) => deleteScope(appId, scopeId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applicationScopes(variables.appId) });
    },
  });
}

// Users Hooks
export function useUsers() {
  return useQuery({
    queryKey: queryKeys.users,
    queryFn: listUsers,
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
  return useQuery({
    queryKey: queryKeys.devices,
    queryFn: listDevices,
  });
}

// Blocked Events Hook
export function useBlockedEvents() {
  return useQuery({
    queryKey: queryKeys.blockedEvents,
    queryFn: listBlocked,
  });
}

// Groups Hook
export function useGroups() {
  return useQuery({
    queryKey: queryKeys.groups,
    queryFn: listGroups,
  });
}

// Santa Config Hook
export function useSantaConfig() {
  return useQuery({
    queryKey: queryKeys.santaConfig,
    queryFn: getSantaConfig,
  });
}
