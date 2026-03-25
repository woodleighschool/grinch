import { withXsrfHeaders } from "@/api/cookies";
import type { components, paths } from "@/api/openapi";
import createClient from "openapi-fetch";
import { HttpError } from "react-admin";

interface ApiResult<T> {
  data?: T;
  error?: unknown;
  response: Response;
}

type Executable = components["schemas"]["Executable"];
type ExecutableListResponse = components["schemas"]["ExecutableListResponse"];
type ExecutionEvent = components["schemas"]["ExecutionEvent"];
type ExecutionEventListResponse = components["schemas"]["ExecutionEventListResponse"];
type FileAccessEvent = components["schemas"]["FileAccessEvent"];
type FileAccessEventListResponse = components["schemas"]["FileAccessEventListResponse"];
type Group = components["schemas"]["Group"];
type GroupListResponse = components["schemas"]["GroupListResponse"];
type Machine = components["schemas"]["Machine"];
type MachineListResponse = components["schemas"]["MachineListResponse"];
type Rule = components["schemas"]["Rule"];
type RuleListResponse = components["schemas"]["RuleListResponse"];
type User = components["schemas"]["User"];
type UserListResponse = components["schemas"]["UserListResponse"];
type QueryScalar = string | number | boolean;
type QueryParameters = Record<string, QueryScalar | QueryScalar[] | undefined>;
type Compacted<T extends QueryParameters> = { [K in keyof T]?: NonNullable<T[K]> };

const client = createClient<paths>({
  baseUrl: "/api/v1",
  fetch: (request): Promise<Response> =>
    fetch(
      new Request(request, {
        credentials: "include",
        headers: withXsrfHeaders(request.headers),
      }),
    ),
});

const isValidationBody = (value: unknown): value is { errors: Record<string, unknown> } =>
  typeof value === "object" && value !== null && "errors" in value;

const toHttpError = (error: unknown, response: Response): HttpError => {
  if (isValidationBody(error)) {
    const root = (error.errors.root as { serverError?: string } | undefined)?.serverError;
    return new HttpError(root ?? "Validation failed", response.status, error);
  }
  return new HttpError(response.statusText || "Request failed", response.status);
};

const expectBody = async <T>(resultPromise: Promise<ApiResult<T>>): Promise<T> => {
  const { data, error, response } = await resultPromise;

  if (error !== undefined) {
    throw toHttpError(error, response);
  }

  if (data === undefined) {
    throw new HttpError("Empty response", response.status);
  }

  return data;
};

const expectOk = async (resultPromise: Promise<ApiResult<unknown>>): Promise<void> => {
  const { error, response } = await resultPromise;

  if (error !== undefined) {
    throw toHttpError(error, response);
  }
};

const compactQuery = <T extends QueryParameters>(query: T): Compacted<T> => {
  const result: Compacted<T> = {};

  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined) {
      (result as Record<string, unknown>)[key] = value;
    }
  }

  return result;
};

const withQuery = <T extends QueryParameters>(query: T): { params: { query: Compacted<T> } } => ({
  params: { query: compactQuery(query) },
});

const withPath = <T extends string>(id: T): { params: { path: { id: T } } } => ({
  params: { path: { id } },
});

const withGroupUserPath = <T extends string, U extends string>(
  groupID: T,
  userID: U,
): { params: { path: { id: T; user_id: U } } } => ({
  params: { path: { id: groupID, user_id: userID } },
});

const withGroupMachinePath = <T extends string, U extends string>(
  groupID: T,
  machineID: U,
): { params: { path: { id: T; machine_id: U } } } => ({
  params: { path: { id: groupID, machine_id: machineID } },
});

const list =
  <R>(path: keyof paths) =>
  (query: QueryParameters, signal?: AbortSignal): Promise<R> =>
    expectBody(client.GET(path as never, { ...withQuery(query), signal } as never));

const getOne =
  <R>(path: keyof paths) =>
  (id: string, signal?: AbortSignal): Promise<R> =>
    expectBody(client.GET(path as never, { ...withPath(id), signal } as never));

const createOne =
  <R>(path: keyof paths) =>
  (body: unknown): Promise<R> =>
    expectBody(client.POST(path as never, { body } as never));

const updateOne =
  <R>(path: keyof paths) =>
  (id: string, body: unknown): Promise<R> =>
    expectBody(client.PUT(path as never, { ...withPath(id), body } as never));

const deleteOne =
  (path: keyof paths) =>
  (id: string): Promise<void> =>
    expectOk(client.DELETE(path as never, withPath(id) as never));

export const machinesApi = {
  list: list<MachineListResponse>("/machines"),
  get: getOne<Machine>("/machines/{id}"),
  delete: deleteOne("/machines/{id}"),
};

export const executablesApi = {
  list: list<ExecutableListResponse>("/executables"),
  get: getOne<Executable>("/executables/{id}"),
};

export const executionEventsApi = {
  list: list<ExecutionEventListResponse>("/execution-events"),
  get: getOne<ExecutionEvent>("/execution-events/{id}"),
  delete: deleteOne("/execution-events/{id}"),
};

export const fileAccessEventsApi = {
  list: list<FileAccessEventListResponse>("/file-access-events"),
  get: getOne<FileAccessEvent>("/file-access-events/{id}"),
  delete: deleteOne("/file-access-events/{id}"),
};

export const rulesApi = {
  list: list<RuleListResponse>("/rules"),
  get: getOne<Rule>("/rules/{id}"),
  create: createOne<Rule>("/rules"),
  update: updateOne<Rule>("/rules/{id}"),
  delete: deleteOne("/rules/{id}"),
};

export const groupsApi = {
  list: list<GroupListResponse>("/groups"),
  get: getOne<Group>("/groups/{id}"),
  create: createOne<Group>("/groups"),
  update: updateOne<Group>("/groups/{id}"),
  delete: deleteOne("/groups/{id}"),
  addUser: (groupID: string, userID: string): Promise<void> =>
    expectOk(client.PUT("/groups/{id}/users/{user_id}", withGroupUserPath(groupID, userID))),
  removeUser: (groupID: string, userID: string): Promise<void> =>
    expectOk(client.DELETE("/groups/{id}/users/{user_id}", withGroupUserPath(groupID, userID))),
  addMachine: (groupID: string, machineID: string): Promise<void> =>
    expectOk(client.PUT("/groups/{id}/machines/{machine_id}", withGroupMachinePath(groupID, machineID))),
  removeMachine: (groupID: string, machineID: string): Promise<void> =>
    expectOk(client.DELETE("/groups/{id}/machines/{machine_id}", withGroupMachinePath(groupID, machineID))),
};

export const usersApi = {
  list: list<UserListResponse>("/users"),
  get: getOne<User>("/users/{id}"),
};
