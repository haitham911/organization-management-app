import React, { useState, useEffect } from 'react';
import axios from './api';

const SubscribeProduct = () => {
  const [products, setProducts] = useState([]);
  const [organizations, setOrganizations] = useState([]);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [selectedOrganization, setSelectedOrganization] = useState(null);
  const [quantity, setQuantity] = useState(1);

  useEffect(() => {
    axios.get('/products')
      .then(response => {
        setProducts(response.data);
      })
      .catch(error => {
        console.error('There was an error fetching the products!', error);
      });

    axios.get('/organizations')
      .then(response => {
        setOrganizations(response.data);
      })
      .catch(error => {
        console.error('There was an error fetching the organizations!', error);
      });
  }, []);

  const handleSubscribe = async () => {
    try {
      const response = await axios.post('/subscriptions', {
        organization_id: selectedOrganization.ID,
        product_id: selectedProduct.ID,
        quantity: quantity
      });
      console.log(response.data);
    } catch (error) {
      console.error('There was an error creating the subscription!', error);
    }
  };

  return (
    <div>
      <h2>Subscribe to a Product</h2>
      <select onChange={(e) => setSelectedOrganization(organizations.find(o => o.ID === parseInt(e.target.value)))}>
        <option value="">Select an Organization</option>
        {organizations.map(organization => (
          <option key={organization.ID} value={organization.ID}>{organization.Name}</option>
        ))}
      </select>
      <select onChange={(e) => setSelectedProduct(products.find(p => p.ID === parseInt(e.target.value)))}>
        <option value="">Select a Product</option>
        {products.map(product => (
          <option key={product.ID} value={product.ID}>{product.Name}</option>
        ))}
      </select>
      <input
        type="number"
        value={quantity}
        onChange={(e) => setQuantity(parseInt(e.target.value))}
        min="1"
        placeholder="Quantity"
      />
      <button onClick={handleSubscribe}>Subscribe</button>
    </div>
  );
};

export default SubscribeProduct;
