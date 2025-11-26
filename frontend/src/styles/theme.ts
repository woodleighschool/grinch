import type { PaletteMode, PaletteOptions, ThemeOptions } from "@mui/material";
import { createTheme } from "@mui/material/styles";

// Theme generated with https://muiv6-theme-creator.web.app, using the "Tech Startup" preset.
const basePalette = {
  primary: {
    main: "#d047fd",
    light: "#fd71ff",
    dark: "#8a00c7",
  },
  secondary: {
    main: "#00E5FF",
    light: "#18FFFF",
    dark: "#00B8D4",
  },
  error: {
    main: "#f44336",
    light: "#e57373",
    dark: "#d32f2f",
  },
  warning: {
    main: "#ffa726",
    light: "#ffb74d",
    dark: "#f57c00",
  },
  info: {
    main: "#29b6f6",
    light: "#4fc3f7",
    dark: "#0288d1",
  },
  success: {
    main: "#66bb6a",
    light: "#81c784",
    dark: "#388e3c",
  },
  grey: {
    50: "#fafafa",
    100: "#f5f5f5",
    200: "#eeeeee",
    300: "#e0e0e0",
    400: "#bdbdbd",
    500: "#9e9e9e",
    600: "#757575",
    700: "#616161",
    800: "#424242",
    900: "#212121",
  },
} satisfies PaletteOptions;

const paletteByMode: Record<PaletteMode, PaletteOptions> = {
  dark: {
    ...basePalette,
    mode: "dark",
    background: {
      default: "#121212",
      paper: "#1E1E1E",
    },
    text: {
      primary: "#ffffff",
      secondary: "rgba(255, 255, 255, 0.7)",
      disabled: "rgba(255, 255, 255, 0.5)",
    },
    divider: "rgba(255, 255, 255, 0.12)",
    action: {
      active: "rgba(255, 255, 255, 0.7)",
      hover: "rgba(255, 255, 255, 0.08)",
      selected: "rgba(255, 255, 255, 0.16)",
      disabled: "rgba(255, 255, 255, 0.3)",
      disabledBackground: "rgba(255, 255, 255, 0.12)",
    },
  },
  light: {
    ...basePalette,
    mode: "light",
    background: {
      default: "#ffffff",
      paper: "#f5f5f5",
    },
    text: {
      primary: "rgba(0, 0, 0, 0.87)",
      secondary: "rgba(0, 0, 0, 0.6)",
      disabled: "rgba(0, 0, 0, 0.38)",
    },
    divider: "rgba(0, 0, 0, 0.12)",
    action: {
      active: "rgba(0, 0, 0, 0.54)",
      hover: "rgba(0, 0, 0, 0.04)",
      selected: "rgba(0, 0, 0, 0.08)",
      disabled: "rgba(0, 0, 0, 0.26)",
      disabledBackground: "rgba(0, 0, 0, 0.12)",
    },
  },
};

const baseThemeOptions: Omit<ThemeOptions, "palette"> = {
  typography: {
    fontFamily: '"Inter", sans-serif',
    h1: {
      fontWeight: 700,
    },
    button: {
      textTransform: "none",
    },
  },
  shape: {
    borderRadius: 16,
  },
  transitions: {
    duration: {
      standard: 300,
    },
  },
};

export const createAppTheme = (mode: PaletteMode = "light") =>
  createTheme({
    ...baseThemeOptions,
    palette: paletteByMode[mode],
  });
