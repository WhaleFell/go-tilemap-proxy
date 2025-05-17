# ref: https://www.saybackend.com/blog/02-golang-dockerfile

# Stage 1: Install dependencies
FROM golang:1.23-bookworm AS deps


# set build args
ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY

WORKDIR /app

COPY go.mod go.sum ./


RUN go mod download

# Stage 2: Build the application
FROM golang:1.23-bookworm AS builder

WORKDIR /app

COPY --from=deps /go/pkg /go/pkg
COPY . .

# Enable them if you need them
# ENV CGO_ENABLED=0
# ENV GOOS=linux

RUN go build -o ./bin/map-server -trimpath -buildvcs=false -ldflags="-s -w -buildid= -checklinkname=0" -v ./cmd/server

# ============================================================
# Use Debian as the base image for the final stage
# # Final stage: Run the application
# FROM debian:bookworm-slim

# WORKDIR /app

# # Create a non-root user and group
# # RUN groupadd -r appuser && useradd -r -g appuser appuser

# # Copy the built application
# COPY --from=builder /app/bin/map-server ./map-server

# # Install curl for healthcheck
# RUN set -x && apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# # Change ownership of the application binary
# # RUN chown appuser:appuser /app/main

# # Switch to the non-root user
# # USER appuser

# # healthcheck
# # HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
# #   CMD curl -f http://localhost:8080/health || exit 1

# CMD ["./map-server", "-c", "./config/config.yaml"]
# ============================================================

# ============================================================
# Use Alpine as the base image for the final stage
# Final stage: Run the application
FROM alpine:latest

WORKDIR /app

# # Create a non-root user and group
# RUN addgroup -S appuser && adduser -S appuser -G appuser

# Install curl for healthcheck
RUN apk update && apk add --no-cache curl bash libc6-compat

# Copy the built application
COPY --from=builder /app/bin/map-server ./map-server
RUN chmod +x ./map-server

# # Change ownership of the application binary
# RUN chown appuser:appuser /app/main

# # Switch to the non-root user
# USER appuser

CMD ["./map-server", "-c", "./config/config.yaml"]
# ============================================================