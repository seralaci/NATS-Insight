# NATS Insight

A web-based GUI for [NATS](https://nats.io/) — built for **local development and testing**. Spin it up alongside your NATS server during development to inspect streams, browse messages, manage KV stores, and tail subjects in real time. Think of it as Redis Insight for NATS.

Free and open-source. Runs as a single Docker image.

> **Note:** This project was built as a vibe coding experiment — developed entirely with AI assistance to explore what's possible with this approach.

## Features

- **Connection Management** — Save and switch between multiple NATS servers
- **Server Dashboard** — Real-time server metrics, JetStream status, cluster info
- **Stream Management** — Create, edit, delete, mirror/duplicate streams; browse and inspect messages
- **Consumer Management** — Create, configure, and monitor JetStream consumers
- **Key-Value Store** — Browse, create, update, and delete KV buckets and keys with revision history and live watch
- **Object Store** — File browser interface for JetStream object stores
- **Live Tail** — Real-time message monitoring with subject wildcards
- **Message Publishing** — Publish messages with headers, request/reply support

## Quick Start

### Docker Compose (with NATS server)

```yaml
services:
  nats:
    image: nats:latest
    command: ["--js", "-m", "8222"]
    ports:
      - "4222:4222"
      - "8222:8222"

  nats-insight:
    image: ghcr.io/seralaci/nats-insight:latest
    ports:
      - "8080:8080"
    environment:
      NATS_URL: nats://nats:4222
      NATS_MONITOR_URL: http://nats:8222
```

```bash
docker compose up
```

Then open [http://localhost:8080](http://localhost:8080).

### .NET Aspire

```csharp
var nats = builder.AddNats("nats", port: 4222)
    .WithJetStream()
    .WithManagementPort(8222);

builder.AddContainer("nats-insight", "ghcr.io/seralaci/nats-insight", "latest")
    .WithHttpEndpoint(port: 8080, targetPort: 8080)
    .WithEnvironment("NATS_URL", nats.GetEndpoint("tcp"))
    .WithEnvironment("NATS_MONITOR_URL", $"http://nats:8222")
    .WaitFor(nats);
```

### Docker (standalone)

```bash
docker run -p 8080:8080 \
  -e NATS_URL=nats://host.docker.internal:4222 \
  -e NATS_MONITOR_URL=http://host.docker.internal:8222 \
  ghcr.io/seralaci/nats-insight:latest
```

## Architecture

```
Browser (Vue 3 SPA)
    |          |
    | REST     | WebSocket
    v          v
Go Backend (single binary)
    |
    |--- HTTP Router (chi) --- REST API Handlers
    |                              |
    |--- WebSocket Hub             |
    |    - Live tail               v
    |    - KV watch           NATS Service Layer
    |                          (Streams, KV, ObjStore,
    |--- SQLite                 Consumers, Pub/Sub)
    |    (connection profiles)     |
    |                         _____|_____
    |                        |           |
    v                        v           v
                        NATS Server  NATS Monitoring
                        :4222        :8222 (HTTP)
```

The Go backend acts as a proxy between the browser and NATS (since browsers cannot connect directly to the NATS TCP protocol). The compiled Vue SPA is embedded into the Go binary via `//go:embed`, resulting in a single binary / single Docker image deployment.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go with [nats.go](https://github.com/nats-io/nats.go) client |
| Frontend | Vue 3 + TypeScript + Vite |
| UI | Tailwind CSS 4 |
| State | Pinia |
| Database | SQLite (connection profiles, via modernc.org/sqlite — pure Go, no CGO) |
| Deployment | Single Docker image (Go binary with embedded SPA) |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NATS_URL` | `nats://localhost:4222` | NATS server URL (also seeds a default connection) |
| `NATS_MONITOR_URL` | `http://localhost:8222` | NATS monitoring HTTP endpoint |
| `NATS_USERNAME` | | Username auth for default connection |
| `NATS_PASSWORD` | | Password auth (with NATS_USERNAME) |
| `NATS_TOKEN` | | Token auth (takes precedence over username/password) |
| `LISTEN_ADDR` | `:8080` | HTTP listen address |
| `DATA_DIR` | `~/.nats-insight` | SQLite database directory |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |

## Development

### Prerequisites

- Go 1.26+
- Node.js 20+
- Docker (optional)

### Setup

```bash
git clone https://github.com/seralaci/NATS-Insight.git
cd NATS-Insight

# Development with hot-reload (requires Air: go install github.com/air-verse/air@latest)
make dev

# Or build and run
make build
./bin/nats-insight
```

### Project Structure

```
├── cmd/nats-insight/main.go          # Go entrypoint, embeds frontend
├── internal/
│   ├── api/                          # HTTP handlers and routes (chi)
│   ├── nats/                         # NATS client wrapper
│   ├── ws/                           # WebSocket handlers (tail, KV watch)
│   ├── store/                        # SQLite storage
│   └── config/                       # Environment-based config
├── web/                              # Vue 3 frontend (Vite)
│   └── src/
│       ├── components/               # Shared UI components
│       ├── features/                 # Feature modules (streams, kv, objects, tail...)
│       ├── composables/              # Vue composables
│       ├── stores/                   # Pinia stores
│       ├── lib/                      # API client, WebSocket client
│       └── router/                   # Vue Router config
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
└── Makefile
```

## License

[MIT](LICENSE)
