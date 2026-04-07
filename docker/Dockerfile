# Stage 1: Build frontend
FROM node:24-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.26-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/cmd/nats-insight/web ./cmd/nats-insight/web
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /nats-insight ./cmd/nats-insight/

# Stage 3: Final image
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=backend-builder /nats-insight /usr/local/bin/nats-insight
EXPOSE 8080
ENTRYPOINT ["nats-insight"]
