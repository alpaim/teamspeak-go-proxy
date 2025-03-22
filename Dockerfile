# Build stage
FROM golang:1.24-alpine AS builder
ARG TARGETOS TARGETARCH
WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o teamspeak-server-proxy .

# Final stage
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/teamspeak-server-proxy /app/teamspeak-server-proxy

# Add non-root user
RUN addgroup -S ts3proxy && \
    adduser -S ts3proxy -G ts3proxy && \
    chown -R ts3proxy:ts3proxy /app

USER ts3proxy

EXPOSE 9987/udp

ENTRYPOINT ["/app/teamspeak-server-proxy"]