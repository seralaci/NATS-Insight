package api

import (
	"encoding/base64"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/nats-io/nats.go"
	natsmgr "github.com/seralaci/nats-insight/internal/nats"
)

// PublishHandler handles message publish and request/reply endpoints.
type PublishHandler struct {
	manager *natsmgr.Manager
}

// NewPublishHandler creates a new PublishHandler.
func NewPublishHandler(m *natsmgr.Manager) *PublishHandler {
	return &PublishHandler{manager: m}
}

type publishRequest struct {
	Subject string              `json:"subject"`
	Data    string              `json:"data"`    // plain text or base64
	Headers map[string][]string `json:"headers"`
}

type requestRequest struct {
	Subject string              `json:"subject"`
	Data    string              `json:"data"`
	Headers map[string][]string `json:"headers"`
	Timeout string              `json:"timeout"` // Go duration, default "5s"
}

type publishResponse struct {
	Status string `json:"status"`
}

type requestResponse struct {
	Subject  string              `json:"subject"`
	Data     string              `json:"data"`
	DataText string              `json:"dataText,omitempty"`
	Headers  map[string][]string `json:"headers,omitempty"`
}

// Publish handles POST /api/v1/publish
func (h *PublishHandler) Publish(w http.ResponseWriter, r *http.Request) {
	conn, err := h.manager.Conn()
	if err != nil {
		Error(w, http.StatusServiceUnavailable, "not connected to NATS")
		return
	}

	var req publishRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Subject == "" {
		Error(w, http.StatusBadRequest, "subject is required")
		return
	}

	msg := &nats.Msg{
		Subject: req.Subject,
		Data:    []byte(req.Data),
	}

	if len(req.Headers) > 0 {
		msg.Header = nats.Header(req.Headers)
	}

	if err := conn.PublishMsg(msg); err != nil {
		Error(w, http.StatusInternalServerError, "failed to publish: "+err.Error())
		return
	}

	JSON(w, http.StatusOK, publishResponse{Status: "published"})
}

// Request handles POST /api/v1/request
func (h *PublishHandler) Request(w http.ResponseWriter, r *http.Request) {
	conn, err := h.manager.Conn()
	if err != nil {
		Error(w, http.StatusServiceUnavailable, "not connected to NATS")
		return
	}

	var req requestRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Subject == "" {
		Error(w, http.StatusBadRequest, "subject is required")
		return
	}

	timeout := 5 * time.Second
	if req.Timeout != "" {
		d, err := time.ParseDuration(req.Timeout)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid timeout duration: "+err.Error())
			return
		}
		timeout = d
	}

	msg := &nats.Msg{
		Subject: req.Subject,
		Data:    []byte(req.Data),
	}
	if len(req.Headers) > 0 {
		msg.Header = nats.Header(req.Headers)
	}

	reply, err := conn.RequestMsg(msg, timeout)
	if err != nil {
		Error(w, http.StatusInternalServerError, "request failed: "+err.Error())
		return
	}

	resp := requestResponse{
		Subject: reply.Subject,
		Data:    base64.StdEncoding.EncodeToString(reply.Data),
	}
	if utf8.Valid(reply.Data) {
		resp.DataText = string(reply.Data)
	}
	if len(reply.Header) > 0 {
		resp.Headers = map[string][]string(reply.Header)
	}

	JSON(w, http.StatusOK, resp)
}
