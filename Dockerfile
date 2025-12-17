# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build \
    -tags "containers_image_ostree_stub exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp" \
    -o bin/replicated \
    cli/main.go

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates curl git && \
    update-ca-certificates

ENV IN_CONTAINER=1

WORKDIR /out

# Copy binary from builder stage
COPY --from=builder /build/bin/replicated /replicated

LABEL com.replicated.vendor_cli=true

ENTRYPOINT ["/replicated"]
