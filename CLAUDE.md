# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

NATS Insight is a web-based GUI for NATS messaging system — focused on local development. It runs as a single Docker image with Go backend + embedded Vue frontend.

## Build & Run Commands

```bash
# Full production build (frontend + backend)
make build

# Development with hot-reload (requires Air: go install github.com/air-verse/air@latest)
make dev

# Docker (build + run with NATS server)
make docker-down && make docker && make docker-up
# App at http://localhost:8080

# Go only
go build ./cmd/nats-insight/
go test ./...

# Frontend only
cd web && npm install && npm run build
cd web && npx vue-tsc -b          # type check only
```

## Architecture

**Single binary deployment**: Go backend embeds the compiled Vue SPA via `//go:embed`. One binary, one Docker image, one port.

```
Browser (Vue 3 SPA) ←→ Go Backend ←→ NATS Server (:4222)
                REST + WebSocket        ↕
                                    NATS Monitoring (:8222)
```

### Backend (Go)

- **Entrypoint**: `cmd/nats-insight/main.go` — embeds frontend, starts HTTP server, seeds default connection
- **Router**: `internal/api/router.go` — chi v5, all REST + WebSocket routes registered here
- **API handlers**: `internal/api/` — one file per domain: `connections.go`, `streams.go`, `consumers.go`, `kv.go`, `server.go`
- **NATS manager**: `internal/nats/manager.go` — wraps `*nats.Conn` + `jetstream.JetStream`, manages connection lifecycle. The package is named `nats` which conflicts with the nats.go client — always import as `natsclient "github.com/nats-io/nats.go"`
- **WebSocket handlers**: `internal/ws/` — `tail.go` (live message subscribe), `kvwatch.go` (KV change stream)
- **Storage**: `internal/store/` — SQLite via `modernc.org/sqlite` (pure Go, no CGO) for connection profiles
- **Shared utils**: `internal/api/utils.go` — `ParseByteSize()` for human-readable sizes ("100MB" → bytes)
- **Config**: `internal/config/config.go` — all config via environment variables

### Frontend (Vue 3 + TypeScript)

- **Build**: Vite 8, output to `cmd/nats-insight/web/` (embedded by Go)
- **CSS**: Tailwind CSS 4 with `@custom-variant dark` for dark mode (class-based, not media query)
- **State**: Pinia stores in `web/src/stores/` — `connections.ts`, `server.ts`, `streams.ts`, `kv.ts`, `objects.ts`, `ui.ts`
- **API client**: `web/src/lib/api.ts` — typed fetch wrapper for all REST endpoints
- **WebSocket**: `web/src/lib/ws.ts` (raw client), `web/src/composables/useTail.ts`, `web/src/composables/useKvWatch.ts`
- **Features**: `web/src/features/` — each domain has its own directory (dashboard, connections, streams, kv, objects, tail, publish, account, welcome)
- **Routing**: `web/src/router/index.ts` — routes: `/`, `/account`, `/streams/:stream?`, `/kv/:bucket?`, `/objects/:store?`, `/tail`

### Key Patterns

- **Handler pattern**: Each API handler struct holds a `*nats.Manager`, uses `requireJetStream()` helper, maps NATS errors to HTTP status codes
- **WebSocket routes** are registered before the Compress middleware group (compression conflicts with WS upgrades)
- **Frontend modals** use `<Teleport to="body">` to escape parent overflow constraints
- **Large message buffers** use `shallowRef` (not `ref`) to avoid deep reactivity overhead
- **The `useTail` composable** connects with subject as URL query param (`/api/v1/ws/tail?subject=orders.>`), not via JSON message — backend reads subject from URL

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NATS_URL` | `nats://localhost:4222` | NATS server URL (also triggers auto-connect + default connection seed) |
| `NATS_MONITOR_URL` | `http://localhost:8222` | NATS monitoring HTTP endpoint |
| `NATS_USERNAME` | | Username auth for default connection |
| `NATS_PASSWORD` | | Password auth (with NATS_USERNAME) |
| `NATS_TOKEN` | | Token auth (takes precedence over username/password) |
| `LISTEN_ADDR` | `:8080` | HTTP listen address |
| `DATA_DIR` | `~/.nats-insight` | SQLite database directory |
| `SEED_TEST_DATA` | | Set to `true` to create test streams/consumers/messages on startup |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |

## Style Conventions

### Go
- Package `nats` (internal) conflicts with `nats.go` client — use `natsclient` alias
- NATS manager import: `natsmgr "github.com/seralaci/nats-insight/internal/nats"` in handler files
- Error mapping: `jetstream.ErrBucketNotFound` → 404, `jetstream.ErrStreamNotFound` → 404, etc.
- Duration fields: accept Go duration strings ("5m", "24h"), size fields: accept human-readable ("100MB", "1GB")

### Vue/TypeScript
- No `transition-colors` or CSS transitions (causes flicker on dark/light mode switch)
- Dark mode classes: `bg-white dark:bg-gray-950`, `text-gray-900 dark:text-gray-100`, `border-gray-200 dark:border-gray-800`
- Input class: `w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent`
- Card style: `bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4`
- Primary: emerald-600 (dark: emerald-500), accent: emerald-400, Destructive: red-500/600
- Use `text-sm` for most text, `text-xs` for secondary

## Agent-First Workflow

Always delegate implementation work to specialized agents. Act as coordinator, not implementer.

Launch backend and frontend agents in parallel when tasks are independent. Coordinate results, fix build errors, and commit.

