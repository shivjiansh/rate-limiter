FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rate-limiter ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/rate-limiter /app/rate-limiter
EXPOSE 8080
ENTRYPOINT ["/app/rate-limiter"]
