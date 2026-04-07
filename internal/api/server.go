package api

import (
	"encoding/json"
	"net/http"

	natsmgr "github.com/seralaci/nats-insight/internal/nats"
)

// AccountLimits holds JetStream account-level resource limits.
type AccountLimits struct {
	MaxMemory           int64 `json:"max_memory"`
	MaxStorage          int64 `json:"max_storage"`
	MaxStreams          int   `json:"max_streams"`
	MaxConsumers        int   `json:"max_consumers"`
	MaxAckPending       int   `json:"max_ack_pending"`
	MemoryMaxStreamBytes int64 `json:"memory_max_stream_bytes"`
	StoreMaxStreamBytes int64 `json:"store_max_stream_bytes"`
	MaxBytesRequired    bool  `json:"max_bytes_required"`
}

// AccountInfoResponse is the combined response for GET /api/v1/server/account-info.
type AccountInfoResponse struct {
	// From varz
	ServerName    string `json:"server_name"`
	ServerVersion string `json:"server_version"`
	MaxPayload    int64  `json:"max_payload"`

	// From JetStream AccountInfo
	JetStream bool   `json:"jetstream"`
	Memory    uint64 `json:"memory"`
	Storage   uint64 `json:"storage"`
	Streams   int    `json:"streams"`
	Consumers int    `json:"consumers"`
	Domain    string `json:"domain"`
	APITotal  uint64 `json:"api_total"`
	APIErrors uint64 `json:"api_errors"`

	// Limits (nil when JetStream is unavailable)
	Limits *AccountLimits `json:"limits,omitempty"`
}

type ServerHandler struct {
	manager *natsmgr.Manager
}

func NewServerHandler(m *natsmgr.Manager) *ServerHandler {
	return &ServerHandler{manager: m}
}

func (h *ServerHandler) requireConnection(w http.ResponseWriter) bool {
	if !h.manager.IsConnected() {
		Error(w, http.StatusServiceUnavailable, "not connected to NATS")
		return false
	}
	return true
}

func (h *ServerHandler) proxyMonitor(w http.ResponseWriter, endpoint string) {
	data, err := h.manager.FetchMonitorEndpoint(endpoint)
	if err != nil {
		Error(w, http.StatusBadGateway, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *ServerHandler) Varz(w http.ResponseWriter, r *http.Request) {
	if !h.requireConnection(w) {
		return
	}
	h.proxyMonitor(w, "/varz")
}

func (h *ServerHandler) Jsz(w http.ResponseWriter, r *http.Request) {
	if !h.requireConnection(w) {
		return
	}
	h.proxyMonitor(w, "/jsz")
}

func (h *ServerHandler) Connz(w http.ResponseWriter, r *http.Request) {
	if !h.requireConnection(w) {
		return
	}
	h.proxyMonitor(w, "/connz")
}

func (h *ServerHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	if !h.requireConnection(w) {
		return
	}
	h.proxyMonitor(w, "/healthz")
}

func (h *ServerHandler) Accountz(w http.ResponseWriter, r *http.Request) {
	if !h.requireConnection(w) {
		return
	}
	h.proxyMonitor(w, "/accountz")
}

// AccountInfo returns a combined response of varz server info and JetStream
// account-level usage/limits obtained via the JetStream client API.
func (h *ServerHandler) AccountInfo(w http.ResponseWriter, r *http.Request) {
	if !h.requireConnection(w) {
		return
	}

	resp := AccountInfoResponse{}

	// --- varz: server name, version, max_payload ---
	varzData, err := h.manager.FetchMonitorEndpoint("/varz")
	if err != nil {
		Error(w, http.StatusBadGateway, "failed to fetch varz: "+err.Error())
		return
	}
	var varz struct {
		ServerName string `json:"server_name"`
		Version    string `json:"version"`
		MaxPayload int64  `json:"max_payload"`
	}
	if err := json.Unmarshal(varzData, &varz); err != nil {
		Error(w, http.StatusInternalServerError, "failed to parse varz: "+err.Error())
		return
	}
	resp.ServerName = varz.ServerName
	resp.ServerVersion = varz.Version
	resp.MaxPayload = varz.MaxPayload

	// --- JetStream AccountInfo (best-effort; not fatal if unavailable) ---
	js, err := h.manager.JetStream()
	if err == nil {
		info, err := js.AccountInfo(r.Context())
		if err == nil {
			resp.JetStream = true
			resp.Memory = info.Memory
			resp.Storage = info.Store
			resp.Streams = info.Streams
			resp.Consumers = info.Consumers
			resp.Domain = info.Domain
			resp.APITotal = info.API.Total
			resp.APIErrors = info.API.Errors
			resp.Limits = &AccountLimits{
				MaxMemory:            info.Limits.MaxMemory,
				MaxStorage:           info.Limits.MaxStore,
				MaxStreams:           info.Limits.MaxStreams,
				MaxConsumers:         info.Limits.MaxConsumers,
				MaxAckPending:        info.Limits.MaxAckPending,
				MemoryMaxStreamBytes: info.Limits.MemoryMaxStreamBytes,
				StoreMaxStreamBytes:  info.Limits.StoreMaxStreamBytes,
				MaxBytesRequired:     info.Limits.MaxBytesRequired,
			}
		}
	}

	JSON(w, http.StatusOK, resp)
}
