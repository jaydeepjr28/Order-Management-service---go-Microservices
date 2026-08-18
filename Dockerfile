# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.work go.work.sum ./
COPY common/go.mod common/go.sum common/
COPY gateway/go.mod gateway/go.sum gateway/
COPY orders/go.mod orders/go.sum orders/
COPY payments/go.mod payments/go.sum payments/
COPY stock/go.mod stock/go.sum stock/
COPY kitchen/go.mod kitchen/go.sum kitchen/
COPY cart/go.mod cart/go.sum cart/
COPY notification/go.mod notification/go.sum notification/

RUN --mount=type=cache,target=/go/pkg/mod \
    for module in common gateway orders payments stock kitchen cart notification; do \
      (cd "/src/$module" && go mod download); \
    done

COPY common/ common/
COPY gateway/ gateway/
COPY orders/ orders/
COPY payments/ payments/
COPY stock/ stock/
COPY kitchen/ kitchen/
COPY cart/ cart/
COPY notification/ notification/

ARG SERVICE
RUN mkdir -p /out && test -n "$SERVICE" && test -f "/src/$SERVICE/go.mod"
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -C "$SERVICE" -trimpath -ldflags='-s -w' -o /out/main .

FROM alpine:3.22 AS runtime

RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/main /app/main

USER app
ENTRYPOINT ["/app/main"]
