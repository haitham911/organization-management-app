import React, { useEffect, useState } from 'react';
import { loadStripe } from '@stripe/stripe-js';
import { Elements } from '@stripe/react-stripe-js';
import axios from './api';
import CheckoutForm from './CheckoutForm';
import './SubscribeProduct.css';

const stripePromise = loadStripe('pk_test_51PV0Z9Lq8P7MVUmbJbUWkGNUDn1C4hOiaDIVLRKBvqTWSmFwkacAFyci9T5RNXFX1XKPgoSDf22iD1WiFkjLXz7B00Bz3bwr2q');

const SubscribeProduct = () => {
  const [products, setProducts] = useState([]);
  const [selectedPrice, setSelectedPrice] = useState('');
  const [organizations, setOrganizations] = useState([]);
  const [selectedOrganization, setSelectedOrganization] = useState('');
  const [quantity, setQuantity] = useState(1);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await axios.get('/products');
        setProducts(response.data);
      } catch (error) {
        console.error('Error fetching products:', error);
      }
    };

    const fetchOrganizations = async () => {
      try {
        const response = await axios.get('/organizations');
        setOrganizations(response.data);
      } catch (error) {
        console.error('Error fetching organizations:', error);
      }
    };

    fetchProducts();
    fetchOrganizations();
  }, []);

  return (
    <div className="container">
      <h2>Subscribe to a Product</h2>
      <Elements stripe={stripePromise}>
        <CheckoutForm
          products={products}
          selectedPrice={selectedPrice}
          setSelectedPrice={setSelectedPrice}
          organizations={organizations}
          selectedOrganization={selectedOrganization}
          setSelectedOrganization={setSelectedOrganization}
          quantity={quantity}
          setQuantity={setQuantity}
        />
      </Elements>
    </div>
  );
};

export default SubscribeProduct;
