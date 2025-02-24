import { common } from "@mui/material/colors";
import { createTheme } from "@mui/material/styles";

const ORCHEST_BLUE = "#1700A2";

declare module "@mui/material/styles" {
  interface Theme {
    borderColor?: string;
  }
  // allow configuration using `createTheme`
  interface ThemeOptions {
    borderColor?: string;
  }
}

const theme = createTheme({
  palette: {
    primary: {
      main: ORCHEST_BLUE,
      light: "#d1ccec",
      dark: "#4533b5",
      contrastText: "#fff",
      50: "#edeaf7",
      100: "#d1ccec",
      200: "#a299da",
      300: "#7466c7",
      400: "#4533b5",
      500: "#1700a2",
      600: "#14008a",
      700: "#14008a",
      800: "#0c0051",
      900: "#0d003b",
    },
    secondary: {
      main: "#28a1ce",
      light: "#c0eeff",
      dark: "#003346",
      contrastText: "#fff",
      50: "#ecfaff",
      100: "#c0eeff",
      200: "#95e2ff",
      300: "#6ad6ff",
      400: "#3bbff0",
      500: "#28a1ce",
      600: "#1884ac",
      700: "#0c688a",
      800: "#044d68",
      900: "#003346",
    },
    success: {
      main: "#4db04f",
      light: "#dbefdc",
      dark: "#173518",
      contrastText: "#fff",
      50: "#edf7ed",
      100: "#dbefdc",
      200: "#b8dfb9",
      300: "#94d095",
      400: "#70c072",
      500: "#4db04f",
      600: "#419643",
      700: "#367b37",
      800: "#275828",
      900: "#173518",
    },
    info: {
      main: "#3bbff0",
      light: "#ebf8fd",
      dark: "#123948",
      contrastText: "#fff",
      50: "#ebf8fd",
      100: "#d8f2fc",
      200: "#b1e5f9",
      300: "#89d9f6",
      400: "#3bccf3",
      500: "#3bbff0",
      600: "#32a2cc",
      700: "#2986a8",
      800: "#1e6078",
      900: "#123948",
    },
    error: {
      main: "#ef4444",
      light: "#fef2f2",
      dark: "#7f1d1d",
      contrastText: "#fff",
      50: "#fef2f2",
      100: "#fee2e2",
      200: "#fecaca",
      300: "#fca5a5",
      400: "#f87171",
      500: "#ef4444",
      600: "#dc2626",
      700: "#b91c1c",
      800: "#991b1b",
      900: "#7f1d1d",
    },
    warning: {
      main: "#ff9800",
      light: "#fff5e5",
      dark: "#4d2e00",
      contrastText: "#fff",
      50: "#fff5e5",
      100: "#ffebcc",
      200: "#ffd699",
      300: "#ffc266",
      400: "#ffad33",
      500: "#ff9800",
      600: "#d98100",
      700: "#b36a00",
      800: "#804c00",
      900: "#4d2e00",
    },
    background: { default: common.white },
  },
  borderColor: "rgba(0, 0, 0, 0.12)",
  typography: {
    fontFamily: '"Quicksand", "Open Sans", sans-serif',
  },
  shape: {
    borderRadius: 8,
  },
});

export default theme;
