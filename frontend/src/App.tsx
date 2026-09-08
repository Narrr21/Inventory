import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import Dashboard from "./page/Dashboard/Dashboard";
import ProjectPage from "./page/Project/Project";
import AnalyticPage from "./page/Analytic/AnalyticPage";
import MapPage from "./page/Map/MapPage";

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

        {/* MAP PAGE */}
        <Route path="/map" element={<MapPage />}></Route>

        {/* ANALYTIC PAGE */}
        <Route path="/analytic" element={<AnalyticPage />}></Route>
      </Routes>
    </Router>
  );
}

export default App;
