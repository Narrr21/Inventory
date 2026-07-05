import type { PaletteMode } from "@mui/material";
import type { ThemeOptions } from "@mui/material/styles";

import '@mui/material/styles';

declare module '@mui/material/styles' {
  interface TypeBackground {
    nav?: string;
  }
}

export const getDesignTokens = (mode: PaletteMode): ThemeOptions => ({
  palette: {
    mode,
    ...(mode === "light"
      ? {
          // ===== LIGHT MODE =====
          primary: {
            light: "#8AC4CD",
            main: "#58A4B0",
            dark: "#3D7A85",
            contrastText: "#FFFFFF",
          },
          secondary: {
            light: "#5C637A",
            main: "#373F51",
            dark: "#24293A",
            contrastText: "#FFFFFF",
          },
          info: {
            light: "#C7D3E0",
            main: "#A9BCD0",
            dark: "#7F97B3",
            contrastText: "#172B4D",
          },
          warning: {
            light: "#E8C2BB",
            main: "#DAA49A",
            dark: "#C17E70",
            contrastText: "#172B4D",
          },
          success: {
            light: "#A9CBB7",
            main: "#7FA88E",
            dark: "#5D8069",
            contrastText: "#FFFFFF",
          },
          error: {
            light: "#DE9A9E",
            main: "#C1666B",
            dark: "#994951",
            contrastText: "#FFFFFF",
          },
          text: {
            primary: "#172B4D",
            secondary: "#5C637A",
            disabled: "#A9BCD0",
          },
          background: {
            default: "#F9FAFB",
            paper: "#e1ecf7",
            nav: "#8dcdd9",
          },
          divider: "#9AA1AE",
          grey: {
            50: "#F9FAFB",
            100: "#F4F5F7",
            200: "#E9EBEF",
            300: "#D8DBE2",
            400: "#B9BFC9",
            500: "#9AA1AE",
            600: "#6B778C",
            700: "#4E5568",
            800: "#373F51",
            900: "#24293A",
          },
        }
      : {
          // ===== DARK MODE =====
          primary: {
            light: "#8AC4CD",
            main: "#6FB6C1",
            dark: "#3D7A85",
            contrastText: "#0A1116",
          },
          secondary: {
            light: "#A9BCD0",
            main: "#8892A6",
            dark: "#5C637A",
            contrastText: "#0A1116",
          },
          info: {
            light: "#C7D3E0",
            main: "#A9BCD0",
            dark: "#7F97B3",
            contrastText: "#0A1116",
          },
          warning: {
            light: "#E8C2BB",
            main: "#DAA49A",
            dark: "#C17E70",
            contrastText: "#0A1116",
          },
          success: {
            light: "#A9CBB7",
            main: "#8FBBA1",
            dark: "#5D8069",
            contrastText: "#0A1116",
          },
          error: {
            light: "#E4A7AB",
            main: "#D08085",
            dark: "#994951",
            contrastText: "#0A1116",
          },
          text: {
            primary: "#F4F5F7",
            secondary: "#B9BFC9",
            disabled: "#6B778C",
          },
          background: {
            default: "#181C25",
            paper: "#232838",
            nav: "#1c274d",
          },
          divider: "#373F51",
          grey: {
            50: "#F9FAFB",
            100: "#E9EBEF",
            200: "#D8DBE2",
            300: "#B9BFC9",
            400: "#9AA1AE",
            500: "#6B778C",
            600: "#4E5568",
            700: "#373F51",
            800: "#24293A",
            900: "#181C25",
          },
        }),
  },
  shape: {
    borderRadius: 10,
  },
  typography: {
    fontFamily: [
      "Inter",
      "-apple-system",
      "BlinkMacSystemFont",
      '"Segoe UI"',
      "Roboto",
      "Arial",
      "sans-serif",
    ].join(","),
  },
});
