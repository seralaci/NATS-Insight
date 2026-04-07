package ws

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"time"
	"unicode/utf8"

	natsclient "github.com/nats-io/nats.go"
	"github.com/seralaci/nats-insight/internal/nats"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type TailHandler struct {
	manager *nats.Manager
}

func NewTailHandler(manager *nats.Manager) *TailHandler {
	return &TailHandler{manager: manager}
}

type tailStartedMsg struct {
	Type    string `json:"type"`
	Subject string `json:"subject"`
}

type tailMessage struct {
	Type       string              `json:"type"`
	Subject    string              `json:"subject"`
	Data       string              `json:"data"`
	DataText   string              `json:"dataText,omitempty"`
	Headers    map[string][]string `json:"headers,omitempty"`
	ReceivedAt string              `json:"receivedAt"`
	Size       int                 `json:"size"`
}

type clientMsg struct {
	Type string `json:"type"`
}

func (h *TailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subject := r.URL.Query().Get("subject")
	if subject == "" {
		http.Error(w, "subject query param is required", http.StatusBadRequest)
		return
	}

	nc, err := h.manager.Conn()
	if err != nil {
		http.Error(w, "not connected to NATS", http.StatusServiceUnavailable)
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

	msgCh := make(chan *natsclient.Msg, 256)
	sub, err := nc.ChanSubscribe(subject, msgCh)
	if err != nil {
		_ = wsjson.Write(ctx, conn, map[string]string{"type": "error", "message": err.Error()})
		conn.Close(websocket.StatusInternalError, "subscribe failed")
		return
	}
	defer func() {
		_ = sub.Unsubscribe()
	}()

	if err := wsjson.Write(ctx, conn, tailStartedMsg{Type: "started", Subject: subject}); err != nil {
		return
	}

	// Goroutine to read client control messages (stop / disconnect detection).
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

	for {
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "context done")
			return
		case <-stopCh:
			_ = wsjson.Write(ctx, conn, map[string]string{"type": "stopped"})
			conn.Close(websocket.StatusNormalClosure, "stopped")
			return
		case msg, ok := <-msgCh:
			if !ok {
				conn.Close(websocket.StatusNormalClosure, "subscription closed")
				return
			}

			encoded := base64.StdEncoding.EncodeToString(msg.Data)
			var dataText string
			if utf8.Valid(msg.Data) {
				dataText = string(msg.Data)
			}

			var hdrs map[string][]string
			if msg.Header != nil {
				hdrs = map[string][]string(msg.Header)
			}

			out := tailMessage{
				Type:       "message",
				Subject:    msg.Subject,
				Data:       encoded,
				DataText:   dataText,
				Headers:    hdrs,
				ReceivedAt: time.Now().UTC().Format(time.RFC3339Nano),
				Size:       len(msg.Data),
			}

			if err := wsjson.Write(ctx, conn, out); err != nil {
				slog.Debug("WebSocket write failed", "error", err)
				return
			}
		}
	}
}
