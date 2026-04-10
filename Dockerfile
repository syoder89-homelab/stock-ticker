# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app

# Use distroless image for SSL certs
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app /app
USER nonroot
ENTRYPOINT ["/app"]
