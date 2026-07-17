# OMS Go Microservices

> A production-inspired Order Management System built with **Go**,
> **gRPC**, **RabbitMQ**, **Consul**, **OpenTelemetry**, and **Gin**.

## Overview

This project demonstrates how a distributed backend system can be built
using a microservices architecture. It combines synchronous
communication through **gRPC** with asynchronous event-driven
communication using **RabbitMQ**. Service discovery is handled by
**Consul**, while **OpenTelemetry** and **Jaeger** provide end-to-end
distributed tracing.

## Key Features

-   API Gateway exposing REST endpoints
-   Independent Go microservices
-   gRPC for internal synchronous communication
-   RabbitMQ for asynchronous workflows
-   Consul service discovery and health registration
-   OpenTelemetry tracing across HTTP, gRPC and AMQP
-   Retry handling for message consumers
-   Razorpay payment webhook integration
-   Docker-based local development

------------------------------------------------------------------------

# High-Level Architecture

``` text
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

------------------------------------------------------------------------

# Communication

## Synchronous (gRPC)

-   Gateway → Orders
-   Orders → Stock
-   Kitchen → Orders

## Asynchronous (RabbitMQ)

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

------------------------------------------------------------------------

# Typical Request Flow

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

------------------------------------------------------------------------

# License

Educational project for learning distributed systems and Go
microservices.
