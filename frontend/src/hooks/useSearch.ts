import { useMemo, useState } from "react";
import Fuse, { IFuseOptions } from "fuse.js";

export interface SearchConfig<T> {
  keys: Array<string | { name: string; weight: number }>;
  threshold?: number;
  includeScore?: boolean;
  includeMatches?: boolean;
  minMatchCharLength?: number;
  shouldSort?: boolean;
}

export function useSearch<T>(items: T[], config: SearchConfig<T>, initialSearchTerm = "") {
  const [searchTerm, setSearchTerm] = useState(initialSearchTerm);

  const fuse = useMemo(() => {
    const options: IFuseOptions<T> = {
      keys: config.keys,
      threshold: config.threshold ?? 0.3,
      includeScore: config.includeScore ?? false,
      includeMatches: config.includeMatches ?? false,
      minMatchCharLength: config.minMatchCharLength ?? 1,
      shouldSort: config.shouldSort ?? true,
      // Additional useful options
      findAllMatches: false,
      location: 0,
      distance: 100,
    };
    return new Fuse(items, options);
  }, [items, config]);

  const filteredItems = useMemo(() => {
    const term = searchTerm.trim();
    if (!term) {
      return items;
    }

    const results = fuse.search(term);
    return results.map((result) => result.item);
  }, [fuse, searchTerm, items]);

  const clearSearch = () => setSearchTerm("");

  return {
    searchTerm,
    setSearchTerm,
    filteredItems,
    clearSearch,
    resultCount: filteredItems.length,
    hasResults: filteredItems.length > 0,
    isSearching: searchTerm.trim().length > 0,
  };
}

// Predefined search configurations for common use cases
export const searchConfigs = {
  users: {
    keys: [
      { name: "display_name", weight: 0.7 },
      { name: "principal_name", weight: 0.6 },
      { name: "given_name", weight: 0.5 },
      { name: "surname", weight: 0.5 },
      { name: "mail", weight: 0.4 },
    ],
    threshold: 0.4,
  },
  devices: {
    keys: [
      { name: "hostname", weight: 0.8 },
      { name: "serial_number", weight: 0.7 },
      { name: "machine_id", weight: 0.6 },
      { name: "primary_user_principal", weight: 0.5 },
      { name: "primary_user_display_name", weight: 0.5 },
    ],
    threshold: 0.3,
  },
  groups: {
    keys: [
      { name: "display_name", weight: 0.8 },
      { name: "description", weight: 0.4 },
    ],
    threshold: 0.4,
  },
};
