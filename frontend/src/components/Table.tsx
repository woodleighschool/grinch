import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  flexRender,
  ColumnDef,
  SortingState,
  ColumnFiltersState,
  PaginationState,
} from "@tanstack/react-table";
import { useState, useMemo } from "react";
import { Icons } from "./Icons";
import { Button } from "./Button";

export interface TableProps<TData> {
  data: TData[];
  columns: ColumnDef<TData, any>[];
  globalFilter?: string;
  onGlobalFilterChange?: (value: string) => void;
  pagination?: boolean;
  pageSize?: number;
  sorting?: boolean;
  filtering?: boolean;
}

export function Table<TData>({
  data,
  columns,
  globalFilter = "",
  onGlobalFilterChange,
  pagination = false,
  pageSize = 10,
  sorting: enableSorting = true,
  filtering: enableFiltering = true,
}: TableProps<TData>) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [paginationState, setPaginationState] = useState<PaginationState>({
    pageIndex: 0,
    pageSize,
  });

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnFilters,
      globalFilter,
      pagination: pagination ? paginationState : undefined,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange,
    onPaginationChange: pagination ? setPaginationState : undefined,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: enableSorting ? getSortedRowModel() : undefined,
    getFilteredRowModel: enableFiltering ? getFilteredRowModel() : undefined,
    getPaginationRowModel: pagination ? getPaginationRowModel() : undefined,
    enableSorting: enableSorting,
    enableColumnFilters: enableFiltering,
    enableGlobalFilter: enableFiltering,
  });

  const totalPages = pagination ? table.getPageCount() : 1;
  const currentPage = pagination ? table.getState().pagination.pageIndex + 1 : 1;
  const totalRows = data.length;
  const visibleRows = table.getRowModel().rows.length;

  return (
    <div className="table-container">
      <div style={{ overflowX: "auto" }}>
        <table className="table">
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th key={header.id} style={{ position: "relative" }}>
                    {header.isPlaceholder ? null : (
                      <div
                        className={`table-header ${header.column.getCanSort() ? "sortable" : ""}`}
                        onClick={header.column.getToggleSortingHandler()}
                        style={{
                          cursor: header.column.getCanSort() ? "pointer" : "default",
                          display: "flex",
                          alignItems: "center",
                          gap: "4px",
                          userSelect: "none",
                        }}
                      >
                        {flexRender(header.column.columnDef.header, header.getContext())}
                        {header.column.getCanSort() && (
                          <span style={{ display: "flex", flexDirection: "column", opacity: 0.6 }}>
                            {header.column.getIsSorted() === "asc" ? (
                              <Icons.ChevronUp size={14} />
                            ) : header.column.getIsSorted() === "desc" ? (
                              <Icons.ChevronDown size={14} />
                            ) : (
                              <Icons.ChevronsUpDown size={14} />
                            )}
                          </span>
                        )}
                      </div>
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row) => (
              <tr key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <td key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {pagination && totalPages > 1 && (
        <div className="table-pagination">
          <div className="table-pagination-info">
            Showing {visibleRows} of {totalRows} rows
          </div>
          <div className="table-pagination-controls">
            <Button variant="secondary" size="sm" onClick={() => table.setPageIndex(0)} disabled={!table.getCanPreviousPage()}>
              First
            </Button>
            <Button variant="secondary" size="sm" onClick={() => table.previousPage()} disabled={!table.getCanPreviousPage()}>
              <Icons.ChevronLeft size={16} />
            </Button>
            <span style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              Page {currentPage} of {totalPages}
            </span>
            <Button variant="secondary" size="sm" onClick={() => table.nextPage()} disabled={!table.getCanNextPage()}>
              <Icons.ChevronRight size={16} />
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => table.setPageIndex(table.getPageCount() - 1)}
              disabled={!table.getCanNextPage()}
            >
              Last
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
