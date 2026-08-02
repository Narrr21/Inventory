import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import Dashboard from "./Dashboard/Dashboard";
import ProjectPage from "./Dashboard/Project";

function App() {
  return (
    <Router>
      <Routes>
        {/* LANDING PAGE */}
        <Route path="/" element={<Navigate to="/dashboard" replace />}></Route>

        {/* DASHBOARD PAGE */}
        <Route path="/dashboard" element={<Dashboard />}></Route>

        {/* PROJECT PAGE */}
        <Route path="/project" element={<ProjectPage />}></Route>
      </Routes>
    </Router>
  );
}

export default App;
