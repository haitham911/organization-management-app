import React, { useState } from 'react';
import axios from './api';

const InviteUser = () => {
  const [email, setEmail] = useState('');
  const [organizationID, setOrganizationID] = useState('');
  const [message, setMessage] = useState('');

  const handleInvite = async () => {
    try {
      const response = await axios.post('/invite', { email, organization_id: organizationID });
      setMessage(response.data.message);
    } catch (error) {
      setMessage('Failed to send invitation');
    }
  };

  return (
    <div>
      <h2>Invite User</h2>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Enter email"
      />
      <input
        type="text"
        value={organizationID}
        onChange={(e) => setOrganizationID(e.target.value)}
        placeholder="Enter organization ID"
      />
      <button onClick={handleInvite}>Send Invitation</button>
      {message && <p>{message}</p>}
    </div>
  );
};

export default InviteUser;
