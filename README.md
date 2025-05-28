# Subscription Management Microservice API Documentation

## Table of Contents

1.  [Introduction](#introduction)
2.  [Setup and Deployment](#setup-and-deployment)
    - [Prerequisites](#prerequisites)
    - [Environment Variables](#environment-variables)
    - [Running Locally](#running-locally)
    - [Building for Production](#building-for-production)
    - [Deployment](#deployment)
3.  [Authentication](#authentication)
    - [JWT (JSON Web Token)](#jwt-json-web-token)
    - [Login](#login)
    - [Register](#register)
4.  [API Endpoints](#api-endpoints)
    - [Health Check](#health-check)
    - [Auth Endpoints](#auth-endpoints)
    - [Plan Endpoints](#plan-endpoints)
    - [Subscription Endpoints](#subscription-endpoints)
5.  [Data Models](#data-models)
    - [User](#user)
    - [Plan](#plan)
    - [Subscription](#subscription)
    - [Request/Response Payloads](#requestresponse-payloads)
6.  [Admin Functionality](#admin-functionality)

---

## 1. Introduction

The Subscription Management Microservice provides a RESTful API for managing user subscriptions to various service plans. It handles user authentication, plan creation and management (admin-only), and user subscription lifecycle (creation, retrieval, cancellation, and upsert logic for updates/renewals).

The backend is built with Go (Golang) using the Gin Gonic web framework and MongoDB for data persistence. It also includes a simple frontend for demonstration and testing.

---

## 2. Setup and Deployment

### Prerequisites

- Go (version 1.19 or later recommended)
- MongoDB instance (local or MongoDB Atlas)

### Environment Variables

The application is configured using environment variables. Create a `.env` file in the root of the project or set these variables in your deployment environment:

| Variable        | Description                                           | Example                                                  | Required |
| --------------- | ----------------------------------------------------- | -------------------------------------------------------- | -------- |
| `PORT`          | Port the server will listen on                        | `7000`                                                   | Yes      |
| `MONGO_URI`     | MongoDB connection string                             | `mongodb+srv://<user>:<password>@<cluster>.mongodb.net/` | Yes      |
| `DATABASE_NAME` | Name of the MongoDB database                          | `subscription_db`                                        | Yes      |
| `JWT_SECRET`    | Secret key for signing JWT tokens                     | `your-super-secret-jwt-key-should-be-long-and-random`    | Yes      |
| `JWT_EXPIRY`    | Duration for JWT token validity (e.g., `24h`, `720m`) | `24h`                                                    | Yes      |
| `GIN_MODE`      | Gin framework mode (`debug` or `release`)             | `release` (for production)                               | No       |
| `REDIS_URL`     | Redis connection URL (if message queue is used)       | `redis://localhost:6379`                                 | No       |

_(Code Reference: [core/config/config.go](./core/config/config.go))_

### Running Locally

1.  Clone the repository.
2.  Ensure MongoDB is running and accessible.
3.  Create a `.env` file with your local configuration.
4.  Navigate to the project root and run:
    ```json
    go run main.go
    ```
    The server will start on the port specified in `PORT` (e.g., `http://localhost:7000`).

### Building for Production

To create an optimized, statically-linked binary for production:

```bash
CGO_ENABLED=0 GOOS=linux go build -tags netgo -ldflags '-s -w -extldflags "-static"' -o subservice.out ./main.go
```

This command creates a Linux executable named `subservice.out`.

### Deployment

This service is designed to be deployed on platforms like Render, Heroku, or any Docker-compatible environment.

- **Render**:
  - Use the build command above.
  - Set the start command to `./subservice.out`.
  - Configure environment variables in the Render dashboard. Render automatically sets a `PORT` variable, which the application will use.
- **MongoDB Atlas**: Recommended for cloud-hosted MongoDB.

---

## 3. Authentication

Authentication is handled using JSON Web Tokens (JWT).

### JWT (JSON Web Token)

- **Algorithm**: HS256
- **Claims**:
  - `user_id`: The unique ID of the user (MongoDB ObjectID as hex string).
  - `username`: The username of the user.
  - `exp`: Expiration time.
  - `iat`: Issued at time.
    _(Code Reference: [utils/jwt.go](utils/jwt.go))_

### Register

Users can register themselves on the / page of the app. The username should be unique. Admin username is reserved.

### Login

Once the User is created in the database, you can loging on the same / page. This will create JWT and store in the localStorage of your browser.


### Using the Token

The JWT must be included in the `Authorization` header for all protected endpoints, prefixed with `Bearer `:

Authorization: Bearer <your_jwt_token>

_(Code Reference: [core/middleware/auth.go](core/middleware/auth.go))_

---

## 4. API Endpoints

All API endpoints are prefixed with `/api`.

### Health Check

- **GET `/health`**
  - **Description**: Checks the health of the service, including database connectivity.
  - **Response (Success `200 OK`)**:
    ```json
    {
      "success": true,
      "status": "healthy",
      "timestamp": "2025-05-29T12:00:00Z"
    }
    ```
  - **Response (DB Error `503 Service Unavailable`)**:
    ```json
    {
      "success": false,
      "status": "unhealthy",
      "error": "database connection failed"
    }
    ```

### Auth Endpoints

_(Code Reference: [core/controllers/user_controller.go](core/controllers/user_controller.go))_

- **POST `/api/auth/register`**

  - **Description**: Registers a new user.
  - **Request Body**: `RegisterRequest` (see [Data Models](#data-models))
  - **Response (Success `201 Created`)**: `User` object (excluding password).
    ```json
    {
      "success": true,
      "message": "User registered successfully",
      "data": {
        "id": "60c72b2f9b1e8b5a9f8b4567",
        "username": "newuser",
        "name": "New User"
      }
    }
    ```

- **POST `/api/auth/login`**
  - **Description**: Logs in an existing user and returns a JWT.
  - **Request Body**: `LoginRequest` (see [Data Models](#data-models))
  - **Response (Success `200 OK`)**: `LoginResponse` containing token and user details.
    ```json
    {
      "success": true,
      "message": "Login successful",
      "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": {
          "id": "60c72b2f9b1e8b5a9f8b4567",
          "username": "testuser",
          "name": "Test User"
        }
      }
    }
    ```

### Plan Endpoints

_(Code Reference: [core/controllers/plan_controller.go](core/controllers/plan_controller.go))_

- **GET `/api/plans`** (Public)

  - **Description**: Retrieves a list of all available subscription plans.
  - **Response (Success `200 OK`)**: Array of `Plan` objects.
    ```json
    {
      "success": true,
      "message": "Plans retrieved successfully",
      "data": [
        {
          "id": "60d0b5f0c721e72d0c1b2e3f",
          "name": "Basic",
          "price": 999,
          "features": ["Feature A", "Feature B"],
          "duration": "monthly"
        }
      ]
    }
    ```

- **POST `/api/plans`** (Admin Only, Protected)

  - **Description**: Creates a new subscription plan. Requires admin privileges.
  - **Request Body**: `Plan` object (see [Data Models](#data-models))
  - **Response (Success `201 Created`)**: The created `Plan` object.

- **PUT `/api/plans/:id`** (Admin Only, Protected)

  - **Description**: Updates an existing subscription plan. Requires admin privileges.
  - **Path Parameter**: `id` (string - Plan ObjectID)
  - **Request Body**: `Plan` object with fields to update.
  - **Response (Success `200 OK`)**: The updated `Plan` object.

- **DELETE `/api/plans/:id`** (Admin Only, Protected)
  - **Description**: Deletes a subscription plan. Requires admin privileges.
  - **Path Parameter**: `id` (string - Plan ObjectID)
  - **Response (Success `200 OK`)**:
    ```json
    {
      "success": true,
      "message": "Plan deleted successfully"
    }
    ```

### Subscription Endpoints

_(Code Reference: [core/controllers/subscriptions_controller.go](core/controllers/subscriptions_controller.go))_
All subscription endpoints are protected and require JWT authentication.

- **POST `/api/subscriptions`** (Protected)

  - **Description**: Creates a new subscription for the authenticated user or updates/renews an existing one (upsert logic).
  - **Request Body**: `CreateSubscriptionRequest` (see [Data Models](#data-models))
  - **Response (Success `200 OK` or `201 Created`)**: The created/updated `Subscription` object.
    ```json
    {
      "success": true,
      "message": "Subscription processed successfully",
      "data": {
        "id": "60d0c5f0c721e72d0c1b2e4a",
        "user_id": "60c72b2f9b1e8b5a9f8b4567",
        "plan_id": "60d0b5f0c721e72d0c1b2e3f",
        "status": "ACTIVE",
        "start_date": "2025-05-29T10:00:00Z",
        "expires_at": "2025-06-29T10:00:00Z",
        "created_at": "2025-05-29T10:00:00Z",
        "plan": {
          /* Plan details */
        }
      }
    }
    ```

- **PUT `/api/subscriptions`** (Protected)

  - **Description**: Same as `POST /api/subscriptions`. Updates or creates a subscription.
  - **Request Body**: `CreateSubscriptionRequest`.
  - **Response (Success `200 OK`)**: The updated `Subscription` object.

- **GET `/api/subscriptions/:userId`** (Protected)

  - **Description**: Retrieves the current subscription for the specified user. The authenticated user can typically only fetch their own subscription.
  - **Path Parameter**: `userId` (string - User ObjectID)
  - **Response (Success `200 OK`)**: The user's `Subscription` object. If the subscription has expired, its status will be updated to `EXPIRED` upon fetch.

- **DELETE `/api/subscriptions/:userId`** (Protected)
  - **Description**: Cancels the active subscription for the specified user.
  - **Path Parameter**: `userId` (string - User ObjectID)
  - **Response (Success `200 OK`)**:
    ```json
    {
      "success": true,
      "message": "Subscription cancelled successfully"
    }
    ```

---

## 5. Data Models

_(Code References: [core/models/user.go](core/models/user.go), [core/models/plan.go](core/models/plan.go), [core/models/subscriptions.go](core/models/subscriptions.go))_

### User

```JSON
{
"id": "primitive.ObjectID", // MongoDB ObjectID
"username": "string", // Unique, min 3 characters
"name": "string"
// "password" is not exposed in responses
}
```

### Plan

```JSON
{
"id": "primitive.ObjectID",
"name": "string", // Required
"price": "float64", // Required, min 0
"features": ["string"], // Array of strings, required
"duration": "string" // Required, "monthly" or "yearly"
}
```

### Subscription

```JSON
{
"id": "primitive.ObjectID",
"user*id": "string", // User's ObjectID
"plan_id": "primitive.ObjectID", // Plan's ObjectID
"plan": { /* Plan object, populated on fetch \_/ },
"status": "string", // "ACTIVE", "INACTIVE", "CANCELLED", "EXPIRED"
"start_date": "time.Time", // ISO 8601 format
"expires_at": "time.Time", // ISO 8601 format
"created_at": "time.Time" // ISO 8601 format
}
```

**Note**: `updated_at` field was removed from the `Subscription` model as per prior requests.

### Request/Response Payloads

- **RegisterRequest**:
  ```json
  {
    "username": "string", // Required, min 3
    "name": "string", // Required
    "password": "string" // Required, min 6
  }
  ```
- **LoginRequest**:
  ```json
  {
    "username": "string", // Required
    "password": "string" // Required
  }
  ```
- **LoginResponse**:
  ```json
  {
    "token": "string", // JWT
    "user": {
      /* User object */
    }
  }
  ```
- **CreateSubscriptionRequest**:
  ```json
  {
    "user_id": "string", // User's ObjectID
    "plan_id": "primitive.ObjectID" // Plan's ObjectID
  }
  ```
- **API Standard Response Wrapper**:
  _(Code Reference: [utils/response.go](utils/response.go))_
  All API responses are wrapped in this structure:
  ```json
  {
    "success": "boolean",
    "message": "string", // Optional, for success or general error messages
    "data": "object|array", // Optional, response data
    "error": "string" // Optional, detailed error message
  }
  ```

---

## 6. Admin Functionality

A user with the username `admin` has special privileges for plan management.

- The `admin` user can Create, Update, and Delete plans.
- These actions are protected by an `AdminMiddleware` which checks if the authenticated user's username is `admin`.
  _(Code Reference: [core/middleware/admin.go](core/middleware/admin.go))_
- A dedicated frontend admin dashboard is available at `/admin` for the admin user.
