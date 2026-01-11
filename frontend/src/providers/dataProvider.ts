import { fetchUtils, type DataProvider } from "react-admin";
import simpleRestProvider from "ra-data-simple-rest";

const API_URL = "/api";

type HttpClient = (url: string, options?: fetchUtils.Options) => ReturnType<typeof fetchUtils.fetchJson>;

const httpClient: HttpClient = (url, options = {}): ReturnType<typeof fetchUtils.fetchJson> =>
  fetchUtils.fetchJson(url, {
    ...options,
    credentials: "include",
  });

export const dataProvider: DataProvider = simpleRestProvider(API_URL, httpClient);
