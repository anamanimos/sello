# Sello API Documentation

## Base URL

- **Local**: `http://localhost:8080`
- **Production**: `https://sello.id`

## Authentication

### Register

**POST** `/api/auth/register`
Create a new user account.

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "store_name": "My Great Store"
}
```

### Login

**POST** `/api/auth/login`
Authenticate user and get JWT token.

```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

### Google Login (OAuth)

- **GET** `/api/auth/google/login` - Redirects to Google Sign-In.
- **GET** `/api/auth/google/callback` - Callback URL handling token generation.

## Store Management

### Get Store Details

**GET** `/api/store/:slug` (Public)
**GET** `/api/stores/:store_id` (Private/Owner)

### Update Theme

**PUT** `/api/stores/:store_id/theme`

```json
{
  "theme_id": "starter"
}
```

### Save Theme Configuration

**POST** `/api/stores/:store_id/theme/config`

```json
{
  "config": { "colors": {...}, "hero": {...} }
}
```

## E-Commerce & Checkout

### Create Order (Checkout)

**POST** `/api/stores/:store_id/checkout`

```json
{
  "items": [
    {
      "product_id": "uuid-product-id",
      "quantity": 2,
      "variant": "Red/L"
    }
  ],
  "customer": {
    "name": "Buyer Name",
    "whatsapp": "08123456789",
    "address": "Jakarta..."
  },
  "coupon_code": "SALE10" // Optional
}
```

### Upload Payment Proof

**POST** `/api/orders/:order_id/pay`

- **Content-Type**: `multipart/form-data`
- **Field**: `payment_proof` (Image File)

### Get Order Details

**GET** `/api/orders/:order_id`

## Merchant Dashboard

### List Orders

**GET** `/api/stores/:store_id/orders`

### Update Order Status

**PUT** `/api/stores/:store_id/orders/:order_id/status`

```json
{
  "status": "paid" // pending, waiting_confirmation, paid, shipped, cancelled
}
```

### Manage Coupons

- **GET** `/api/stores/:store_id/coupons`
- **POST** `/api/stores/:store_id/coupons`

```json
{
  "code": "RAMADHAN",
  "type": "fixed", // or percentage
  "value": 10000,
  "min_purchase": 50000,
  "usage_limit": 100,
  "expires_at": "2024-12-31"
}
```

- **DELETE** `/api/coupons/:id`

## Super Admin Console

Base URL: `/console/api`

### Auth

- **POST** `/console/api/auth/login`

### Plan Management

- **GET** `/console/api/plans`
- **POST** `/console/api/plans`
- **PUT** `/console/api/plans/:id`
- **DELETE** `/console/api/plans/:id`

### Order Management (Centralized)

- **GET** `/console/api/orders` - List all orders
- **PUT** `/console/api/orders/:id/status` - Update status

### User Management

- **GET** `/console/api/users`
- **PUT** `/console/api/users/:id/suspend`
