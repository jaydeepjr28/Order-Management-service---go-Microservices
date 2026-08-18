# OMS Go Microservices

> A production-inspired Order Management System built with Go, gRPC, RabbitMQ, Consul, OpenTelemetry, Gin, PostgreSQL, Redis, and React.

## Overview

This project demonstrates a robust, distributed backend built with independently deployable Go microservices, coupled with a modern React frontend. The system showcases best practices in distributed systems architecture, combining **synchronous gRPC** communication for real-time, critical-path interactions with **asynchronous RabbitMQ** events for decoupled, eventual consistency. 

Service discovery is dynamically managed by **HashiCorp Consul**, while **OpenTelemetry** and **Jaeger** provide comprehensive end-to-end distributed tracing across HTTP, gRPC, and AMQP boundaries.

---

## 🎯 Key Features

- **REST API Gateway**: Built with Gin, featuring JWT authentication, request routing, and acting as the single entry point for clients.
- **Microservices Architecture**: Domain-driven Go services communicating via gRPC and RabbitMQ.
- **Database-per-Service Pattern**: Isolated PostgreSQL databases for the `gateway`, `orders`, and `payments` services to ensure loose coupling.
- **Caching & Ephemeral State**: Redis-backed shopping carts for low-latency operations.
- **Service Discovery**: HashiCorp Consul integration for dynamic service registration and health checks.
- **Observability**: OpenTelemetry tracing integrated with Jaeger for monitoring request lifecycles across the distributed system.
- **External Integration**: Razorpay payment gateway integration with webhook support for real-time payment confirmation.
- **Frontend App**: A React application built with Vite providing a user interface for the OMS.
- **Containerized Environment**: Complete local development setup using Docker Compose.

---

## 🏗️ High-Level Architecture & Data Flow

```text
Client (React Frontend)
  |
HTTP / REST (JWT)
  |
API Gateway
  |---------------------> Cart (Redis)
  |---------------------> Orders ----gRPC----> Stock (Inventory Check)
  |                                      |
  |                                      +---- RabbitMQ ("order.created") ----> Payments
  |                                                         
  |                                      +---- RabbitMQ ("order.paid") -------> Notification
  |                                                                           -> Kitchen (Prep)
  |                                                                           -> Stock (Deduct)
  |
  +---- PostgreSQL databases (Isolated per service: Gateway DB, Orders DB, Payments DB)
```

### The Order Lifecycle
1. **Creation**: A user checks out via the API Gateway. The Gateway makes a synchronous gRPC call to the **Orders** service.
2. **Validation**: The Orders service makes a synchronous gRPC call to the **Stock** service to ensure inventory is available.
3. **Persistence**: If valid, the order is saved in the Orders PostgreSQL database.
4. **Event Published**: The Orders service publishes an `order.created` event to RabbitMQ.
5. **Payment Link**: The **Payments** service consumes `order.created` and generates a Razorpay payment link.
6. **Payment Confirmation**: Upon successful payment, Razorpay hits a webhook on the Payments service, which verifies the signature and publishes an `order.paid` event.
7. **Fulfillment**: 
    - The **Kitchen** service consumes `order.paid` and begins preparation. It updates the order status via gRPC as prep progresses.
    - The **Stock** service consumes `order.paid` to permanently deduct the reserved inventory.
    - The **Notification** service consumes `order.paid` to simulate sending a receipt to the user.

---

## 📂 Project Structure (Domain Services)

| Service | Responsibility | Port (Dev) | Datastore |
| --- | --- | --- | --- |
| `gateway` | REST API, JWT authentication, rate limiting, and request routing. | `8080` | PostgreSQL (`gateway`) |
| `orders` | Creates orders, validates stock, and maintains order state (`OrderService`). | `2000` | PostgreSQL (`orders`) |
| `payments` | Creates payment links and processes Razorpay webhooks. | `2001` | PostgreSQL (`payments`) |
| `stock` | Provides inventory validation and final inventory updates (`StockService`). | `2002` | In-memory/Configured |
| `kitchen` | Consumes paid orders and marks preparation progress. | `2003` | None (Stateless) |
| `cart` | Provides Redis-backed shopping cart operations (`CartService`). | `2005` | Redis |
| `notification`| Consumes `user.registered` and `order.paid` events. | `2006` | None (Stateless) |
| `frontend` | React-based user interface for interacting with the OMS. | `5173` | N/A |
| `common` | Shared code (protobufs, telemetry, broker configs, error models). | N/A | N/A |

---

## 🚀 Getting Started

### Prerequisites

- **Docker & Docker Compose**: For infrastructure (Postgres, Redis, RabbitMQ, Consul, Jaeger).
- **Go 1.20+**: If running Go services directly on your host machine.
- **Make**: For using `Makefile` commands.
- **Node.js & npm**: For the React frontend.

### 1. Environment Configuration

Create a `.env` file in the root directory.

```env
# Production/Standard Database Passwords
GATEWAY_POSTGRES_PASSWORD=password
ORDERS_POSTGRES_PASSWORD=password
PAYMENTS_POSTGRES_PASSWORD=password

# Development Database Passwords (Used by docker-compose.dev.yml)
DEV_GATEWAY_POSTGRES_PASSWORD=password
DEV_ORDERS_POSTGRES_PASSWORD=password
DEV_PAYMENTS_POSTGRES_PASSWORD=password

# Razorpay Credentials (Required for payments service)
RAZORPAY_KEY_ID=your_key_id
RAZORPAY_KEY_SECRET=your_key_secret
RAZORPAY_WEBHOOK_SECRET=your_webhook_secret
```

### 2. Running the System (Fully Dockerized)

The easiest way to start the entire system (infrastructure + all microservices) is using Docker Compose:

```bash
docker compose -f docker-compose.dev.yml up --build
```

This provisions:
- 3x PostgreSQL instances
- Redis & RabbitMQ
- HashiCorp Consul & Jaeger
- All Go microservices

### 3. Running Services Locally (Hybrid Mode)

For faster iterative development, run the infrastructure in Docker and the Go services on your host:

Start infrastructure dependencies:
```bash
make deps
```

Start all Go services (Note: this opens new cmd windows for each service, configured for Windows):
```bash
make run
```

### 4. Running the Frontend

```bash
cd frontend
npm install
npm run dev
```
The React application will be available at `http://localhost:5173`.

---

## 📡 API Endpoints (Gateway)

The API Gateway is exposed at `http://localhost:8080`.

**Authentication (`UserService`)**
- `POST /api/users/register` - Register a user (publishes `user.registered`). Returns JWT.
- `POST /api/users/login` - Authenticate a user. Returns JWT.

**Orders (`OrderService`)**
- `POST /api/customers/:customerID/orders` - Create an order (Body: `[{ "ID": "item_1", "Quantity": 2 }]`).
- `GET /api/customers/:customerID/orders/:orderID` - Fetch specific order details.
- `GET /api/customers/:customerID/orders/:orderID/status` - Fetch order status.

**Inventory & Cart (`CartService` & `StockService`)**
- `GET /api/inventory` - List available items.
- `GET /api/cart` - Fetch the active cart.
- `POST /api/cart/items` - Add item to cart (Body: `{ "ProductID": "item_1", "Quantity": 1 }`).
- `DELETE /api/cart/items/:productID` - Remove item from cart.
- `DELETE /api/cart` - Clear the cart.

---

## 🔍 Observability & Tracing

The system uses **OpenTelemetry** to trace requests as they travel across HTTP, gRPC, and RabbitMQ.

When the system is running, access the **Jaeger UI** at:
👉 `http://localhost:16686`

This allows you to visualize latency, pinpoint bottlenecks, and understand the exact path of an order from creation to payment to fulfillment.

---

## 🛠️ Technology Stack

| Component | Technology |
| --- | --- |
| **Backend Language** | Go (Golang) |
| **Frontend Language** | React, Vite, JavaScript |
| **API Gateway** | Gin (HTTP Framework) |
| **RPC Framework** | gRPC & Protocol Buffers |
| **Message Broker** | RabbitMQ |
| **Service Registry**| HashiCorp Consul |
| **Databases** | PostgreSQL (Relational) |
| **Cache** | Redis (In-memory) |
| **Tracing** | OpenTelemetry & Jaeger |
| **Deployment** | Docker & Docker Compose |

---

## License

Educational project for learning distributed systems, event-driven architectures, and Go microservices.
