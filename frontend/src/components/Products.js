import React, { useState, useEffect } from 'react';
import axios from './api';

const Products = () => {
  const [products, setProducts] = useState([]);

  useEffect(() => {
    axios.get('/products')
      .then(response => {
        setProducts(response.data);
      })
      .catch(error => {
        console.error('There was an error fetching the products!', error);
      });
  }, []);

  return (
    <div>
      <h2>Products</h2>
      <ul>
        {products.map(product => (
          <li key={product.ID}>{product.Name}</li>
        ))}
      </ul>
    </div>
  );
};

export default Products;
