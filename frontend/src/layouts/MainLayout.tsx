import React, { useState, useMemo } from "react";
import {
  AppBar,
  Toolbar,
  Container,
  Box,
  Button,
  IconButton,
  CssBaseline,
} from "@mui/material";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import type { PaletteMode } from "@mui/material";
import Brightness4Icon from "@mui/icons-material/Brightness4";
import Brightness7Icon from "@mui/icons-material/Brightness7";
import { Link, useLocation } from "react-router-dom"; // Import routing
import { getDesignTokens } from "../ColorPalette";

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const [mode, setMode] = useState<PaletteMode>("light");
  const location = useLocation(); // Ambil lokasi URL saat ini

  const toggleColorMode = () => {
    setMode((prevMode) => (prevMode === "light" ? "dark" : "light"));
  };

  const theme = useMemo(() => createTheme(getDesignTokens(mode)), [mode]);

  // Cek apakah halaman aktif adalah /dashboard
  const isDashboardActive = location.pathname === "/dashboard";
  const isProjectActive = location.pathname === "/project";
  const isMapActive = location.pathname === "/map";
  const isAnalyticActive = location.pathname === "/analytic";

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />

      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          minHeight: "100vh",
          bgcolor: "background.default",
        }}
      >
        {/* Navigation Bar */}
        <AppBar
          position="sticky"
          elevation={0}
          sx={{
            bgcolor: "background.nav",
            color: "text.primary",
            borderBottom: "1px solid",
            borderColor: "divider",
          }}
        >
          <Toolbar sx={{ justifyContent: "space-between" }}>
            {/* Logo / Icon di sebelah kiri -> Tautkan ke '/' */}
            <Box
              component={Link}
              to="/"
              sx={{
                display: "flex",
                alignItems: "center",
                textDecoration: "none",
                mr: 2,
              }}
            >
              <Box
                component="img"
                src="/favicon.svg"
                alt="App Icon"
                sx={{
                  width: 32,
                  height: 32,
                  cursor: "pointer",
                  transition: "transform 0.2s",
                  "&:hover": {
                    transform: "scale(1.05)",
                  },
                }}
              />
            </Box>

            {/* Menu Navigasi di sebelah kanan */}
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Button
                component={Link}
                to="/dashboard"
                sx={{
                  color: isDashboardActive
                    ? "primary.contrastText"
                    : "text.primary",
                  bgcolor: isDashboardActive ? "primary.dark" : "transparent",
                  fontWeight: isDashboardActive ? 600 : 400,
                  "&:hover": isDashboardActive
                    ? {}
                    : {
                        bgcolor: "action.hover",
                      },
                }}
              >
                Dashboard
              </Button>
              <Button
                component={Link}
                to="/project"
                sx={{
                  color: isProjectActive
                    ? "primary.contrastText"
                    : "text.primary",
                  bgcolor: isProjectActive ? "primary.dark" : "transparent",
                  fontWeight: isProjectActive ? 600 : 400,
                  "&:hover": isProjectActive
                    ? {}
                    : {
                        bgcolor: "action.hover",
                      },
                }}
              >
                Project
              </Button>
              <Button
                component={Link}
                to="/map"
                sx={{
                  color: isMapActive ? "primary.contrastText" : "text.primary",
                  bgcolor: isMapActive ? "primary.dark" : "transparent",
                  fontWeight: isMapActive ? 600 : 500,
                  "&:hover": isMapActive
                    ? {}
                    : {
                        bgcolor: "action.hover",
                      },
                }}
              >
                Map
              </Button>
              <Button
                component={Link}
                to="/analytic"
                sx={{
                  color: isAnalyticActive
                    ? "primary.contrastText"
                    : "text.primary",
                  bgcolor: isAnalyticActive ? "primary.dark" : "transparent",
                  fontWeight: isAnalyticActive ? 600 : 500,
                  "&:hover": isAnalyticActive
                    ? {}
                    : {
                        bgcolor: "action.hover",
                      },
                }}
              >
                Analytic
              </Button>

              <IconButton
                onClick={toggleColorMode}
                color="inherit"
                aria-label="toggle color mode"
              >
                {mode === "light" ? <Brightness4Icon /> : <Brightness7Icon />}
              </IconButton>
            </Box>
          </Toolbar>
        </AppBar>

        {/* Main Content */}
        <Container
          component="main"
          maxWidth="lg"
          sx={{
            mt: 4,
            mb: 4,
            flexGrow: 1,
          }}
        >
          {children}
        </Container>
      </Box>
    </ThemeProvider>
  );
};

export default MainLayout;
