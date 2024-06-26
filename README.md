# Organization Management App

This project is a web application that manages organizations as vendors providing products. The application includes features such as subscriptions, user management, and integration with Stripe for payment processing.
## Business Logic

1. **User Creation**: When a new user is created, the system first checks if the associated organization has enough subscription capacity. The user is then created and linked to the organization and the relevant Stripe subscription.

2. **Subscription Creation**: When an organization subscribes to a product, a subscription is created in Stripe. The subscription details, including the quantity, are stored in the database.

3. **Subscription Validation**: Before adding a new user to an organization, the system checks if the organization can add more subscriptions based on the quantity of the existing subscriptions.

4. **Payment Handling**: Payment methods are attached to organizations when subscriptions are created. Stripe webhooks are used to handle events such as payment successes and subscription updates.

5. **Product Access Check**: The system validates if users within an organization have access to a product based on their subscriptions.

## Features

- **Subscribe to Products**: Organizations can subscribe to products and manage their subscriptions.
- **User Management**: Users can be added to organizations, and their subscriptions are managed.
- **Subscription Validation**: Checks if organizations can add more subscriptions based on their current subscription limits.
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

- **Check if Organization Can Add More Subscriptions**: `POST /api/organizations/can-add-subscriptions`
    - Request Body:
      ```json
      {
        "organization_id": 1,
        "stripe_subscription_id": "sub_123"
      }
      ```
    - Response:
      ```json
      {
        "can_add_more_subscriptions": true
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

### Users

- **Create User**: `POST /api/users`
    - Request Body:
      ```json
      {
        "name": "User Name",
        "email": "user@example.com",
        "password": "password123",
        "organization_id": 1,
        "stripe_subscription_id": "sub_123"
      }
      ```
    - Response:
      ```json
      {
        "id": 1,
        "name": "User Name",
        "email": "user@example.com"
      }
      ```

## Architecture Diagram

```mermaid
graph TD
    A[Frontend React]
    B[Backend Golang]
    C[PostgreSQL Database]
    D[Stripe API]
    
    A -->|API Requests| B
    B -->|Database Queries| C
    B -->|Payment Processing| D
    C -->|Data Storage| B
    D -->|Payment Events| B
