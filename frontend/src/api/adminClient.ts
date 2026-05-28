import { withXsrfHeaders } from "@/api/cookies";
import type {
  Executable,
  ExecutableListResponse,
  ExecutionEvent,
  ExecutionEventListResponse,
  FileAccessEvent,
  FileAccessEventListResponse,
  Group,
  GroupListResponse,
  Machine,
  MachineListResponse,
  MachineRuleListResponse,
  Membership,
  MembershipListResponse,
  Rule,
  RuleListResponse,
  RuleMachineListResponse,
  User,
  UserListResponse,
} from "@/api/openapi";
import { HttpError } from "react-admin";

interface ApiResult<T> {
  data?: T;
  error?: unknown;
  response: Response;
}

type QueryScalar = string | number | boolean;
type QueryParameters = Record<string, QueryScalar | QueryScalar[] | undefined>;
type Compacted<T extends QueryParameters> = { [K in keyof T]?: NonNullable<T[K]> };

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

const appendQuery = (url: URL, query?: QueryParameters): void => {
  if (!query) {
    return;
  }

  for (const [key, value] of Object.entries(compactQuery(query))) {
    if (Array.isArray(value)) {
      for (const item of value) {
        url.searchParams.append(key, String(item));
      }
    } else {
      url.searchParams.set(key, String(value));
    }
  }
};

const parseResponseBody = async (response: Response): Promise<unknown> => {
  if (response.status === 204 || response.headers.get("Content-Length") === "0") {
    return undefined;
  }

  const text = await response.text();
  if (text === "") {
    return undefined;
  }

  const contentType = response.headers.get("Content-Type") ?? "";
  return contentType.includes("application/json") ? JSON.parse(text) : text;
};

const request = async <T>(
  method: string,
  path: string,
  options: { body?: unknown; query?: QueryParameters; signal?: AbortSignal } = {},
): Promise<ApiResult<T>> => {
  const url = new URL(`/api/v1${path}`, globalThis.location.origin);
  appendQuery(url, options.query);

  const headers = withXsrfHeaders(options.body === undefined ? undefined : { "Content-Type": "application/json" });
  const init: RequestInit = {
    credentials: "include",
    headers,
    method,
    ...(options.body === undefined ? {} : { body: JSON.stringify(options.body) }),
    ...(options.signal ? { signal: options.signal } : {}),
  };
  const response = await fetch(url, init);
  const payload = await parseResponseBody(response);

  if (!response.ok) {
    return { error: payload, response };
  }

  return { data: payload as T, response };
};

const withId = (path: string, id: string): string => path.replace("{id}", encodeURIComponent(id));

const list =
  <R>(path: string) =>
  (query: QueryParameters, signal?: AbortSignal): Promise<R> =>
    expectBody(request<R>("GET", path, { query, ...(signal ? { signal } : {}) }));

const getOne =
  <R>(path: string) =>
  (id: string, signal?: AbortSignal): Promise<R> =>
    expectBody(request<R>("GET", withId(path, id), signal ? { signal } : {}));

const createOne =
  <R>(path: string) =>
  (body: unknown): Promise<R> =>
    expectBody(request<R>("POST", path, { body }));

const updateOne =
  <R>(path: string) =>
  (id: string, body: unknown): Promise<R> =>
    expectBody(request<R>("PUT", withId(path, id), { body }));

const deleteOne =
  (path: string) =>
  (id: string): Promise<void> =>
    expectOk(request<unknown>("DELETE", withId(path, id)));

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

export const membershipsApi = {
  list: list<MembershipListResponse>("/memberships"),
  get: getOne<Membership>("/memberships/{id}"),
  create: createOne<Membership>("/memberships"),
  delete: deleteOne("/memberships/{id}"),
};

export const usersApi = {
  list: list<UserListResponse>("/users"),
  get: getOne<User>("/users/{id}"),
};
