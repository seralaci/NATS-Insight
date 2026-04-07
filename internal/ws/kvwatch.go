package ws

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/seralaci/nats-insight/internal/nats"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type KvWatchHandler struct {
	manager *nats.Manager
}

func NewKvWatchHandler(manager *nats.Manager) *KvWatchHandler {
	return &KvWatchHandler{manager: manager}
}

type watchRequest struct {
	Type      string `json:"type"`
	Bucket    string `json:"bucket"`
	KeyFilter string `json:"keyFilter"`
}

type kvEntry struct {
	Type      string `json:"type"`
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	ValueText string `json:"valueText,omitempty"`
	Revision  uint64 `json:"revision"`
	Operation string `json:"operation"`
	Created   string `json:"created"`
	Size      int    `json:"size"`
}

func operationString(op jetstream.KeyValueOp) string {
	switch op {
	case jetstream.KeyValuePut:
		return "put"
	case jetstream.KeyValueDelete:
		return "delete"
	case jetstream.KeyValuePurge:
		return "purge"
	default:
		return "unknown"
	}
}

func (h *KvWatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	js, err := h.manager.JetStream()
	if err != nil {
		http.Error(w, "JetStream not available", http.StatusServiceUnavailable)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		slog.Error("WebSocket accept failed", "error", err)
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	// Read the initial watch request from the client.
	var req watchRequest
	if err := wsjson.Read(ctx, conn, &req); err != nil {
		slog.Debug("KvWatch: failed to read watch request", "error", err)
		return
	}

	if req.Type != "watch" || req.Bucket == "" {
		_ = wsjson.Write(ctx, conn, map[string]string{"type": "error", "message": "expected {type:watch, bucket:...}"})
		conn.Close(websocket.StatusPolicyViolation, "invalid request")
		return
	}

	kv, err := js.KeyValue(ctx, req.Bucket)
	if err != nil {
		_ = wsjson.Write(ctx, conn, map[string]string{"type": "error", "message": err.Error()})
		conn.Close(websocket.StatusInternalError, "bucket not found")
		return
	}

	keyFilter := strings.TrimSpace(req.KeyFilter)

	var watcher jetstream.KeyWatcher
	if keyFilter == "" || keyFilter == ">" {
		watcher, err = kv.WatchAll(ctx, jetstream.UpdatesOnly())
	} else {
		watcher, err = kv.Watch(ctx, keyFilter, jetstream.UpdatesOnly())
	}
	if err != nil {
		_ = wsjson.Write(ctx, conn, map[string]string{"type": "error", "message": err.Error()})
		conn.Close(websocket.StatusInternalError, "watch failed")
		return
	}
	defer func() {
		_ = watcher.Stop()
	}()

	_ = wsjson.Write(ctx, conn, map[string]string{"type": "watching", "bucket": req.Bucket})

	// Goroutine to detect stop/disconnect from client.
	stopCh := make(chan struct{})
	go func() {
		defer close(stopCh)
		for {
			var cm clientMsg
			if err := wsjson.Read(ctx, conn, &cm); err != nil {
				return
			}
			if cm.Type == "stop" {
				return
			}
		}
	}()

	updates := watcher.Updates()

	for {
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "context done")
			return
		case <-stopCh:
			_ = wsjson.Write(ctx, conn, map[string]string{"type": "stopped"})
			conn.Close(websocket.StatusNormalClosure, "stopped")
			return
		case entry, ok := <-updates:
			if !ok {
				conn.Close(websocket.StatusNormalClosure, "watcher closed")
				return
			}

			// nil entry signals initial-values-done from the watcher.
			if entry == nil {
				_ = wsjson.Write(ctx, conn, map[string]string{"type": "init_done"})
				continue
			}

			val := entry.Value()
			encoded := base64.StdEncoding.EncodeToString(val)
			var valueText string
			if utf8.Valid(val) {
				valueText = string(val)
			}

			out := kvEntry{
				Type:      "entry",
				Bucket:    req.Bucket,
				Key:       entry.Key(),
				Value:     encoded,
				ValueText: valueText,
				Revision:  entry.Revision(),
				Operation: operationString(entry.Operation()),
				Created:   entry.Created().UTC().Format("2006-01-02T15:04:05Z"),
				Size:      len(val),
			}

			if err := wsjson.Write(ctx, conn, out); err != nil {
				slog.Debug("KvWatch: write failed", "error", err)
				return
			}
		}
	}
}
