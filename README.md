# Wake-on-HTTP

A lightweight Go-based HTTP server that triggers wake-on-LAN commands via REST API.

## Features

- **GET /**: Sends WOL packet to default MAC address (from `DEFAULT_MAC` environment variable)
- **GET /\<MAC\>**: Sends WOL packet to specified MAC address (format: `XX:XX:XX:XX:XX:XX`)
- Logs all operations
- Lightweight Alpine-based Docker image

## Prerequisites

- Docker and Docker Compose
- Network access to broadcast address

## Configuration

Edit `docker-compose.yml` to set environment variables:

```yaml
environment:
  - DEFAULT_MAC=AA:BB:CC:DD:EE:FF
  - BROADCAST_ADDR=192.168.178.255
  - PORT=8009
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DEFAULT_MAC` | (none) | MAC address to wake when calling `GET /` (format: `XX:XX:XX:XX:XX:XX`) |
| `BROADCAST_ADDR` | `255.255.255.255` | Broadcast address for WOL packets. You may have to adapt if you need to broadcast over a specific subnet in order to e.g. route via a specific network interface. |
| `PORT` | `8009` | HTTP server listen port |

## Usage

### Start the service
```bash
docker-compose up -d
```

### Wake default device
```bash
curl http://localhost:8009/
```

### Wake specific device
```bash
curl http://localhost:8009/AA:BB:CC:DD:EE:FF
```

### View logs
```bash
docker-compose logs -f wake-on-http
```

### Stop the service
```bash
docker-compose down
```

## API Responses

### Success (200 OK)
```
Wake-on-LAN signal sent to AA:BB:CC:DD:EE:FF
```

### Error Responses

- **400 Bad Request**: Invalid MAC address format
- **500 Internal Server Error**: DEFAULT_MAC not set or WOL transmission failed
- **405 Method Not Allowed**: Only GET requests are supported
- **404 Not Found**: Path not recognized

## MAC Address Format

MAC addresses should be formatted as: `XX:XX:XX:XX:XX:XX` (colons are required)

Valid examples:
- `AA:BB:CC:DD:EE:FF`
- `00:11:22:33:44:55`

## Building Manually

```bash
# Build the Docker image
docker build -t wake-on-http .

# Run the container with default settings
docker run -p 8009:8009 -e DEFAULT_MAC=AA:BB:CC:DD:EE:FF wake-on-http

# Run with custom broadcast address and port
docker run -p 8080:8080 -e DEFAULT_MAC=AA:BB:CC:DD:EE:FF -e BROADCAST_ADDR=192.168.1.255 -e PORT=8080 wake-on-http
```
