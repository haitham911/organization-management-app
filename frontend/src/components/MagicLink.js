import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import axios from './api';

const useQuery = () => {
  return new URLSearchParams(useLocation().search);
};

const MagicLink = () => {
  const query = useQuery();
  const token = query.get('token');
  const [message, setMessage] = useState('');

  useEffect(() => {
    const verifyToken = async () => {
      try {
        const response = await axios.get(`/verify-magic-link?token=${token}`);
        localStorage.setItem('token', response.data.token);
        setMessage('Successfully logged in');
      } catch (error) {
        setMessage('Failed to verify token');
      }
    };

    if (token) {
      verifyToken();
    }
  }, [token]);

  return (
    <div>
      <h2>Magic Link Verification</h2>
      {message && <p>{message}</p>}
    </div>
  );
};

export default MagicLink;
