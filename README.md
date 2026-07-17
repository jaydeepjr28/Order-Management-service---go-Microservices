# OMS Go Microservices

A production-inspired **Order Management System (OMS)** built using **Go Microservices**, **gRPC**, **RabbitMQ**, and **OpenTelemetry**.

The project demonstrates how modern backend systems are designed using synchronous and asynchronous communication patterns, distributed tracing, and loosely coupled services.

---

# Features

* Built using **5 independent Go microservices**
* REST API exposed through an **API Gateway**
* High-performance **gRPC** communication between internal services
* **RabbitMQ** for asynchronous event-driven workflows
* Distributed tracing using **OpenTelemetry + Jaeger**
* Protocol Buffers for efficient service contracts
* Containerized infrastructure using Docker
* Clean separation of responsibilities across services

---

# Architecture

```text
                           Client
                              │
                     HTTP REST Requests
                              │
                     ┌────────────────┐
                     │  API Gateway   │
                     └────────────────┘
                         │        │
                     gRPC│        │gRPC
                         │        │
               ┌─────────┘        └─────────┐
               │                            │
      ┌────────────────┐          ┌────────────────┐
      │ Orders Service │◄────────►│ Stock Service  │
      └────────────────┘   gRPC   └────────────────┘
               │
               │ Publish
               ▼
         ┌─────────────────┐
         │    RabbitMQ     │
         └─────────────────┘
             │         │
             │         │
             ▼         ▼
    ┌────────────────┐ ┌────────────────┐
    │ Payments       │ │ Kitchen        │
    │ Service        │ │ Service        │
    └────────────────┘ └────────────────┘
             │                 │
             └────Publish──────┘
                    Events
                       │
                       ▼
               Orders Service
```

---

# Order Processing Flow

### Step 1

A client creates an order using the REST API.

```
POST /api/customers/{customerId}/orders
```

↓

### Step 2

The API Gateway forwards the request to the **Orders Service** using **gRPC**.

↓

### Step 3

Orders Service validates inventory by calling the **Stock Service**.

↓

### Step 4

If stock is available:

* Order is created
* Order status becomes **Pending Payment**
* `OrderCreated` event is published to RabbitMQ

↓

### Step 5

Payments Service consumes the event.

It:

* Generates a payment link
* Simulates payment
* Publishes a `PaymentCompleted` event

↓

### Step 6

Orders Service consumes the payment event and updates the order status.

↓

### Step 7

Kitchen Service consumes the payment event.

It:

* Starts preparing the order
* Simulates cooking
* Publishes `OrderPrepared`

↓

### Step 8

Orders Service receives the final event and marks the order as completed.

---

# Microservices

## API Gateway

Responsibilities

* Exposes REST APIs
* Converts HTTP requests into gRPC calls
* Acts as the single entry point
* Starts distributed traces

---

## Orders Service

Responsibilities

* Creates orders
* Stores order information
* Validates stock
* Updates order status
* Publishes order events
* Consumes payment and kitchen events

---

## Stock Service

Responsibilities

* Maintains inventory
* Validates stock availability
* Returns stock information through gRPC

---

## Payments Service

Responsibilities

* Consumes order events
* Simulates payment processing
* Generates payment links
* Publishes payment completion events

---

## Kitchen Service

Responsibilities

* Waits for successful payment
* Simulates cooking
* Publishes order completion events

---

# Communication Patterns

## Synchronous Communication

Used when an immediate response is required.

Technology:

* gRPC

Examples

Gateway → Orders

Orders → Stock

---

## Asynchronous Communication

Used for background workflows.

Technology:

* RabbitMQ

Examples

Orders → Payments

Payments → Orders

Payments → Kitchen

Kitchen → Orders

---

# Event Flow

| Event            | Publisher | Consumer        |
| ---------------- | --------- | --------------- |
| OrderCreated     | Orders    | Payments        |
| PaymentCompleted | Payments  | Orders, Kitchen |
| OrderPrepared    | Kitchen   | Orders          |

---

# Repository Structure

```text
oms-go-microservices/

├── gateway/
├── orders/
├── stock/
├── payments/
├── kitchen/
│
├── common/
│   ├── api/
│   ├── broker/
│   ├── tracing/
│   └── utils/
│
├── docker-compose.yml
├── go.work
└── README.md
```

---

# Technologies Used

| Technology       | Purpose                                |
| ---------------- | -------------------------------------- |
| Go               | Microservice implementation            |
| gRPC             | High-performance service communication |
| Protocol Buffers | API contract generation                |
| RabbitMQ         | Event-driven messaging                 |
| Gin              | REST API Gateway                       |
| OpenTelemetry    | Distributed tracing                    |
| Jaeger           | Trace visualization                    |
| Docker           | Infrastructure containerization        |

---

# Prerequisites

* Go 1.18+
* Docker
* RabbitMQ
* Protocol Buffers Compiler
* protoc-gen-go
* protoc-gen-go-grpc

# Getting Started

## Clone Repository

```bash
git clone https://github.com/jaydeepjr28/oms-go-microservices.git

cd oms-go-microservices
```

---

## Generate Protobuf Code

```bash
protoc \
--go_out=./common/api \
--go_opt=paths=source_relative \
--go-grpc_out=./common/api \
--go-grpc_opt=paths=source_relative \
./common/api/oms.proto
```

---

## Start RabbitMQ

```bash
docker run -d \
--hostname rabbitmq \
--name rabbitmq \
-p 5672:5672 \
-p 15672:15672 \
rabbitmq:3-management
```

RabbitMQ Dashboard

```
http://localhost:15672
```

---

## Start Jaeger

```bash
docker run -d \
--name jaeger \
-e COLLECTOR_OTLP_ENABLED=true \
-p 16686:16686 \
-p 4317:4317 \
jaegertracing/all-in-one:latest
```

Jaeger UI

```
http://localhost:16686
```

---

## Run Services

Start every service in a separate terminal.

```bash
cd stock && go run .

cd orders && go run .

cd payments && go run .

cd kitchen && go run .

cd gateway && go run .
```

---

# API Endpoints

### Create Order

```
POST /api/customers/{customerId}/orders
```

Example Body

```json
[
  {
    "ID": "item1",
    "Quantity": 2
  },
  {
    "ID": "item2",
    "Quantity": 1
  }
]
```

---

### Get Order

```
GET /api/customers/{customerId}/orders/{orderId}
```

---

# Service Ports

| Service     | Default Port |
| ----------- | ------------ |
| Gateway     | 8080         |
| Orders      | 50051        |
| Stock       | 50052        |
| Payments    | 50053        |
| Kitchen     | 50054        |
| RabbitMQ UI | 15672        |
| Jaeger UI   | 16686        |

---

# Observability

Every incoming request creates a distributed trace that propagates across all services.

```
Gateway
   │
Orders
   │
Stock
   │
RabbitMQ
   │
Payments
   │
Kitchen
   │
Orders
```

The complete lifecycle of an order can be visualized inside **Jaeger**, making it easy to debug latency and understand service interactions.

---

# Design Patterns Demonstrated

* API Gateway Pattern
* Microservices Architecture
* Event-Driven Architecture
* Publisher–Subscriber Pattern
* Request–Reply Pattern (gRPC)
* Shared Protobuf Contracts
* Distributed Tracing
* Loose Service Coupling

---

# Learning Objectives

This project was built to explore production-grade backend concepts, including:

* Designing scalable microservices
* Building event-driven systems
* Implementing synchronous and asynchronous communication
* Using Protocol Buffers and gRPC
* Understanding distributed tracing
* Applying clean service boundaries
* Building observable backend systems

---

# License

This project is intended for learning and educational purposes. Feel free to fork, modify, and experiment with it.
