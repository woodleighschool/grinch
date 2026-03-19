import {
  executablesApi,
  executionEventsApi,
  fileAccessEventsApi,
  groupMembershipsApi,
  groupsApi,
  machineRulesApi,
  machinesApi,
  ruleMachinesApi,
  rulesApi,
  usersApi,
} from "@/api/adminClient";
import type { GroupMembershipCreateRequest } from "@/api/types";
import type {
  CreateParams,
  CreateResult,
  DataProvider,
  DeleteManyParams,
  DeleteManyResult,
  DeleteParams,
  DeleteResult,
  GetListParams,
  GetListResult,
  GetManyParams,
  GetManyReferenceParams,
  GetManyReferenceResult,
  GetManyResult,
  GetOneParams,
  GetOneResult,
  Identifier,
  RaRecord,
  UpdateManyParams,
  UpdateManyResult,
  UpdateParams,
  UpdateResult,
} from "react-admin";

type RecordShape = Record<string, unknown>;
interface ListResult {
  data: RaRecord[];
  total: number;
}

const asRecord = (value: unknown): RecordShape =>
  typeof value === "object" && value !== null ? (value as RecordShape) : {};

const getOptionalString = (value: unknown): string | undefined => {
  if (typeof value !== "string") {
    return undefined;
  }

  const trimmed = value.trim();
  return trimmed === "" ? undefined : trimmed;
};

const getSearch = (filter?: RecordShape): string | undefined => getOptionalString(filter?.search);

const getSort = (parameters: GetListParams | GetManyReferenceParams): string | undefined => {
  const field = getOptionalString(parameters.sort?.field);
  return field;
};

const getOrder = (parameters: GetListParams | GetManyReferenceParams): string | undefined =>
  typeof parameters.sort?.order === "string" ? parameters.sort.order.toLowerCase() : undefined;

const asListQuery = (
  parameters: GetListParams | GetManyReferenceParams,
  extra?: Record<string, string | number | undefined>,
): Record<string, string | number | undefined> => {
  const filter = asRecord(parameters.filter);
  const page = parameters.pagination?.page;
  const perPage = parameters.pagination?.perPage;

  return {
    limit: typeof perPage === "number" ? perPage : undefined,
    offset: typeof page === "number" && typeof perPage === "number" ? (page - 1) * perPage : undefined,
    search: getSearch(filter),
    sort: getSort(parameters),
    order: getOrder(parameters),
    ...extra,
  };
};

const toListResult = (payload: { rows: unknown[]; total: number }): ListResult => ({
  data: payload.rows as RaRecord[],
  total: payload.total,
});

const unsupported = (operation: string, resource: string): never => {
  throw new Error(`${operation} not supported for resource: ${resource}`);
};

type ListHandler = (parameters: GetListParams | GetManyReferenceParams, signal?: AbortSignal) => Promise<ListResult>;
type GetOneHandler = (id: Identifier, signal?: AbortSignal) => Promise<RaRecord>;
type CreateHandler = (data: RecordShape) => Promise<RaRecord>;
type UpdateHandler = (id: Identifier, data: RecordShape) => Promise<RaRecord>;
type DeleteHandler = (id: Identifier) => Promise<void>;

type ResourceName =
  | "users"
  | "groups"
  | "group-memberships"
  | "machines"
  | "machine-rules"
  | "executables"
  | "execution-events"
  | "file-access-events"
  | "rule-machines"
  | "rules";

const listHandlers: Record<ResourceName, ListHandler> = {
  users: async (parameters, signal): Promise<ListResult> =>
    toListResult(await usersApi.list(asListQuery(parameters), signal)),

  groups: async (parameters, signal): Promise<ListResult> => {
    return toListResult(await groupsApi.list(asListQuery(parameters), signal));
  },

  "group-memberships": async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await groupMembershipsApi.list(
        asListQuery(parameters, {
          group_id: getOptionalString(filter.group_id),
          user_id: getOptionalString(filter.user_id),
          machine_id: getOptionalString(filter.machine_id),
        }),
        signal,
      ),
    );
  },

  machines: async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await machinesApi.list(
        asListQuery(parameters, {
          user_id: getOptionalString(filter.user_id),
        }),
        signal,
      ),
    );
  },

  "machine-rules": async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await machineRulesApi.list(
        asListQuery(parameters, {
          machine_id: getOptionalString(filter.machine_id),
        }),
        signal,
      ),
    );
  },

  executables: async (parameters, signal): Promise<ListResult> =>
    toListResult(await executablesApi.list(asListQuery(parameters), signal)),

  "execution-events": async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await executionEventsApi.list(
        asListQuery(parameters, {
          machine_id: getOptionalString(filter.machine_id),
          user_id: getOptionalString(filter.user_id),
          executable_id: getOptionalString(filter.executable_id),
        }),
        signal,
      ),
    );
  },

  "file-access-events": async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await fileAccessEventsApi.list(
        asListQuery(parameters, {
          machine_id: getOptionalString(filter.machine_id),
          executable_id: getOptionalString(filter.executable_id),
        }),
        signal,
      ),
    );
  },

  rules: async (parameters, signal): Promise<ListResult> =>
    toListResult(await rulesApi.list(asListQuery(parameters), signal)),

  "rule-machines": async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await ruleMachinesApi.list(
        asListQuery(parameters, {
          rule_id: getOptionalString(filter.rule_id),
        }),
        signal,
      ),
    );
  },
};

const getOneHandlers: Partial<Record<ResourceName, GetOneHandler>> = {
  users: (id, signal): Promise<RaRecord> => usersApi.get(String(id), signal) as Promise<RaRecord>,
  groups: (id, signal): Promise<RaRecord> => groupsApi.get(String(id), signal) as Promise<RaRecord>,
  "group-memberships": (id, signal): Promise<RaRecord> =>
    groupMembershipsApi.get(String(id), signal) as Promise<RaRecord>,
  machines: (id, signal): Promise<RaRecord> => machinesApi.get(String(id), signal) as Promise<RaRecord>,
  executables: (id, signal): Promise<RaRecord> => executablesApi.get(String(id), signal) as Promise<RaRecord>,
  "execution-events": (id, signal): Promise<RaRecord> =>
    executionEventsApi.get(String(id), signal) as Promise<RaRecord>,
  "file-access-events": (id, signal): Promise<RaRecord> =>
    fileAccessEventsApi.get(String(id), signal) as Promise<RaRecord>,
  rules: (id, signal): Promise<RaRecord> => rulesApi.get(String(id), signal) as Promise<RaRecord>,
};

const getGetOneHandler = (resourceName: ResourceName): GetOneHandler => {
  const handler = getOneHandlers[resourceName];
  if (!handler) {
    return unsupported("GetOne", resourceName);
  }

  return handler;
};

const createHandlers: Partial<Record<ResourceName, CreateHandler>> = {
  rules: (data): Promise<RaRecord> => rulesApi.create(data) as Promise<RaRecord>,
  groups: (data): Promise<RaRecord> => groupsApi.create(data) as Promise<RaRecord>,
  "group-memberships": (data): Promise<RaRecord> =>
    groupMembershipsApi.create(data as GroupMembershipCreateRequest) as Promise<RaRecord>,
};

const updateHandlers: Partial<Record<ResourceName, UpdateHandler>> = {
  rules: (id, data): Promise<RaRecord> => rulesApi.update(String(id), data) as Promise<RaRecord>,
  groups: (id, data): Promise<RaRecord> => groupsApi.update(String(id), data) as Promise<RaRecord>,
};

const deleteHandlers: Partial<Record<ResourceName, DeleteHandler>> = {
  "group-memberships": (id): Promise<void> => groupMembershipsApi.delete(String(id)),
  rules: (id): Promise<void> => rulesApi.delete(String(id)),
  groups: (id): Promise<void> => groupsApi.delete(String(id)),
  machines: (id): Promise<void> => machinesApi.delete(String(id)),
  "execution-events": (id): Promise<void> => executionEventsApi.delete(String(id)),
  "file-access-events": (id): Promise<void> => fileAccessEventsApi.delete(String(id)),
};

const isResourceName = (value: string): value is ResourceName => value in listHandlers;

const assertResourceName = (operation: string, resource: string): ResourceName => {
  if (!isResourceName(resource)) {
    unsupported(operation, resource);
  }

  return resource as ResourceName;
};

const getCreateHandler = (resourceName: ResourceName): CreateHandler => {
  const handler = createHandlers[resourceName];
  if (!handler) {
    return unsupported("Create", resourceName);
  }

  return handler;
};

const getUpdateHandler = (resourceName: ResourceName): UpdateHandler => {
  const handler = updateHandlers[resourceName];
  if (!handler) {
    return unsupported("Update", resourceName);
  }

  return handler;
};

const getDeleteHandler = (operation: string, resourceName: ResourceName): DeleteHandler => {
  const handler = deleteHandlers[resourceName];
  if (!handler) {
    return unsupported(operation, resourceName);
  }

  return handler;
};

export const dataProvider: DataProvider = {
  supportAbortSignal: true,

  async getList<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: GetListParams,
  ): Promise<GetListResult<RecordType>> {
    const resourceName = assertResourceName("GetList", resource);
    const handler = listHandlers[resourceName];

    const result = await handler(parameters, parameters.signal);
    return { data: result.data as RecordType[], total: result.total };
  },

  async getOne<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: GetOneParams<RecordType>,
  ): Promise<GetOneResult<RecordType>> {
    const resourceName = assertResourceName("GetOne", resource);
    const handler = getGetOneHandler(resourceName);

    const data = await handler(parameters.id, parameters.signal);
    return { data: data as RecordType };
  },

  async getMany<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: GetManyParams<RecordType>,
  ): Promise<GetManyResult<RecordType>> {
    const resourceName = assertResourceName("GetMany", resource);
    const handler = getGetOneHandler(resourceName);

    const records = await Promise.all(parameters.ids.map((id): Promise<RaRecord> => handler(id, parameters.signal)));
    return { data: records as RecordType[] };
  },

  async getManyReference<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: GetManyReferenceParams,
  ): Promise<GetManyReferenceResult<RecordType>> {
    const resourceName = assertResourceName("GetManyReference", resource);
    const handler = listHandlers[resourceName];

    const result = await handler(
      {
        ...parameters,
        filter: {
          ...asRecord(parameters.filter),
          [parameters.target]: String(parameters.id),
        },
      },
      parameters.signal,
    );

    return { data: result.data as RecordType[], total: result.total };
  },

  async create<
    RecordType extends Omit<RaRecord, "id"> = Omit<RaRecord, "id">,
    ResultRecordType extends RaRecord = RecordType & RaRecord,
  >(resource: string, parameters: CreateParams<RecordType>): Promise<CreateResult<ResultRecordType>> {
    const resourceName = assertResourceName("Create", resource);
    const handler = getCreateHandler(resourceName);

    const created = await handler(asRecord(parameters.data));
    return { data: created as ResultRecordType };
  },

  async update<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: UpdateParams,
  ): Promise<UpdateResult<RecordType>> {
    const resourceName = assertResourceName("Update", resource);
    const handler = getUpdateHandler(resourceName);
    const id: Identifier = parameters.id as Identifier;

    const updated = await handler(id, asRecord(parameters.data));
    return { data: updated as RecordType };
  },

  updateMany<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: UpdateManyParams,
  ): Promise<UpdateManyResult<RecordType>> {
    return Promise.reject(
      new Error(`UpdateMany is not supported for ${resource} (${String(parameters.ids.length)} ids)`),
    );
  },

  async delete<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: DeleteParams<RecordType>,
  ): Promise<DeleteResult<RecordType>> {
    const resourceName = assertResourceName("Delete", resource);
    const handler = getDeleteHandler("Delete", resourceName);

    await handler(parameters.id);

    return {
      data: parameters.previousData ?? ({ id: parameters.id } as RecordType),
    };
  },

  async deleteMany<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: DeleteManyParams<RecordType>,
  ): Promise<DeleteManyResult<RecordType>> {
    const resourceName = assertResourceName("DeleteMany", resource);
    const handler = getDeleteHandler("DeleteMany", resourceName);

    const ids: Identifier[] = [...parameters.ids];
    await Promise.all(ids.map((id: Identifier): Promise<void> => handler(id)));
    return { data: parameters.ids };
  },
};
