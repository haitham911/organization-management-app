import React from 'react';
import { Link } from 'react-router-dom';

const Navigation = () => {
  return (
    <nav>
      <ul>
        <li><Link to="/">Home</Link></li>
        <li><Link to="/organizations">Organizations</Link></li>
        <li><Link to="/add-organization">Add Organizations</Link></li>
        <li><Link to="/users">Users</Link></li>
        <li><Link to="/add-user">Add User</Link> </li>
        <li><Link to="/products">Products</Link></li>
        <li><Link to="/subscriptions">Subscriptions</Link></li>
        <li><Link to="/invite">Invite User</Link></li>
        <li><Link to="/subscribe">Subscribe</Link></li>
      </ul>
    </nav>
  );
};

export default Navigation;
