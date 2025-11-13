import { Alert, Snackbar, type AlertColor } from "@mui/material";

export type PageToast = {
  open: boolean;
  message: string;
  severity?: AlertColor;
};

interface PageSnackbarProps {
  toast: PageToast;
  onClose: () => void;
  autoHideDuration?: number;
}

export function PageSnackbar({ toast, onClose, autoHideDuration = 4000 }: PageSnackbarProps) {
  return (
    <Snackbar open={toast.open} autoHideDuration={autoHideDuration} onClose={onClose} anchorOrigin={{ vertical: "bottom", horizontal: "center" }}>
      <Alert severity={toast.severity ?? "error"} onClose={onClose} variant="filled">
        {toast.message}
      </Alert>
    </Snackbar>
  );
}
