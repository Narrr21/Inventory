import React, { useState, useMemo } from "react";
import {
  AppBar,
  Toolbar,
  Typography,
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
import { getDesignTokens } from "../ColorPalette";

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const [mode, setMode] = useState<PaletteMode>("light");

  const toggleColorMode = () => {
    setMode((prevMode) => (prevMode === "light" ? "dark" : "light"));
  };

  const theme = useMemo(() => createTheme(getDesignTokens(mode)), [mode]);

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
          <Toolbar>
            <Typography
              variant="h6"
              component="div"
              sx={{ flexGrow: 1, fontWeight: 600, color: "text.primary" }}
            >
              Inventory App
            </Typography>

            <Button sx={{ mr: 1, color: "text.primary" }}>Dashboard</Button>
            <Button sx={{ mr: 1, color: "text.primary" }}>Items</Button>

            <IconButton
              onClick={toggleColorMode}
              color="inherit"
              aria-label="toggle color mode"
              sx={{ ml: 1 }}
            >
              {mode === "light" ? <Brightness4Icon /> : <Brightness7Icon />}
            </IconButton>
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
