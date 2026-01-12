import { getCookie, XSRF_COOKIE_NAME, XSRF_HEADER_NAME } from "@/api/cookies";
import simpleRestProvider from "ra-data-simple-rest";
import { fetchUtils, type DataProvider } from "react-admin";

const API_URL = "/api";

type HttpClient = (url: string, options?: fetchUtils.Options) => ReturnType<typeof fetchUtils.fetchJson>;

const httpClient: HttpClient = (url, options = {}): ReturnType<typeof fetchUtils.fetchJson> => {
  const xsrfToken = getCookie(XSRF_COOKIE_NAME);
  const headers = new Headers(options.headers);

  if (xsrfToken) {
    headers.set(XSRF_HEADER_NAME, xsrfToken);
  }

  return fetchUtils.fetchJson(url, {
    ...options,
    credentials: "include",
    headers,
  });
};

export const dataProvider: DataProvider = simpleRestProvider(API_URL, httpClient);
