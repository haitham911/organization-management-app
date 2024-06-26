import React, { useEffect, useState } from 'react';
import axios from './api';

const ProductList = () => {
  const [products, setProducts] = useState([]);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await axios.get('/products');
        setProducts(response.data);
      } catch (error) {
        console.error('Error fetching products:', error);
      }
    };

    fetchProducts();
  }, []);

  return (
    <div>
      <h2>Products</h2>
      <ul>
        {products.map(productWithPrices => (
          <li key={productWithPrices.product.id}>
            {productWithPrices.product.name}: {productWithPrices.product.description}
            <ul>
              {productWithPrices.prices.map(price => (
                <li key={price.id}>
                  {price.currency.toUpperCase()} {price.unit_amount / 100} per {price.recurring.interval}
                </li>
              ))}
            </ul>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default ProductList;
