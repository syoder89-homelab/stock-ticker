# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./...
# TODO - Future version should be from a release pattern or using bump2version with a VERSION file in this repo
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /app ./cmd/stock-ticker

# Use distroless image for SSL certs
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app /app
USER nonroot
ENTRYPOINT ["/app"]
