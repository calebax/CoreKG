# Stage 1: Builder for Go application
FROM golang:1.24-alpine

WORKDIR /app

# Install build dependencies for Go
RUN apk add --no-cache git gcc libc-dev
