import React, { useState, useEffect } from 'react';
import axios from './api';

const Subscriptions = () => {
  const [subscriptions, setSubscriptions] = useState([]);

  useEffect(() => {
    axios.get('/subscriptions')
      .then(response => {
        setSubscriptions(response.data);
      })
      .catch(error => {
        console.error('There was an error fetching the subscriptions!', error);
      });
  }, []);

  return (
    <div>
      <h2>Subscriptions</h2>
      <ul>
        {subscriptions.map(subscription => (
          <li key={subscription.ID}>
            {subscription.ProductID} - {subscription.Quantity} users
          </li>
        ))}
      </ul>
    </div>
  );
};

export default Subscriptions;
