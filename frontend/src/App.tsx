import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Dashboard from './Dashboard/Dashboard';

function App() {
  return (
    <Router>
      <Routes>
        {/* LANDING PAGE */}
        <Route path="/" element={<div>Ini halaman home nantinya</div>}></Route>

        {/* DASHBOARD PAGE */}
        <Route path='/dashboard' element={<Dashboard />}></Route>
      </Routes>
    </Router>
  );
}

export default App;
