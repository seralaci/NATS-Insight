package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/seralaci/nats-insight/internal/nats"
	"github.com/seralaci/nats-insight/internal/store"
)

type ConnectionHandler struct {
	store   *store.Store
	manager *nats.Manager
}

func NewConnectionHandler(s *store.Store, m *nats.Manager) *ConnectionHandler {
	return &ConnectionHandler{store: s, manager: m}
}

func (h *ConnectionHandler) List(w http.ResponseWriter, r *http.Request) {
	conns, err := h.store.ListConnections()
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list connections")
		return
	}
	JSON(w, http.StatusOK, conns)
}

func (h *ConnectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn, err := h.store.GetConnection(id)
	if errors.Is(err, store.ErrNotFound) {
		Error(w, http.StatusNotFound, "connection not found")
		return
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get connection")
		return
	}
	JSON(w, http.StatusOK, conn)
}

func (h *ConnectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var conn store.Connection
	if err := DecodeJSON(r, &conn); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if conn.Name == "" {
		Error(w, http.StatusBadRequest, "connection name is required")
		return
	}
	if conn.URL == "" {
		conn.URL = "nats://localhost:4222"
	}
	if conn.AuthMethod == "" {
		conn.AuthMethod = "none"
	}
	if conn.MonitorURL == "" {
		conn.MonitorURL = "http://localhost:8222"
	}

	if err := h.store.CreateConnection(&conn); err != nil {
		Error(w, http.StatusInternalServerError, "failed to create connection")
		return
	}
	JSON(w, http.StatusCreated, conn)
}

func (h *ConnectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var conn store.Connection
	if err := DecodeJSON(r, &conn); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	conn.ID = id

	if conn.Name == "" {
		Error(w, http.StatusBadRequest, "connection name is required")
		return
	}

	if err := h.store.UpdateConnection(&conn); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			Error(w, http.StatusNotFound, "connection not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to update connection")
		return
	}

	updated, _ := h.store.GetConnection(id)
	JSON(w, http.StatusOK, updated)
}

func (h *ConnectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Disconnect if this is the active connection
	if h.manager.ActiveConnectionID() == id {
		h.manager.Disconnect()
	}

	if err := h.store.DeleteConnection(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			Error(w, http.StatusNotFound, "connection not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to delete connection")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *ConnectionHandler) Connect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn, err := h.store.GetConnection(id)
	if errors.Is(err, store.ErrNotFound) {
		Error(w, http.StatusNotFound, "connection not found")
		return
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get connection")
		return
	}

	if err := h.manager.Connect(conn); err != nil {
		Error(w, http.StatusBadGateway, err.Error())
		return
	}

	JSON(w, http.StatusOK, h.manager.Status())
}

func (h *ConnectionHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	h.manager.Disconnect()
	JSON(w, http.StatusOK, h.manager.Status())
}

func (h *ConnectionHandler) Status(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, h.manager.Status())
}
