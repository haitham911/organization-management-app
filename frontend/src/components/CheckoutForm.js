import React, { useState, useEffect } from 'react';
import { useStripe, useElements, CardElement } from '@stripe/react-stripe-js';
import axios from './api'; // Use the custom axios instance
import './CheckoutForm.css';

const CheckoutForm = ({
  products,
  selectedPrice,
  setSelectedPrice,
  organizations,
  selectedOrganization,
  setSelectedOrganization,
  quantity,
  setQuantity,
}) => {
  const stripe = useStripe();
  const elements = useElements();
  const [useDefault, setUseDefault] = useState(true);
  const [selectedProduct, setSelectedProduct] = useState('');

  useEffect(() => {
    if (useDefault) {
      setSelectedOrganization(organizations.length > 0 ? organizations[0].ID : '');
      if (products.length > 0 && products[0].prices.length > 0) {
        setSelectedPrice(products[0].prices[0].id);
        setSelectedProduct(products[0].id);
      }
      setQuantity(1);
    }
  }, [organizations, products, useDefault]);

  const handleSubscribe = async (event) => {
    event.preventDefault();

    if (!stripe || !elements) {
      return;
    }

    let paymentMethodId;

    if (useDefault) {
      paymentMethodId = 'pm_card_visa';
    } else {
      const cardElement = elements.getElement(CardElement);
      const { error, paymentMethod } = await stripe.createPaymentMethod({
        type: 'card',
        card: cardElement,
      });

      if (error) {
        console.error(error);
        return;
      }

      paymentMethodId = paymentMethod.id;
    }

    try {
      const response = await axios.post('/subscriptions', {
        organization_id: Number(selectedOrganization),
        product_id: Number(selectedProduct),
        price_id: selectedPrice,
        quantity: Number(quantity),
        payment_method_id: paymentMethodId,
      });
      alert(`Subscribed successfully: ${response.data.subscriptionId}`);
    } catch (error) {
      console.error('Error subscribing:', error);
      alert('Subscription failed');
    }
  };

  return (
    <form onSubmit={handleSubscribe}>
      <label>
        Select Organization:
        <select
          value={selectedOrganization}
          onChange={(e) => setSelectedOrganization(e.target.value)}
        >
          <option value="" disabled>Select an organization</option>
          {organizations.map(org => (
            <option key={org.ID} value={org.ID}>{org.name}</option>
          ))}
        </select>
      </label>
      {products.map(productWithPrices => (
        <div key={productWithPrices.product.id}>
          <h3>{productWithPrices.product.name}</h3>
          <p>{productWithPrices.product.description}</p>
          <ul>
            {productWithPrices.prices.map(price => (
              <li key={price.id}>
                <label>
                  <input
                    type="radio"
                    name="price"
                    value={price.id}
                    onChange={() => {
                      setSelectedPrice(price.id);
                      setSelectedProduct(productWithPrices.product.id);
                    }}
                    checked={selectedPrice === price.id}
                  />
                  {price.currency.toUpperCase()} {price.unit_amount / 100} per {price.recurring.interval}
                </label>
              </li>
            ))}
          </ul>
        </div>
      ))}
      <label>
        Quantity:
        <input
          type="number"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
          min="1"
        />
      </label>
      {!useDefault && <CardElement className="CardElement" />}
      <button type="submit" disabled={!stripe || !selectedPrice || !selectedOrganization}>
        Subscribe
      </button>
      <label>
        <input
          type="checkbox"
          checked={useDefault}
          onChange={(e) => setUseDefault(e.target.checked)}
        />
        Use default test payment method
      </label>
    </form>
  );
};

export default CheckoutForm;
