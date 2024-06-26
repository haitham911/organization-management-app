# Organization Management App

This project is a web application that manages organizations as vendors providing products. The application includes features such as subscriptions, user management, and integration with Stripe for payment processing.

## Features

- **Add Products**: Administrators can add new products to Stripe and the application's database.
- **Subscribe to Products**: Organizations can subscribe to products and manage their subscriptions.
- **User Management**: Check if an organization can add more users based on their current subscriptions.
- **Subscription Validation**: Check if users belong to an organization with a subscription to a product.
- **Stripe Integration**: Seamless integration with Stripe for handling payments and subscriptions.

## Technologies Used

- **Frontend**: React, Stripe Elements
- **Backend**: Golang, Gin, GORM, PostgreSQL, Stripe API
- **Database**: PostgreSQL

## Prerequisites

- Node.js
- Go
- PostgreSQL
- Stripe Account

## Setup

### Backend

1. **Clone the repository**:

    ```sh
    git clone https://github.com/your-username/organization-management-app.git
    cd organization-management-app
    ```

2. **Install dependencies**:

    ```sh
    go mod download
    ```

3. **Setup PostgreSQL database**:

    Create a database and update the database configuration in `config/config.go`.

4. **Run the backend**:

    ```sh
    go run main.go
    ```

### Frontend

1. **Navigate to the frontend directory**:

    ```sh
    cd src/frontend
    ```

2. **Install dependencies**:

    ```sh
    npm install
    ```

3. **Start the frontend**:

    ```sh
    npm start
    ```

## API Endpoints

### Products

- **Add Product**: `POST /api/products`
    - Request Body: 
      ```json
      {
        "name": "Product Name",
        "description": "Product Description",
        "price": 1000,
        "currency": "usd",
        "interval": "month"
      }
      ```
    - Response:
      ```json
      {
        "id": 1,
        "name": "Product Name",
        "description": "Product Description",
        "price": 1000,
        "currency": "usd",
        "interval": "month",
        "stripe_product_id": "prod_123"
      }
      ```

### Subscriptions

- **Create Subscription**: `POST /api/subscriptions`
    - Request Body:
      ```json
      {
        "organization_id": 1,
        "product_id": 1,
        "price_id": "price_123",
        "quantity": 1,
        "payment_method_id": "pm_card_visa"
      }
      ```
    - Response:
      ```json
      {
        "subscriptionId": "sub_123"
      }
      ```

### Organization

- **Check if Organization Can Add More Users**: `POST /api/organizations/can-add-users`
    - Request Body:
      ```json
      {
        "organization_id": 1
      }
      ```
    - Response:
      ```json
      {
        "can_add_more_users": true
      }
      ```

- **Check if Users Have Subscription to Product**: `POST /api/organizations/users-have-subscription`
    - Request Body:
      ```json
      {
        "organization_id": 1,
        "product_id": 1
      }
      ```
    - Response:
      ```json
      {
        "has_subscription": true
      }
      ```

## Directory Structure

