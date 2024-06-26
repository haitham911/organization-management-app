import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import InviteUser from './components/InviteUser';
import MagicLink from './components/MagicLink';
import Organizations from './components/Organizations';
import ProductList from './components/Products';
import Subscriptions from './components/Subscriptions';
import Users from './components/Users';
import SubscribeProduct from './components/SubscribeProduct';
import Navigation from './components/Navigation'; // Import Navigation component
import AddOrganization from './components/AddOrganization';

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
          <Route path="/add-organization" element={<AddOrganization/>} />
          <Route path="/users" element={<Users />} />
          <Route path="/products" element={<ProductList />} />
          <Route path="/subscriptions" element={<Subscriptions />} />
          <Route path="/subscribe" element={<SubscribeProduct />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
