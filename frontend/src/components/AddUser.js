import React, { useState, useEffect } from 'react';
import axios from './api'; // Use the custom axios instance
import './AddUser.css';
const AddUser = () => {
    const [name, setName] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
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
      try {
        const response = await axios.post('/organizations/can-add-subscriptions', {
          organization_id: organizationId,
          stripe_subscription_id: stripeSubscriptionId,
        });
        if (!response.data.can_add_more_subscriptions) {
          alert(`Organization ${organizationId} cannot add more subscriptions`);
          return;
        }
  
        const userResponse = await axios.post('/users', {
          name,
          email,
          password,
          organization_id: organizationId,
          stripe_subscription_id: stripeSubscriptionId,
        });
        alert(`User added successfully: ${userResponse.data.name}`);
        setName('');
        setEmail('');
        setPassword('');
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
          <button type="submit">Add User</button>
        </form>
      </div>
    );
  };
  
  export default AddUser;