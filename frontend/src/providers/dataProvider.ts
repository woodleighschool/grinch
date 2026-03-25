import {
  executablesApi,
  executionEventsApi,
  fileAccessEventsApi,
  groupsApi,
  machinesApi,
  rulesApi,
  usersApi,
} from "@/api/adminClient";
import type {
  CreateParams,
  CreateResult,
  DataProvider,
  DeleteManyResult,
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
  UpdateManyResult,
  UpdateParams,
  UpdateResult,
} from "react-admin";

type RecordShape = Record<string, unknown>;
type QueryScalar = string | number | boolean;
type QueryValue = QueryScalar | QueryScalar[];
type QueryParameters = Record<string, QueryValue | undefined>;
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

const getOptionalStringArray = (value: unknown): string[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined;
  }

  const result = value
    .filter((item): item is string => typeof item === "string")
    .map((item): string => item.trim())
    .filter((item): boolean => item !== "");

  return result.length > 0 ? result : undefined;
};

const getOptionalBooleanArray = (value: unknown): boolean[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined;
  }

  const result = value.filter((item): item is boolean => typeof item === "boolean");
  return result.length > 0 ? result : undefined;
};

const getOptionalIdentifierArray = (value: unknown): string[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined;
  }

  const result = value
    .filter((item): item is Identifier => typeof item === "string" || typeof item === "number")
    .map((item): string => String(item).trim())
    .filter((item): boolean => item !== "");

  return result.length > 0 ? result : undefined;
};

const getSearch = (filter?: RecordShape): string | undefined => getOptionalString(filter?.search);

const getSort = (parameters: GetListParams | GetManyReferenceParams): string | undefined => {
  const field = getOptionalString(parameters.sort?.field);
  return field;
};

const getOrder = (parameters: GetListParams | GetManyReferenceParams): string | undefined =>
  typeof parameters.sort?.order === "string" ? parameters.sort.order.toLowerCase() : undefined;

const asListQuery = (parameters: GetListParams | GetManyReferenceParams, extra?: QueryParameters): QueryParameters => {
  const filter = asRecord(parameters.filter);
  const page = parameters.pagination?.page;
  const perPage = parameters.pagination?.perPage;

  return {
    limit: typeof perPage === "number" ? perPage : undefined,
    offset: typeof page === "number" && typeof perPage === "number" ? (page - 1) * perPage : undefined,
    search: getSearch(filter),
    sort: getSort(parameters),
    order: getOrder(parameters),
    "ids[]": getOptionalIdentifierArray(filter.ids),
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
type UpdateHandler = (id: Identifier, data: RecordShape) => Promise<RaRecord>;
type DeleteHandler = (id: Identifier) => Promise<void>;

type ResourceName =
  | "users"
  | "groups"
  | "machines"
  | "executables"
  | "execution-events"
  | "file-access-events"
  | "rules";

const listHandlers: Record<ResourceName, ListHandler> = {
  users: async (parameters, signal): Promise<ListResult> =>
    toListResult(await usersApi.list(asListQuery(parameters), signal)),

  groups: async (parameters, signal): Promise<ListResult> =>
    toListResult(await groupsApi.list(asListQuery(parameters), signal)),

  machines: async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await machinesApi.list(
        asListQuery(parameters, {
          user_id: getOptionalString(filter.user_id),
          "rule_sync_status[]": getOptionalStringArray(filter.rule_sync_status),
          "client_mode[]": getOptionalStringArray(filter.client_mode),
        }),
        signal,
      ),
    );
  },

  executables: async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await executablesApi.list(
        asListQuery(parameters, {
          "source[]": getOptionalStringArray(filter.source),
        }),
        signal,
      ),
    );
  },

  "execution-events": async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await executionEventsApi.list(
        asListQuery(parameters, {
          machine_id: getOptionalString(filter.machine_id),
          user_id: getOptionalString(filter.user_id),
          executable_id: getOptionalString(filter.executable_id),
          "decision[]": getOptionalStringArray(filter.decision),
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
          "decision[]": getOptionalStringArray(filter.decision),
        }),
        signal,
      ),
    );
  },

  rules: async (parameters, signal): Promise<ListResult> => {
    const filter = asRecord(parameters.filter);

    return toListResult(
      await rulesApi.list(
        asListQuery(parameters, {
          "enabled[]": getOptionalBooleanArray(filter.enabled),
          "rule_type[]": getOptionalStringArray(filter.rule_type),
        }),
        signal,
      ),
    );
  },
};

const getOneHandlers: Partial<Record<ResourceName, GetOneHandler>> = {
  users: (id, signal): Promise<RaRecord> => usersApi.get(String(id), signal) as Promise<RaRecord>,
  groups: (id, signal): Promise<RaRecord> => groupsApi.get(String(id), signal) as Promise<RaRecord>,
  machines: (id, signal): Promise<RaRecord> => machinesApi.get(String(id), signal) as Promise<RaRecord>,
  executables: (id, signal): Promise<RaRecord> => executablesApi.get(String(id), signal) as Promise<RaRecord>,
  "execution-events": (id, signal): Promise<RaRecord> =>
    executionEventsApi.get(String(id), signal) as Promise<RaRecord>,
  "file-access-events": (id, signal): Promise<RaRecord> =>
    fileAccessEventsApi.get(String(id), signal) as Promise<RaRecord>,
  rules: (id, signal): Promise<RaRecord> => rulesApi.get(String(id), signal) as Promise<RaRecord>,
};

const updateHandlers: Partial<Record<ResourceName, UpdateHandler>> = {
  rules: (id, data): Promise<RaRecord> => rulesApi.update(String(id), data) as Promise<RaRecord>,
  groups: (id, data): Promise<RaRecord> => groupsApi.update(String(id), data) as Promise<RaRecord>,
};

const deleteHandlers: Partial<Record<ResourceName, DeleteHandler>> = {
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
    const handler = getOneHandlers[resourceName];
    if (!handler) {
      return unsupported("GetOne", resourceName);
    }

    const data = await handler(parameters.id, parameters.signal);
    return { data: data as RecordType };
  },

  async getMany<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: GetManyParams<RecordType>,
  ): Promise<GetManyResult<RecordType>> {
    const resourceName = assertResourceName("GetMany", resource);
    const listHandler = listHandlers[resourceName];
    const requestedIds = [...new Set(parameters.ids.map(String))];
    if (requestedIds.length === 0) {
      return { data: [] };
    }

    const result = await listHandler(
      {
        filter: {
          ids: requestedIds,
        },
      } as GetListParams,
      parameters.signal,
    );
    const recordsById = new Map<string, RaRecord>(
      result.data.map((record): [string, RaRecord] => [String(record.id), record]),
    );

    return {
      data: parameters.ids.flatMap((id): RecordType[] => {
        const record = recordsById.get(String(id));
        return record ? ([record as RecordType] satisfies RecordType[]) : [];
      }),
    };
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
    switch (resourceName) {
      case "rules": {
        return { data: (await rulesApi.create(asRecord(parameters.data))) as unknown as ResultRecordType };
      }
      case "groups": {
        return { data: (await groupsApi.create(asRecord(parameters.data))) as unknown as ResultRecordType };
      }
      default: {
        return unsupported("Create", resourceName);
      }
    }
  },

  async update<RecordType extends RaRecord = RaRecord>(
    resource: string,
    parameters: UpdateParams,
  ): Promise<UpdateResult<RecordType>> {
    const resourceName = assertResourceName("Update", resource);
    const handler = updateHandlers[resourceName];
    if (!handler) {
      return unsupported("Update", resourceName);
    }
    const id: Identifier = parameters.id as Identifier;

    const updated = await handler(id, asRecord(parameters.data));
    return { data: updated as RecordType };
  },

  updateMany(resource: string, parameters): Promise<UpdateManyResult> {
    return Promise.reject(
      new Error(`UpdateMany is not supported for ${resource} (${String(parameters.ids.length)} ids)`),
    );
  },

  async delete(resource: string, parameters): Promise<DeleteResult> {
    const resourceName = assertResourceName("Delete", resource);
    const handler = deleteHandlers[resourceName];
    if (!handler) {
      return unsupported("Delete", resourceName);
    }

    await handler(parameters.id);
    if (!parameters.previousData) {
      throw new Error(`Delete is missing previousData for ${resourceName}`);
    }

    return {
      data: parameters.previousData,
    };
  },

  async deleteMany(resource: string, parameters): Promise<DeleteManyResult> {
    const resourceName = assertResourceName("DeleteMany", resource);
    const handler = deleteHandlers[resourceName];
    if (!handler) {
      return unsupported("DeleteMany", resourceName);
    }

    await Promise.all(parameters.ids.map((id: Identifier): Promise<void> => handler(id)));
    return { data: parameters.ids };
  },
};
