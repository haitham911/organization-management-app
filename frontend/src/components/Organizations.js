import React, { useState, useEffect } from 'react';
import axios from './api';

const Organizations = () => {
  const [organizations, setOrganizations] = useState([]);

  useEffect(() => {
    axios.get('/organizations')
      .then(response => {
        setOrganizations(response.data);
      })
      .catch(error => {
        console.error('There was an error fetching the organizations!', error);
      });
  }, []);

  return (
    <div>
      <h2>Organizations</h2>
      <ul>
        {organizations.map(org => (
          <li key={org.ID}>{org.Name}</li>
        ))}
      </ul>
    </div>
  );
};

export default Organizations;
