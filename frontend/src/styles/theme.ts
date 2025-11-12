import type { PaletteMode, PaletteOptions } from "@mui/material";
import { createTheme } from "@mui/material/styles";

const paletteOverrides: PaletteOptions = {
  primary: {
    main: "#d047fd",
    light: "#fd47d0",
    dark: "#7547fd",
    contrastText: "#fff9ff",
  },
  secondary: {
    main: "#75fd47",
    light: "#fdd047",
    contrastText: "#08050b",
  },
  error: { main: "#fd4775" },
  warning: { main: "#fdd047" },
};

export const createAppTheme = (mode: PaletteMode = "light") =>
  createTheme({
    palette: { ...paletteOverrides, mode },
    shape: { borderRadius: 12 },
  });
