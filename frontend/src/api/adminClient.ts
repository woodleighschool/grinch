import { withXsrfHeaders } from "@/api/cookies";
import type { components, paths } from "@/api/openapi";
import createClient from "openapi-fetch";
import { HttpError } from "react-admin";

interface ApiResult<T> {
  data?: T;
  error?: unknown;
  response: Response;
}

type Problem = components["schemas"]["Problem"];
type Executable = components["schemas"]["Executable"];
type ExecutableListResponse = components["schemas"]["ExecutableListResponse"];
type ExecutionEvent = components["schemas"]["ExecutionEvent"];
type ExecutionEventListResponse = components["schemas"]["ExecutionEventListResponse"];
type FileAccessEvent = components["schemas"]["FileAccessEvent"];
type FileAccessEventListResponse = components["schemas"]["FileAccessEventListResponse"];
type Group = components["schemas"]["Group"];
type GroupListResponse = components["schemas"]["GroupListResponse"];
type GroupMembership = components["schemas"]["GroupMembership"];
type GroupMembershipListResponse = components["schemas"]["GroupMembershipListResponse"];
type Machine = components["schemas"]["Machine"];
type MachineListResponse = components["schemas"]["MachineListResponse"];
type MachineRuleListResponse = components["schemas"]["MachineRuleListResponse"];
type Rule = components["schemas"]["Rule"];
type RuleListResponse = components["schemas"]["RuleListResponse"];
type RuleMachineListResponse = components["schemas"]["RuleMachineListResponse"];
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

const isProblem = (value: unknown): value is Problem =>
  typeof value === "object" &&
  value !== null &&
  "detail" in value &&
  typeof (value as { detail?: unknown }).detail === "string" &&
  "status" in value &&
  typeof (value as { status?: unknown }).status === "number";

const problemToBody = (problem: Problem): Record<string, unknown> => {
  const errors = Object.fromEntries(
    (problem.field_errors ?? []).map((fieldError): [string, string] => [fieldError.field, fieldError.message]),
  );

  return {
    ...problem,
    ...(Object.keys(errors).length > 0 ? { errors } : {}),
  };
};

const toHttpError = (error: unknown, response: Response): HttpError => {
  const message =
    isProblem(error) && error.detail.trim() !== "" ? error.detail : response.statusText || "Request failed";

  return new HttpError(message, response.status, isProblem(error) ? problemToBody(error) : error);
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

export const machineRulesApi = {
  list: list<MachineRuleListResponse>("/machine-rules"),
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

export const ruleMachinesApi = {
  list: list<RuleMachineListResponse>("/rule-machines"),
};

export const groupsApi = {
  list: list<GroupListResponse>("/groups"),
  get: getOne<Group>("/groups/{id}"),
  create: createOne<Group>("/groups"),
  update: updateOne<Group>("/groups/{id}"),
  delete: deleteOne("/groups/{id}"),
};

export const groupMembershipsApi = {
  list: list<GroupMembershipListResponse>("/group-memberships"),
  get: getOne<GroupMembership>("/group-memberships/{id}"),
  create: createOne<GroupMembership>("/group-memberships"),
  delete: deleteOne("/group-memberships/{id}"),
};

export const usersApi = {
  list: list<UserListResponse>("/users"),
  get: getOne<User>("/users/{id}"),
};
