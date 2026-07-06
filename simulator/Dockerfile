FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY events/ ./events/
COPY simulator/go.mod simulator/go.sum ./simulator/

WORKDIR /src/simulator
RUN go mod download

COPY simulator/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /simulator ./cmd/simulator

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /simulator /app/simulator

CMD ["/app/simulator"]
