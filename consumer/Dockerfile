FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY events/ ./events/
COPY consumer/go.mod consumer/go.sum ./consumer/

WORKDIR /src/consumer
RUN go mod download

COPY consumer/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /consumer ./cmd/consumer

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /consumer /app/consumer

CMD ["/app/consumer"]
