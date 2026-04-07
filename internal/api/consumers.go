package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go/jetstream"
	natsmgr "github.com/seralaci/nats-insight/internal/nats"
)

// ConsumerHandler handles JetStream consumer REST endpoints.
type ConsumerHandler struct {
	manager *natsmgr.Manager
}

// NewConsumerHandler creates a new ConsumerHandler.
func NewConsumerHandler(m *natsmgr.Manager) *ConsumerHandler {
	return &ConsumerHandler{manager: m}
}

func (h *ConsumerHandler) requireJetStream(w http.ResponseWriter) (jetstream.JetStream, bool) {
	js, err := h.manager.JetStream()
	if err != nil {
		Error(w, http.StatusServiceUnavailable, "JetStream not available: "+err.Error())
		return nil, false
	}
	return js, true
}

func consumerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, jetstream.ErrStreamNotFound):
		Error(w, http.StatusNotFound, "stream not found")
	case errors.Is(err, jetstream.ErrConsumerNotFound):
		Error(w, http.StatusNotFound, "consumer not found")
	default:
		Error(w, http.StatusInternalServerError, natsErrMsg(err))
	}
}

// List handles GET /api/v1/streams/{stream}/consumers
func (h *ConsumerHandler) List(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	streamName := chi.URLParam(r, "stream")
	ctx := r.Context()

	stream, err := js.Stream(ctx, streamName)
	if err != nil {
		consumerError(w, err)
		return
	}

	lister := stream.ListConsumers(ctx)

	var consumers []*jetstream.ConsumerInfo
	for info := range lister.Info() {
		consumers = append(consumers, info)
	}
	if err := lister.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "failed to list consumers: "+natsErrMsg(err))
		return
	}

	if consumers == nil {
		consumers = []*jetstream.ConsumerInfo{}
	}
	JSON(w, http.StatusOK, consumers)
}

// Get handles GET /api/v1/streams/{stream}/consumers/{consumer}
func (h *ConsumerHandler) Get(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	streamName := chi.URLParam(r, "stream")
	consumerName := chi.URLParam(r, "consumer")
	ctx := r.Context()

	consumer, err := js.Consumer(ctx, streamName, consumerName)
	if err != nil {
		consumerError(w, err)
		return
	}

	info, err := consumer.Info(ctx)
	if err != nil {
		consumerError(w, err)
		return
	}

	JSON(w, http.StatusOK, info)
}

// Delete handles DELETE /api/v1/streams/{stream}/consumers/{consumer}
func (h *ConsumerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	streamName := chi.URLParam(r, "stream")
	consumerName := chi.URLParam(r, "consumer")
	ctx := r.Context()

	if err := js.DeleteConsumer(ctx, streamName, consumerName); err != nil {
		consumerError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// pauseConsumerRequest is the JSON body for pausing a consumer.
type pauseConsumerRequest struct {
	PauseUntil time.Time `json:"pauseUntil"`
}

// Pause handles POST /api/v1/streams/{stream}/consumers/{consumer}/pause
func (h *ConsumerHandler) Pause(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	streamName := chi.URLParam(r, "stream")
	consumerName := chi.URLParam(r, "consumer")

	var req pauseConsumerRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PauseUntil.IsZero() {
		Error(w, http.StatusBadRequest, "pauseUntil is required")
		return
	}

	ctx := r.Context()
	resp, err := js.PauseConsumer(ctx, streamName, consumerName, req.PauseUntil)
	if err != nil {
		consumerError(w, err)
		return
	}

	JSON(w, http.StatusOK, resp)
}

// Resume handles POST /api/v1/streams/{stream}/consumers/{consumer}/resume
func (h *ConsumerHandler) Resume(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	streamName := chi.URLParam(r, "stream")
	consumerName := chi.URLParam(r, "consumer")
	ctx := r.Context()

	resp, err := js.ResumeConsumer(ctx, streamName, consumerName)
	if err != nil {
		consumerError(w, err)
		return
	}

	JSON(w, http.StatusOK, resp)
}
