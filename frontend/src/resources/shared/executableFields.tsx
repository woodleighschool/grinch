import { Stack, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@mui/material";
import type { ReactElement } from "react";
import { ArrayField, DataTable, DateField, useRecordContext } from "react-admin";

const EntitlementValue = ({ value }: { value: unknown }): ReactElement => {
  if (Array.isArray(value)) {
    return (
      <Stack spacing={0.25} sx={{ py: 0.25 }}>
        {(value as unknown[]).map(
          (v, index): ReactElement => (
            <Typography key={index} variant="body2">
              {String(v)}
            </Typography>
          ),
        )}
      </Stack>
    );
  }
  return <Typography variant="body2">{String(value)}</Typography>;
};

export const SigningChainArrayField = (properties: { source: string }): ReactElement => (
  <ArrayField source={properties.source}>
    <DataTable bulkActionButtons={false} rowClick={false}>
      <DataTable.Col source="common_name" label="Common Name" />
      <DataTable.Col source="organization" label="Organization" />
      <DataTable.Col source="organizational_unit" label="Org Unit" />
      <DataTable.Col source="valid_from" label="Valid From">
        <DateField source="valid_from" showTime />
      </DataTable.Col>
      <DataTable.Col source="valid_until" label="Valid Until">
        <DateField source="valid_until" showTime />
      </DataTable.Col>
    </DataTable>
  </ArrayField>
);

export const ExecutableEntitlementsArrayField = (): ReactElement | undefined => {
  const record = useRecordContext<Record<string, unknown>>();
  if (!record) {
    return undefined;
  }

  const entitlements = Object.entries((record.entitlements as Record<string, unknown> | undefined) ?? {}).map(
    ([key, value]): { id: string; key: string; value: unknown } => ({ id: key, key, value }),
  );

  if (entitlements.length === 0) {
    return undefined;
  }

  return (
    <Table size="small">
      <TableHead>
        <TableRow>
          <TableCell>Key</TableCell>
          <TableCell>Value</TableCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {entitlements.map(
          (row): ReactElement => (
            <TableRow key={row.id}>
              <TableCell sx={{ verticalAlign: "top" }}>{row.key}</TableCell>
              <TableCell sx={{ verticalAlign: "top" }}>
                <EntitlementValue value={row.value} />
              </TableCell>
            </TableRow>
          ),
        )}
      </TableBody>
    </Table>
  );
};
