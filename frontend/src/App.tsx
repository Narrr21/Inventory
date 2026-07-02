import { useState } from "react";
import MainLayout from "./layouts/MainLayout";
import { Typography, Paper, Button } from "@mui/material";

function App() {
  const [count, setCount] = useState(0);

  return (
    <MainLayout>
      <Paper elevation={2} sx={{ p: 4, textAlign: "center" }}>
        <Typography variant="body1" color="textPrimary" gutterBottom>
          This content scales down beautifully on mobile and is centered with
          clean side margins on desktop.
        </Typography>
        <Button
          variant="contained"
          onClick={() => setCount((count) => count + 1)}
        >
          Count is {count}
        </Button>
      </Paper>
    </MainLayout>
  );
}

export default App;
