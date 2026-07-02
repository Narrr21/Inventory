import React from "react";
import {
  AppBar,
  Toolbar,
  Typography,
  Container,
  Box,
  Button,
} from "@mui/material";

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  return (
    <Box sx={{ display: "flex", flexDirection: "column", minHeight: "100vh" }}>
      {/* Navigation Bar */}
      <AppBar position="sticky" elevation={1} sx={{ bgcolor: "inherit"}}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Inventory App
          </Typography>
          <Button color="inherit">Dashboard</Button>
          <Button color="inherit">Items</Button>
        </Toolbar>
      </AppBar>

      {/* Content Container */}
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
  );
};

export default MainLayout;
