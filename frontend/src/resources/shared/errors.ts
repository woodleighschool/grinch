export const getErrorMessage = (error: unknown, fallback: string): string =>
  error instanceof Error && error.message.trim() !== "" ? error.message : fallback;
