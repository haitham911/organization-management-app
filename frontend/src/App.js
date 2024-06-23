import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import InviteUser from './components/InviteUser';
import MagicLink from './components/MagicLink';
import Organizations from './components/Organizations';
import Products from './components/Products';
import Subscriptions from './components/Subscriptions';
import Users from './components/Users';
import SubscribeProduct from './components/SubscribeProduct';
import Navigation from './components/Navigation'; // Import Navigation component

function App() {
  return (
    <Router>
      <div className="App">
        <Navigation /> {/* Add the Navigation component */}
        <Routes>
          <Route path="/" element={<Organizations />} />
          <Route path="/invite" element={<InviteUser />} />
          <Route path="/magic-link" element={<MagicLink />} />
          <Route path="/organizations" element={<Organizations />} />
          <Route path="/users" element={<Users />} />
          <Route path="/products" element={<Products />} />
          <Route path="/subscriptions" element={<Subscriptions />} />
          <Route path="/subscribe" element={<SubscribeProduct />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
