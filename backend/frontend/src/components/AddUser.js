import React, { useState, useEffect } from 'react';
import axios from '../api/axios'; // Use the custom axios instance
import './AddUser.css';

const AddUser = () => {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('User'); // Default role
  const [organizations, setOrganizations] = useState([]);
  const [organizationId, setOrganizationId] = useState('');
  const [stripeSubscriptionId, setStripeSubscriptionId] = useState('');

  useEffect(() => {
    const fetchOrganizations = async () => {
      try {
        const response = await axios.get('/organizations');
        setOrganizations(response.data);
      } catch (error) {
        console.error('Error fetching organizations:', error);
      }
    };

    fetchOrganizations();
  }, []);

  const handleAddUser = async (event) => {
    event.preventDefault();
    if (role !== 'Admin' && (!organizationId || !stripeSubscriptionId)) {
      alert('Organization and Stripe Subscription ID are required for non-admin users.');
      return;
    }

    try {
      if (role !== 'Admin') {
        const response = await axios.post('/organizations/can-add-subscriptions', {
          organization_id: organizationId,
          stripe_subscription_id: stripeSubscriptionId,
        });
        if (!response.data.can_add_more_subscriptions) {
          alert(`Organization ${organizationId} cannot add more subscriptions`);
          return;
        }
      }

      const userResponse = await axios.post('/users', {
        name,
        email,
        password,
        role,
        organization_id: role !== 'Admin' ? organizationId : undefined,
        stripe_subscription_id: role !== 'Admin' ? stripeSubscriptionId : undefined,
      });
      alert(`User added successfully: ${userResponse.data.name}`);
      setName('');
      setEmail('');
      setPassword('');
      setRole('User');
      setOrganizationId('');
      setStripeSubscriptionId('');
    } catch (error) {
      console.error('Error adding user:', error);
      alert('Failed to add user');
    }
  };

  return (
    <div className="add-user-container">
      <h2>Add New User</h2>
      <form onSubmit={handleAddUser}>
        <label>
          Name:
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </label>
        <label>
          Email:
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <label>
          Password:
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>
        <label>
          Role:
          <select
            value={role}
            onChange={(e) => setRole(e.target.value)}
            required
          >
            <option value="User">User</option>
            <option value="Admin">Admin</option>
          </select>
        </label>
        {role !== 'Admin' && (
          <>
            <label>
              Select Organization:
              <select
                value={organizationId}
                onChange={(e) => setOrganizationId(e.target.value)}
                required
              >
                <option value="" disabled>Select an organization</option>
                {organizations.map(org => (
                  <option key={org.id} value={org.id}>{org.name}</option>
                ))}
              </select>
            </label>
            <label>
              Stripe Subscription ID:
              <input
                type="text"
                value={stripeSubscriptionId}
                onChange={(e) => setStripeSubscriptionId(e.target.value)}
                required
              />
            </label>
          </>
        )}
        <button type="submit">Add User</button>
      </form>
    </div>
  );
};

export default AddUser;
