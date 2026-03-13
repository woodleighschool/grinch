import type { ThemeOptions } from "@mui/material/styles";
import { deepmerge } from "@mui/utils";
import { defaultDarkTheme, defaultLightTheme } from "react-admin";

const sharedOverrides: ThemeOptions = {
  shape: {
    borderRadius: 10,
  },
  typography: {
    fontFamily: [
      "Inter",
      "ui-sans-serif",
      "system-ui",
      "-apple-system",
      "BlinkMacSystemFont",
      '"Segoe UI"',
      "sans-serif",
    ].join(", "),
    button: {
      textTransform: "none",
      fontWeight: 600,
    },
  },
  components: {
    MuiButtonBase: {
      defaultProps: {
        disableRipple: true,
      },
    },
    MuiButton: {
      defaultProps: {
        variant: "contained",
      },
      styleOverrides: {
        root: {
          borderRadius: 999,
          boxShadow: "none",
        },
        contained: {
          boxShadow: "none",
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: "none",
        },
        rounded: {
          borderRadius: 14,
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          backgroundImage: "none",
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          backgroundImage: "none",
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        colorSecondary: {
          backgroundImage: "none",
          backdropFilter: "blur(8px)",
        },
      },
    },
    MuiFilledInput: {
      styleOverrides: {
        root: {
          borderRadius: 10,
        },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          borderRadius: 10,
        },
      },
    },
    MuiTextField: {
      defaultProps: {
        variant: "outlined",
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 999,
          fontWeight: 600,
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        head: {
          fontWeight: 700,
        },
      },
    },
  },
};

export const lightTheme = deepmerge(defaultLightTheme, {
  ...sharedOverrides,
  palette: {
    mode: "light",
    primary: {
      main: "#2F6B4F",
      light: "#4E8A69",
      dark: "#234F3A",
      contrastText: "#FFFFFF",
    },
    secondary: {
      main: "#B24A3F",
      light: "#C86A61",
      dark: "#8A372F",
      contrastText: "#FFFFFF",
    },
    success: {
      main: "#3F7A57",
    },
    warning: {
      main: "#C48B3A",
    },
    error: {
      main: "#B24A3F",
    },
    info: {
      main: "#4B6E8C",
    },
    background: {
      default: "#F6F4EF",
      paper: "#FFFDF8",
    },
    divider: "#DED8CC",
    text: {
      primary: "#243127",
      secondary: "#5E675F",
    },
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        colorSecondary: {
          backgroundColor: "rgba(255, 253, 248, 0.9)",
          color: "#243127",
          borderBottom: "1px solid rgba(47, 107, 79, 0.12)",
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          border: "1px solid rgba(47, 107, 79, 0.08)",
        },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        notchedOutline: {
          borderColor: "rgba(47, 107, 79, 0.18)",
        },
      },
    },
  },
} satisfies ThemeOptions);

export const darkTheme = deepmerge(defaultDarkTheme, {
  ...sharedOverrides,
  palette: {
    mode: "dark",
    primary: {
      main: "#7FAF91",
      light: "#9FC3AC",
      dark: "#5F8C71",
      contrastText: "#102017",
    },
    secondary: {
      main: "#D18479",
      light: "#E0A199",
      dark: "#B7665C",
      contrastText: "#201310",
    },
    success: {
      main: "#78A987",
    },
    warning: {
      main: "#D0A35B",
    },
    error: {
      main: "#D18479",
    },
    info: {
      main: "#7DA4C7",
    },
    background: {
      default: "#1C1F1B",
      paper: "#232823",
    },
    divider: "#343B35",
    text: {
      primary: "#E9ECE6",
      secondary: "#B5BDB3",
    },
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        colorSecondary: {
          backgroundColor: "rgba(35, 40, 35, 0.88)",
          color: "#E9ECE6",
          borderBottom: "1px solid rgba(127, 175, 145, 0.12)",
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          border: "1px solid rgba(127, 175, 145, 0.08)",
        },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        notchedOutline: {
          borderColor: "rgba(127, 175, 145, 0.2)",
        },
      },
    },
  },
} satisfies ThemeOptions);
