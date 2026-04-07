package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/seralaci/nats-insight/internal/nats"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type MetricsHandler struct {
	manager *nats.Manager
}

func NewMetricsHandler(manager *nats.Manager) *MetricsHandler {
	return &MetricsHandler{manager: manager}
}

type metricsCommand struct {
	Type     string `json:"type"`
	Interval int    `json:"interval"`
}

type metricsMessage struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp,omitempty"`
	Varz      json.RawMessage `json:"varz,omitempty"`
	Jsz       json.RawMessage `json:"jsz,omitempty"`
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		slog.Error("WebSocket accept failed", "error", err)
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	var cmd metricsCommand
	if err := wsjson.Read(ctx, conn, &cmd); err != nil {
		slog.Debug("metrics: failed to read start command", "error", err)
		return
	}

	if cmd.Type != "start" {
		_ = wsjson.Write(ctx, conn, map[string]string{"type": "error", "message": "expected {type:start}"})
		conn.Close(websocket.StatusPolicyViolation, "invalid request")
		return
	}

	interval := cmd.Interval
	if interval < 1 {
		interval = 5
	}
	if interval > 60 {
		interval = 60
	}

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

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "context done")
			return
		case <-stopCh:
			_ = wsjson.Write(ctx, conn, map[string]string{"type": "stopped"})
			conn.Close(websocket.StatusNormalClosure, "stopped")
			return
		case <-ticker.C:
			varz, err := h.manager.FetchMonitorEndpoint("/varz")
			if err != nil {
				slog.Debug("metrics: failed to fetch varz", "error", err)
				varz = nil
			}

			jsz, err := h.manager.FetchMonitorEndpoint("/jsz")
			if err != nil {
				slog.Debug("metrics: failed to fetch jsz", "error", err)
				jsz = nil
			}

			msg := metricsMessage{
				Type:      "metrics",
				Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
				Varz:      varz,
				Jsz:       jsz,
			}

			if err := wsjson.Write(ctx, conn, msg); err != nil {
				slog.Debug("metrics: write failed", "error", err)
				return
			}
		}
	}
}
