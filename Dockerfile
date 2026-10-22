FROM golang:1.25-alpine AS builder

WORKDIR /src

# Provide the specific service to build
ARG SERVICE

# Copy go.work files to establish the multi-module workspace
COPY go.work go.work.sum ./

# Copy the go.mod and go.sum for all modules first to leverage Docker cache
COPY common/go.mod common/go.sum ./common/
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY orders/go.mod orders/go.sum ./orders/
COPY payments/go.mod payments/go.sum ./payments/
COPY stock/go.mod stock/go.sum ./stock/
COPY kitchen/go.mod kitchen/go.sum ./kitchen/
COPY cart/go.mod cart/go.sum ./cart/
COPY notification/go.mod notification/go.sum ./notification/

# Download dependencies for all modules
RUN for module in common gateway orders payments stock kitchen cart notification; do \
      (cd ./$module && go mod download); \
    done

# Copy the rest of the source code
COPY . .

# Build the specific microservice
RUN cd ${SERVICE} && CGO_ENABLED=0 GOOS=linux go build -o /bin/app .

# Final Stage
FROM alpine:latest

WORKDIR /app

# Install root CA certificates if any service needs to make outbound HTTPS calls (e.g. Razorpay)
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/app .

CMD ["./app"]
