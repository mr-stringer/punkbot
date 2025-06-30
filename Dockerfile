# syntax=docker/dockerfile:1

FROM golang:1.24.0-bookworm AS builder

# Set destination for COPY
WORKDIR /app/punkbot

# Ensure static binary is built
ENV CGO_ENABLED=0

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY . ./

# Download Go modules
RUN go mod download

# Build
RUN make

# Get certs from alpine
FROM alpine:latest AS certs
RUN apk --update add ca-certificates

# Final State
FROM scratch
WORKDIR /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/punkbot/punkbot /app/punkbot

ENTRYPOINT ["/app/punkbot"]