import { useState } from 'react';
import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Wardrobe from './pages/Wardrobe';
import Settings from './pages/Settings';
import AddClothes from './pages/AddClothes';

function App() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <Layout sidebarOpen={sidebarOpen} setSidebarOpen={setSidebarOpen}>
      <Routes>
        <Route path="/" element={<Wardrobe />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/add" element={<AddClothes />} />
      </Routes>
    </Layout>
  );
}

export default App;
