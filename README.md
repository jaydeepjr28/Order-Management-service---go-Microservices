# OMS Go Microservices

> A production-inspired Order Management System built with **Go**,
<<<<<<< HEAD
> **gRPC**, **RabbitMQ**, **Consul**, **OpenTelemetry**, and **Gin**.
=======
> **gRPC**, **RabbitMQ**, **Consul**, **OpenTelemetry**, **Gin**, and **PostgreSQL**.
>>>>>>> c6b2e5a (Updated project)

## Overview

This project demonstrates how a distributed backend system can be built
using a microservices architecture. It combines synchronous
communication through **gRPC** with asynchronous event-driven
communication using **RabbitMQ**. Service discovery is handled by
**Consul**, while **OpenTelemetry** and **Jaeger** provide end-to-end
<<<<<<< HEAD
distributed tracing.

## Key Features

-   API Gateway exposing REST endpoints
-   Independent Go microservices
-   gRPC for internal synchronous communication
-   RabbitMQ for asynchronous workflows
=======
distributed tracing. Data is persisted securely in **PostgreSQL** and **Redis**.

## Key Features

-   API Gateway exposing REST endpoints secured via **JWT Authentication**
-   Independent Go microservices
-   gRPC for internal synchronous communication
-   RabbitMQ for asynchronous fanout workflows
>>>>>>> c6b2e5a (Updated project)
-   Consul service discovery and health registration
-   OpenTelemetry tracing across HTTP, gRPC and AMQP
-   Retry handling for message consumers
-   Razorpay payment webhook integration
<<<<<<< HEAD
=======
-   Data persistence mapped across isolated **PostgreSQL** instances
>>>>>>> c6b2e5a (Updated project)
-   Docker-based local development

------------------------------------------------------------------------

# High-Level Architecture

``` text
<<<<<<< HEAD
                    Client
                       │
                 HTTP / REST
                       │
                 ┌─────────────┐
                 │ API Gateway │
                 └─────────────┘
                       │
                     gRPC
                       │
                 ┌─────────────┐
                 │   Orders    │
                 └─────────────┘
                 │           │
          gRPC   │           │ Publish
                 ▼           ▼
          ┌──────────┐   RabbitMQ
          │  Stock   │       │
          └──────────┘       │
                             ▼
                     OrderCreated
                             │
                             ▼
                     Payments Service
                             │
                  Razorpay Webhook
                             │
                             ▼
                         OrderPaid
                        Exchange
                       /        \
                      ▼          ▼
                 Orders      Kitchen
                                 │
                               gRPC
                                 ▼
                             Orders
```

# Services

## API Gateway

-   Public REST API
-   Converts HTTP requests to gRPC
-   Starts distributed traces

## Orders

-   Creates and stores orders
-   Calls Stock through gRPC
-   Publishes `OrderCreated`
-   Consumes payment updates
-   Maintains order state

## Stock

-   Inventory validation
-   Synchronous gRPC API

## Payments

-   Consumes `OrderCreated`
-   Creates payment links
-   Processes Razorpay webhook
-   Publishes `OrderPaid`

## Kitchen

-   Consumes `OrderPaid`
-   Simulates preparation
-   Calls Orders via gRPC to mark status
=======
                     Client
                        │
                 HTTP / REST (JWT)
                        │
                  ┌─────────────┐
                  │ API Gateway │
                  └─────────────┘
                        │
                      gRPC
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
  ┌──────────┐    ┌──────────┐    ┌──────────┐
  │   Cart   │    │  Orders  │    │  Stock   │
  └──────────┘    └──────────┘    └──────────┘
  (Redis DB)           │               │
                       │               │
                 gRPC  │               │ Publish
                       ▼               ▼
                ┌──────────┐      RabbitMQ
                │  Stock   │           │
                └──────────┘           │
                                       ▼
                     UserRegisteredEvent / OrderCreatedEvent
                                       │
                     ┌─────────────────┴─────────────────┐
                     ▼                                   ▼
              Notification                        Payments Service
                Service                                  │
                                                  Razorpay Webhook
                                                         │
                                                         ▼
                                                     OrderPaid
                                                    Exchange
                                                  /      |     \
                                                 ▼       ▼      ▼
                                            Orders    Kitchen  Stock
                                                         │
                                                       gRPC
                                                         ▼
                                                     Orders
```

> Note: Kitchen updates the Orders service through **gRPC** after
> processing; it does **not** publish an `OrderPrepared` event.

------------------------------------------------------------------------
# API Endpoints

The API Gateway exposes the following REST endpoints:

-   `POST /api/users/register` - Registers a user, publishes event, returns JWT.
-   `POST /api/users/login` - Authenticates a user, returns JWT.
-   `POST /api/customers/:customerID/orders` - Creates a new order.
-   `GET /api/customers/:customerID/orders/:orderID` - Fetches the full details of an order.
-   `GET /api/customers/:customerID/orders/:orderID/status` - Fetches just the status of an order.
-   `GET /api/inventory` - Lists all available inventory items from the stock service.
-   `GET /api/cart` - Fetches the active cart for a user.
-   `POST /api/cart/items` - Adds an item to the cart.
-   `DELETE /api/cart/items/:productID` - Removes an item from the cart.
-   `DELETE /api/cart` - Clears the cart.

------------------------------------------------------------------------

# Services

## API Gateway
-   Public REST API.
-   Rate limiting and **JWT Authentication**.
-   Converts HTTP requests to gRPC.
-   Starts distributed traces.
-   Stores user credentials in its dedicated `gateway-postgres` DB.
-   Publishes `UserRegisteredEvent`.

## Orders
-   Creates and stores orders in `orders-postgres` using a hybrid relational/JSONB approach.
-   Calls Stock through gRPC for inventory validation.
-   Publishes `OrderCreated`.
-   Consumes payment updates (`OrderPaid`).
-   Maintains order state.

## Stock
-   Inventory validation.
-   Synchronous gRPC API.
-   Consumes `OrderPaid` to manage final inventory counts.

## Payments
-   Consumes `OrderCreated`.
-   Creates payment links via Razorpay (falls back to an in-memory mock if keys are missing for local dev).
-   Processes Razorpay webhook.
-   Publishes `OrderPaid`.
-   Uses `payments-postgres` for persistent webhook deduplication.

## Kitchen
-   Consumes `OrderPaid`.
-   Simulates preparation.
-   Calls Orders via gRPC to mark status.

## Cart
-   Manages user shopping carts.
-   Exposes gRPC API.
-   Utilizes **Redis** for fast, in-memory storage.

## Notification
-   Background consumer service.
-   Consumes `UserRegisteredEvent` and `OrderCreatedEvent` to send simulated emails/SMS.
>>>>>>> c6b2e5a (Updated project)

------------------------------------------------------------------------

# Communication

## Synchronous (gRPC)

<<<<<<< HEAD
-   Gateway → Orders
-   Orders → Stock
=======
-   Gateway → Orders, Stock, Cart
-   Orders → Stock
-   Payments → Orders
>>>>>>> c6b2e5a (Updated project)
-   Kitchen → Orders

## Asynchronous (RabbitMQ)

<<<<<<< HEAD
-   Orders → OrderCreated
-   Payments → OrderPaid
-   Orders consumes OrderPaid
-   Kitchen consumes OrderPaid

------------------------------------------------------------------------

# Event Flow

  Event          Publisher   Consumers
  -------------- ----------- -----------------
  OrderCreated   Orders      Payments
  OrderPaid      Payments    Orders, Kitchen
=======
-   Gateway → UserRegisteredEvent
-   Orders → OrderCreatedEvent
-   Payments → OrderPaidEvent
-   Notification consumes UserRegisteredEvent & OrderCreatedEvent
-   Payments consumes OrderCreatedEvent
-   Orders, Kitchen, Stock consume OrderPaidEvent
>>>>>>> c6b2e5a (Updated project)

------------------------------------------------------------------------

# Typical Request Flow

<<<<<<< HEAD
1.  Client sends REST request to Gateway.
2.  Gateway forwards request to Orders using gRPC.
3.  Orders validates inventory through Stock.
4.  Orders stores order and publishes `OrderCreated`.
5.  Payments consumes the event and creates a payment.
6.  Razorpay webhook confirms payment.
7.  Payments publishes `OrderPaid`.
8.  Orders updates payment status.
9.  Kitchen consumes `OrderPaid`.
10. Kitchen prepares the order.
11. Kitchen updates Orders using gRPC.
=======
1.  Client registers/logs in via Gateway, receiving a JWT.
2.  Client manages cart items via Gateway → Cart.
3.  Client checks out. Gateway forwards request to Orders using gRPC.
4.  Orders validates inventory through Stock.
5.  Orders stores order in Postgres and publishes `OrderCreated`.
6.  Payments consumes the event and generates a payment link.
7.  Notification consumes the event and emails the user the receipt/link.
8.  Razorpay webhook hits Payments, confirming payment.
9.  Payments publishes `OrderPaid`.
10. Orders updates payment status in Postgres.
11. Kitchen consumes `OrderPaid` and prepares the order.
12. Kitchen updates Orders using gRPC.
13. Stock consumes `OrderPaid` and commits inventory deductions.
>>>>>>> c6b2e5a (Updated project)

------------------------------------------------------------------------

# Service Discovery

All services register with **Consul** and can discover each other
dynamically instead of relying on hard-coded addresses.

------------------------------------------------------------------------

# Observability

OpenTelemetry propagates trace context across:

-   HTTP
-   gRPC
-   RabbitMQ message headers

Jaeger visualizes the complete lifecycle of a request.

------------------------------------------------------------------------

<<<<<<< HEAD
# Reliability

-   Retry handling for failed consumers
-   Health registration with Consul
-   Decoupled asynchronous processing
-   Loosely coupled services via events

------------------------------------------------------------------------

# Folder Structure

``` text
gateway/
orders/
stock/
payments/
kitchen/
common/
```

------------------------------------------------------------------------

=======
>>>>>>> c6b2e5a (Updated project)
# Technology Stack

  Technology         Purpose
  ------------------ -----------------------
  Go                 Microservices
  Gin                REST Gateway
  gRPC               Service communication
  Protocol Buffers   Contracts
  RabbitMQ           Messaging
  Consul             Service discovery
  OpenTelemetry      Tracing
  Jaeger             Trace visualization
  Docker             Local infrastructure
<<<<<<< HEAD

------------------------------------------------------------------------

# License

Educational project for learning distributed systems and Go
microservices.
=======
  PostgreSQL         Persistent data storage
  Redis              In-memory data store

------------------------------------------------------------------------
>>>>>>> c6b2e5a (Updated project)
