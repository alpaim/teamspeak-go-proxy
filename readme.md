# TeamSpeak 3 UDP Proxy

> [!CAUTION]
> Please, read this segment carefully

## Disclaimer

This is a personal learning project developed during my Golang studies. It's not production-ready and should be used for educational/experimental purposes only. Use at your own risk.

## What is this project?

A simple UDP proxy for TeamSpeak 3 traffic written in Go. Routes client connections to a TS3 server while providing basic metrics and logging.

## Features

- UDP traffic proxying
- Connection metrics (bytes transferred, active connections)
- Automatic session cleanup
- Docker support
- Multi-architecture builds (amd64/arm64)

## Installation

### Prerequisites
- Go 1.21+ (for development/build from source)
- Docker (optional)

### Method 1: Build from Source
```bash
git clone https://github.com/alpaim/teamspeak-go-proxy
cd teamspeak-server-proxy
go build -o teamspeak-server-proxy .

# Run directly
./teamspeak-server-proxy -proxy :9987 -server ts3.your-server.com:9987
```

### Method 2: Docker CLI
```bash
# Inline Docker run
docker run -d \
  -p 9987:9987/udp \
  -e SERVER_ADDR=ts3.your-server.com:9987 \
  -e PROXY_ADDR=:9987 \
  ghcr.io/alpaim/teamspeak-go-proxy:latest
```

### Method 3: Docker Compose
Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  teamspeak-server-proxy:
    image: ghcr.io/alpaim/teamspeak-go-proxy:latest
    ports:
      - "9987:9987/udp"
    environment:
      - SERVER_ADDR=ts3.your-server.com:9987
    restart: unless-stopped
```

Start the service:
```bash
docker-compose up -d
```

## Configuration

| Environment Variable | Default       | Description                     |
|----------------------|---------------|---------------------------------|
| `SERVER_ADDR`        | **Required**  | TS3 server address (host:port)  |
| `PROXY_ADDR`         | `:9987`       | Proxy listening address         |

## Development

1. Clone repository:
```bash
git clone https://github.com/alpaim/teamspeak-go-proxy
cd teamspeak-server-proxy
```

2. Build:
```bash
# Build binary
go build -o teamspeak-server-proxy .

# Run with debug logging
./teamspeak-server-proxy -proxy :9987 -server localhost:10087 -verbose
```

3. Docker development build:
```bash
docker build -t teamspeak-server-proxy-dev .
```

## Monitoring

The proxy outputs metrics to logs:
- New connection events
- Closed connections (with reason)
- Periodic stats:
  ```
  2025/1/1 10:00:00 Connection stats - Active: 3
  ```

## Project Status

This project was developed as part of my Golang learning journey. It includes:

- Basic UDP proxying
- Connection management
- Simple metrics collection
- Docker packaging

**Not production-ready** - missing features:
- Proper authentication
- Advanced error handling
- Performance optimizations
- TLS support
- Configuration file support
- And a lot of other important features