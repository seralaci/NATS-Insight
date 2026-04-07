package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/seralaci/nats-insight/internal/api"
	"github.com/seralaci/nats-insight/internal/config"
	"github.com/seralaci/nats-insight/internal/nats"
	"github.com/seralaci/nats-insight/internal/store"
)

//go:embed all:web
var webFS embed.FS

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	// Initialize store
	dataDir := cfg.DataDir
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".nats-insight")
	}
	db, err := store.New(dataDir)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize NATS manager
	manager := nats.NewManager()

	// Seed default connection and auto-connect if NATS_URL is set
	seedAndAutoConnect(db, cfg, manager)

	// Seed test data if SEED_TEST_DATA env var is set
	if os.Getenv("SEED_TEST_DATA") == "true" {
		go seedTestData(manager)
	}

	// Setup API router
	r := api.NewRouter(db, manager)

	// Serve embedded frontend
	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		slog.Error("failed to create sub filesystem", "error", err)
		os.Exit(1)
	}
	fileServer := http.FileServer(http.FS(webRoot))

	// SPA fallback: serve index.html for any non-API, non-file route
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		if _, err := fs.Stat(webRoot, path[1:]); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	slog.Info("starting NATS Insight", "addr", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func seedTestData(manager *nats.Manager) {
	// Wait for the connection to be established
	time.Sleep(3 * time.Second)

	js, err := manager.JetStream()
	if err != nil {
		slog.Warn("seedTestData: JetStream not available, skipping", "error", err)
		return
	}

	ctx := context.Background()

	// Create ORDERS stream
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        "ORDERS",
		Description: "Order processing stream",
		Subjects:    []string{"orders.>"},
		Retention:   jetstream.LimitsPolicy,
		Storage:     jetstream.FileStorage,
		MaxAge:      24 * time.Hour,
		Replicas:    1,
	})
	if err != nil {
		slog.Warn("seedTestData: failed to create ORDERS stream", "error", err)
	} else {
		slog.Info("seedTestData: ORDERS stream ready")
	}

	// Create EVENTS stream
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        "EVENTS",
		Description: "Application events",
		Subjects:    []string{"events.>"},
		Storage:     jetstream.FileStorage,
		Replicas:    1,
	})
	if err != nil {
		slog.Warn("seedTestData: failed to create EVENTS stream", "error", err)
	} else {
		slog.Info("seedTestData: EVENTS stream ready")
	}

	// Create durable consumer: order-processor
	_, err = js.CreateOrUpdateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
		Durable:       "order-processor",
		FilterSubject: "orders.>",
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
	})
	if err != nil {
		slog.Warn("seedTestData: failed to create order-processor consumer", "error", err)
	} else {
		slog.Info("seedTestData: order-processor consumer ready")
	}

	// Create durable consumer: order-analytics
	_, err = js.CreateOrUpdateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
		Durable:       "order-analytics",
		FilterSubject: "orders.created",
		AckPolicy:     jetstream.AckNonePolicy,
		DeliverPolicy: jetstream.DeliverNewPolicy,
	})
	if err != nil {
		slog.Warn("seedTestData: failed to create order-analytics consumer", "error", err)
	} else {
		slog.Info("seedTestData: order-analytics consumer ready")
	}

	// Publish messages on a rotating schedule
	subjects := []string{
		"orders.created",
		"orders.shipped",
		"orders.completed",
		"events.user.login",
		"events.user.logout",
	}
	msgTypes := []string{
		"order.created",
		"order.shipped",
		"order.completed",
		"user.login",
		"user.logout",
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		var msgID int
		for range ticker.C {
			nc, connErr := manager.Conn()
			if connErr != nil {
				slog.Warn("seedTestData: not connected, skipping publish", "error", connErr)
				continue
			}

			idx := msgID % len(subjects)
			subject := subjects[idx]
			msgType := msgTypes[idx]
			msgID++

			payload := map[string]any{
				"id":        msgID,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"type":      msgType,
				"data": map[string]any{
					"customer": fmt.Sprintf("User %d", msgID),
					"amount":   float64(msgID) * 10.50,
				},
			}
			data, jsonErr := json.Marshal(payload)
			if jsonErr != nil {
				slog.Warn("seedTestData: failed to marshal message", "error", jsonErr)
				continue
			}

			if pubErr := nc.Publish(subject, data); pubErr != nil {
				slog.Warn("seedTestData: publish failed", "subject", subject, "error", pubErr)
			} else {
				slog.Info("seedTestData: published message", "subject", subject, "id", msgID)
			}
		}
	}()
}

func seedAndAutoConnect(db *store.Store, cfg *config.Config, manager *nats.Manager) {
	if os.Getenv("NATS_URL") == "" {
		return
	}

	conns, err := db.ListConnections()
	if err != nil {
		return
	}

	// Seed default connection if none exist
	var conn *store.Connection
	if len(conns) == 0 {
		conn = &store.Connection{
			Name:       "Default",
			URL:        cfg.NATSUrl,
			AuthMethod: "none",
			MonitorURL: cfg.MonitorUrl,
		}
		// Set auth from env vars: NATS_TOKEN takes precedence over NATS_USERNAME/NATS_PASSWORD
		if cfg.NATSToken != "" {
			conn.AuthMethod = "token"
			conn.Token = cfg.NATSToken
		} else if cfg.NATSUsername != "" {
			conn.AuthMethod = "username_password"
			conn.Username = cfg.NATSUsername
			conn.Password = cfg.NATSPassword
		}
		if err := db.CreateConnection(conn); err != nil {
			slog.Warn("failed to seed default connection", "error", err)
			return
		}
		slog.Info("seeded default connection", "url", cfg.NATSUrl, "auth", conn.AuthMethod)
	} else {
		conn = &conns[0]
	}

	// Auto-connect to the first connection
	if err := manager.Connect(conn); err != nil {
		slog.Warn("auto-connect failed", "error", err)
	} else {
		slog.Info("auto-connected to NATS", "url", conn.URL)
	}
}
