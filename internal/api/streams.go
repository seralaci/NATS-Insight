package api

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go/jetstream"
	natsmgr "github.com/seralaci/nats-insight/internal/nats"
)

// StreamHandler handles JetStream stream REST endpoints.
type StreamHandler struct {
	manager *natsmgr.Manager
}

// NewStreamHandler creates a new StreamHandler.
func NewStreamHandler(m *natsmgr.Manager) *StreamHandler {
	return &StreamHandler{manager: m}
}

func (h *StreamHandler) requireJetStream(w http.ResponseWriter) (jetstream.JetStream, bool) {
	js, err := h.manager.JetStream()
	if err != nil {
		Error(w, http.StatusServiceUnavailable, "JetStream not available: "+err.Error())
		return nil, false
	}
	return js, true
}

func streamError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, jetstream.ErrStreamNotFound):
		Error(w, http.StatusNotFound, "stream not found")
	case errors.Is(err, jetstream.ErrMsgNotFound):
		Error(w, http.StatusNotFound, "message not found")
	default:
		Error(w, http.StatusInternalServerError, natsErrMsg(err))
	}
}

// streamInfoResponse is the JSON representation of a stream.
type streamInfoResponse struct {
	Config  jetstream.StreamConfig `json:"config"`
	State   jetstream.StreamState  `json:"state"`
	Created time.Time              `json:"created"`
}

func streamInfoFromJS(info *jetstream.StreamInfo) streamInfoResponse {
	return streamInfoResponse{
		Config:  info.Config,
		State:   info.State,
		Created: info.Created,
	}
}

// createStreamRequest is the JSON body for creating or updating a stream.
type createStreamRequest struct {
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Subjects             []string          `json:"subjects"`
	Retention            string            `json:"retention"`
	Storage              string            `json:"storage"`
	Replicas             int               `json:"replicas"`
	MaxMsgs              int64             `json:"maxMsgs"`
	MaxBytes             string            `json:"maxBytes"`
	MaxAge               string            `json:"maxAge"`
	MaxMsgSize           string            `json:"maxMsgSize"`
	MaxConsumers         int               `json:"maxConsumers"`
	MaxMsgsPerSubject    int64             `json:"maxMsgsPerSubject"`
	Discard              string            `json:"discard"`
	DiscardNewPerSubject bool              `json:"discardNewPerSubject"`
	Duplicates           string            `json:"duplicates"`
	NoAck                bool              `json:"noAck"`
	Compression          string            `json:"compression"`
	AllowRollup          bool              `json:"allowRollup"`
	DenyDelete           bool              `json:"denyDelete"`
	DenyPurge            bool              `json:"denyPurge"`
	AllowDirect          bool              `json:"allowDirect"`
	MirrorDirect         bool              `json:"mirrorDirect"`
	Metadata             map[string]string `json:"metadata"`
	RepublishSource              string            `json:"republishSource"`
	RepublishDestination         string            `json:"republishDestination"`
	FirstSeq                     uint64            `json:"firstSeq"`
	ConsumerInactiveThreshold    string            `json:"consumerInactiveThreshold"`
	SubjectDeleteMarkerTTL       string            `json:"subjectDeleteMarkerTtl"`
	MaxAckPending                int               `json:"maxAckPending"`
	Tags                         []string          `json:"tags"`
	SubjectTransformSrc          string            `json:"subjectTransformSrc"`
	SubjectTransformDest         string            `json:"subjectTransformDest"`
	AllowMsgCounter              bool              `json:"allowMsgCounter"`
	AllowPerMessageTTL           bool              `json:"allowPerMessageTtl"`
	AllowAtomicPublish           bool              `json:"allowAtomicPublish"`
	AllowMsgSchedules            bool              `json:"allowMsgSchedules"`
	PersistMode                  string            `json:"persistMode"`
	Sources []sourceRequest  `json:"sources"`
	Mirror  *sourceRequest   `json:"mirror"`
}

type sourceRequest struct {
	Name              string                    `json:"name"`
	FilterSubject     string                    `json:"filterSubject"`
	OptStartSeq       uint64                    `json:"optStartSeq"`
	OptStartTime      string                    `json:"optStartTime"`
	External          *externalStreamRequest    `json:"external"`
	SubjectTransforms []subjectTransformRequest `json:"subjectTransforms"`
}

type externalStreamRequest struct {
	APIPrefix     string `json:"apiPrefix"`
	DeliverPrefix string `json:"deliverPrefix"`
}

type subjectTransformRequest struct {
	Source      string `json:"src"`
	Destination string `json:"dest"`
}

func buildStreamConfig(req createStreamRequest) (jetstream.StreamConfig, error) {
	cfg := jetstream.StreamConfig{
		Name:                 req.Name,
		Description:          req.Description,
		Subjects:             req.Subjects,
		MaxMsgs:              req.MaxMsgs,
		MaxConsumers:         req.MaxConsumers,
		MaxMsgsPerSubject:    req.MaxMsgsPerSubject,
		NoAck:                req.NoAck,
		DiscardNewPerSubject: req.DiscardNewPerSubject,
		AllowRollup:          req.AllowRollup,
		DenyDelete:           req.DenyDelete,
		DenyPurge:            req.DenyPurge,
		AllowDirect:          req.AllowDirect,
		MirrorDirect:         req.MirrorDirect,
		Metadata:             req.Metadata,
	}

	// Retention policy
	switch req.Retention {
	case "interest":
		cfg.Retention = jetstream.InterestPolicy
	case "workqueue":
		cfg.Retention = jetstream.WorkQueuePolicy
	default:
		cfg.Retention = jetstream.LimitsPolicy
	}

	// Storage type
	switch req.Storage {
	case "memory":
		cfg.Storage = jetstream.MemoryStorage
	default:
		cfg.Storage = jetstream.FileStorage
	}

	// Replicas
	if req.Replicas > 0 {
		cfg.Replicas = req.Replicas
	} else {
		cfg.Replicas = 1
	}

	// Discard policy
	switch req.Discard {
	case "new":
		cfg.Discard = jetstream.DiscardNew
	default:
		cfg.Discard = jetstream.DiscardOld
	}

	// Compression
	switch req.Compression {
	case "s2":
		cfg.Compression = jetstream.S2Compression
	default:
		cfg.Compression = jetstream.NoCompression
	}

	// MaxBytes
	if req.MaxBytes != "" {
		size, err := ParseByteSize(req.MaxBytes)
		if err != nil {
			return cfg, &validationError{"invalid maxBytes: " + err.Error()}
		}
		cfg.MaxBytes = size
	}

	// MaxMsgSize
	if req.MaxMsgSize != "" {
		size, err := ParseByteSize(req.MaxMsgSize)
		if err != nil {
			return cfg, &validationError{"invalid maxMsgSize: " + err.Error()}
		}
		cfg.MaxMsgSize = int32(size)
	}

	// MaxAge
	if req.MaxAge != "" {
		d, err := time.ParseDuration(req.MaxAge)
		if err != nil {
			return cfg, &validationError{"invalid maxAge duration: " + err.Error()}
		}
		cfg.MaxAge = d
	}

	// Duplicates window
	if req.Duplicates != "" {
		d, err := time.ParseDuration(req.Duplicates)
		if err != nil {
			return cfg, &validationError{"invalid duplicates duration: " + err.Error()}
		}
		cfg.Duplicates = d
	}

	// RePublish
	if req.RepublishSource != "" || req.RepublishDestination != "" {
		cfg.RePublish = &jetstream.RePublish{
			Source:      req.RepublishSource,
			Destination: req.RepublishDestination,
		}
	}

	if req.FirstSeq > 0 {
		cfg.FirstSeq = req.FirstSeq
	}

	if req.SubjectDeleteMarkerTTL != "" {
		d, err := time.ParseDuration(req.SubjectDeleteMarkerTTL)
		if err != nil {
			return cfg, &validationError{"invalid subjectDeleteMarkerTtl duration: " + err.Error()}
		}
		cfg.SubjectDeleteMarkerTTL = d
	}

	if len(req.Tags) > 0 {
		cfg.Placement = &jetstream.Placement{Tags: req.Tags}
	}

	if req.SubjectTransformSrc != "" || req.SubjectTransformDest != "" {
		cfg.SubjectTransform = &jetstream.SubjectTransformConfig{
			Source:      req.SubjectTransformSrc,
			Destination: req.SubjectTransformDest,
		}
	}

	cfg.AllowMsgCounter = req.AllowMsgCounter
	cfg.AllowMsgTTL = req.AllowPerMessageTTL
	cfg.AllowAtomicPublish = req.AllowAtomicPublish
	cfg.AllowMsgSchedules = req.AllowMsgSchedules

	switch req.PersistMode {
	case "async":
		cfg.PersistMode = jetstream.AsyncPersistMode
	default:
		cfg.PersistMode = jetstream.DefaultPersistMode
	}

	if req.ConsumerInactiveThreshold != "" || req.MaxAckPending > 0 {
		cfg.ConsumerLimits = jetstream.StreamConsumerLimits{}
		if req.ConsumerInactiveThreshold != "" {
			d, err := time.ParseDuration(req.ConsumerInactiveThreshold)
			if err != nil {
				return cfg, &validationError{"invalid consumerInactiveThreshold duration: " + err.Error()}
			}
			cfg.ConsumerLimits.InactiveThreshold = d
		}
		if req.MaxAckPending > 0 {
			cfg.ConsumerLimits.MaxAckPending = req.MaxAckPending
		}
	}

	// Sources
	if len(req.Sources) > 0 {
		for _, s := range req.Sources {
			src, err := buildStreamSource(s)
			if err != nil {
				return cfg, err
			}
			cfg.Sources = append(cfg.Sources, src)
		}
	}

	// Mirror
	if req.Mirror != nil {
		src, err := buildStreamSource(*req.Mirror)
		if err != nil {
			return cfg, err
		}
		cfg.Mirror = src
	}

	return cfg, nil
}

func buildStreamSource(s sourceRequest) (*jetstream.StreamSource, error) {
	src := &jetstream.StreamSource{
		Name:          s.Name,
		OptStartSeq:   s.OptStartSeq,
		FilterSubject: s.FilterSubject,
	}
	if s.OptStartTime != "" {
		t, err := time.Parse(time.RFC3339, s.OptStartTime)
		if err != nil {
			return nil, &validationError{"invalid source optStartTime: " + err.Error()}
		}
		src.OptStartTime = &t
	}
	if s.External != nil {
		src.External = &jetstream.ExternalStream{
			APIPrefix:     s.External.APIPrefix,
			DeliverPrefix: s.External.DeliverPrefix,
		}
	}
	for _, t := range s.SubjectTransforms {
		src.SubjectTransforms = append(src.SubjectTransforms, jetstream.SubjectTransformConfig{
			Source:      t.Source,
			Destination: t.Destination,
		})
	}
	return src, nil
}

// validationError is used to distinguish bad-request errors from internal errors.
type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }

// List handles GET /api/v1/streams
func (h *StreamHandler) List(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	ctx := r.Context()
	lister := js.ListStreams(ctx)

	var streams []streamInfoResponse
	for info := range lister.Info() {
		streams = append(streams, streamInfoFromJS(info))
	}
	if err := lister.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "failed to list streams: "+natsErrMsg(err))
		return
	}

	if streams == nil {
		streams = []streamInfoResponse{}
	}
	JSON(w, http.StatusOK, streams)
}

// Create handles POST /api/v1/streams
func (h *StreamHandler) Create(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	var req createStreamRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		Error(w, http.StatusBadRequest, "name is required")
		return
	}

	cfg, err := buildStreamConfig(req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	stream, err := js.CreateStream(ctx, cfg)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create stream: "+natsErrMsg(err))
		return
	}

	info, err := stream.Info(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "stream created but failed to get info: "+natsErrMsg(err))
		return
	}

	JSON(w, http.StatusCreated, streamInfoFromJS(info))
}

// Get handles GET /api/v1/streams/{stream}
func (h *StreamHandler) Get(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "stream")
	ctx := r.Context()

	stream, err := js.Stream(ctx, name)
	if err != nil {
		streamError(w, err)
		return
	}

	info, err := stream.Info(ctx)
	if err != nil {
		streamError(w, err)
		return
	}

	JSON(w, http.StatusOK, streamInfoFromJS(info))
}

// Update handles PUT /api/v1/streams/{stream}
func (h *StreamHandler) Update(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "stream")

	var req createStreamRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Name in config must match URL param
	req.Name = name

	cfg, err := buildStreamConfig(req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	stream, err := js.UpdateStream(ctx, cfg)
	if err != nil {
		streamError(w, err)
		return
	}

	info, err := stream.Info(ctx)
	if err != nil {
		streamError(w, err)
		return
	}

	JSON(w, http.StatusOK, streamInfoFromJS(info))
}

// Delete handles DELETE /api/v1/streams/{stream}
func (h *StreamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "stream")
	ctx := r.Context()

	if err := js.DeleteStream(ctx, name); err != nil {
		streamError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Purge handles POST /api/v1/streams/{stream}/purge
func (h *StreamHandler) Purge(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "stream")
	ctx := r.Context()

	stream, err := js.Stream(ctx, name)
	if err != nil {
		streamError(w, err)
		return
	}

	var opts []jetstream.StreamPurgeOpt

	if subj := r.URL.Query().Get("subject"); subj != "" {
		opts = append(opts, jetstream.WithPurgeSubject(subj))
	}
	if seqStr := r.URL.Query().Get("seq"); seqStr != "" {
		seq, err := strconv.ParseUint(seqStr, 10, 64)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid seq parameter")
			return
		}
		opts = append(opts, jetstream.WithPurgeSequence(seq))
	}
	if keepStr := r.URL.Query().Get("keep"); keepStr != "" {
		keep, err := strconv.ParseUint(keepStr, 10, 64)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid keep parameter")
			return
		}
		opts = append(opts, jetstream.WithPurgeKeep(keep))
	}

	if err := stream.Purge(ctx, opts...); err != nil {
		streamError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "purged"})
}

// msgResponse is the JSON representation of a stream message.
type msgResponse struct {
	Sequence uint64              `json:"sequence"`
	Subject  string              `json:"subject"`
	Data     string              `json:"data"`
	DataText string              `json:"dataText,omitempty"`
	Headers  map[string][]string `json:"headers,omitempty"`
	Time     time.Time           `json:"time"`
	Size     int                 `json:"size"`
}

func msgFromRaw(msg *jetstream.RawStreamMsg) msgResponse {
	m := msgResponse{
		Sequence: msg.Sequence,
		Subject:  msg.Subject,
		Data:     base64.StdEncoding.EncodeToString(msg.Data),
		Time:     msg.Time,
		Size:     len(msg.Data),
	}
	if utf8.Valid(msg.Data) {
		m.DataText = string(msg.Data)
	}
	if len(msg.Header) > 0 {
		m.Headers = map[string][]string(msg.Header)
	}
	return m
}

// listMessagesResponse is the paginated message list response.
type listMessagesResponse struct {
	Messages []msgResponse `json:"messages"`
	FirstSeq uint64        `json:"firstSeq"`
	LastSeq  uint64        `json:"lastSeq"`
	Total    uint64        `json:"total"`
}

// ListMessages handles GET /api/v1/streams/{stream}/messages
func (h *StreamHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "stream")
	ctx := r.Context()

	stream, err := js.Stream(ctx, name)
	if err != nil {
		streamError(w, err)
		return
	}

	info, err := stream.Info(ctx)
	if err != nil {
		streamError(w, err)
		return
	}

	firstSeq := info.State.FirstSeq
	lastSeq := info.State.LastSeq
	total := info.State.Msgs

	if total == 0 {
		JSON(w, http.StatusOK, listMessagesResponse{
			Messages: []msgResponse{},
			FirstSeq: firstSeq,
			LastSeq:  lastSeq,
			Total:    total,
		})
		return
	}

	// Parse query params
	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 200 {
		limit = 200
	}

	subject := r.URL.Query().Get("subject")

	// Determine start sequence — default to newest first (walk backwards from lastSeq)
	var startSeq uint64
	if ssStr := r.URL.Query().Get("startSeq"); ssStr != "" {
		if ss, err := strconv.ParseUint(ssStr, 10, 64); err == nil {
			startSeq = ss
		}
	}
	if startSeq == 0 {
		if sdStr := r.URL.Query().Get("startDate"); sdStr != "" {
			t, parseErr := time.Parse(time.RFC3339Nano, sdStr)
			if parseErr != nil {
				t, parseErr = time.Parse(time.RFC3339, sdStr)
			}
			if parseErr != nil {
				Error(w, http.StatusBadRequest, "invalid startDate: "+parseErr.Error())
				return
			}
			for seq := firstSeq; seq <= lastSeq; seq++ {
				var msg *jetstream.RawStreamMsg
				var getErr error
				if subject != "" {
					msg, getErr = stream.GetMsg(ctx, seq, jetstream.WithGetMsgSubject(subject))
				} else {
					msg, getErr = stream.GetMsg(ctx, seq)
				}
				if getErr != nil {
					if errors.Is(getErr, jetstream.ErrMsgNotFound) {
						continue
					}
					Error(w, http.StatusInternalServerError, "failed to get message: "+getErr.Error())
					return
				}
				if !msg.Time.Before(t) {
					startSeq = seq
					break
				}
			}
			if startSeq == 0 {
				JSON(w, http.StatusOK, listMessagesResponse{
					Messages: []msgResponse{},
					FirstSeq: firstSeq,
					LastSeq:  lastSeq,
					Total:    total,
				})
				return
			}
		}
	}

	userProvidedStartSeq := startSeq > 0

	if startSeq == 0 {
		startSeq = lastSeq
	}

	messages := make([]msgResponse, 0, limit)

	if userProvidedStartSeq {
		for seq := startSeq; seq <= lastSeq && len(messages) < limit; seq++ {
			var msg *jetstream.RawStreamMsg
			var getErr error
			if subject != "" {
				msg, getErr = stream.GetMsg(ctx, seq, jetstream.WithGetMsgSubject(subject))
			} else {
				msg, getErr = stream.GetMsg(ctx, seq)
			}
			if getErr != nil {
				if errors.Is(getErr, jetstream.ErrMsgNotFound) {
					continue
				}
				Error(w, http.StatusInternalServerError, "failed to get message: "+getErr.Error())
				return
			}
			messages = append(messages, msgFromRaw(msg))
		}
	} else {
		for seq := startSeq; seq >= firstSeq && len(messages) < limit; seq-- {
			var msg *jetstream.RawStreamMsg
			var getErr error
			if subject != "" {
				msg, getErr = stream.GetMsg(ctx, seq, jetstream.WithGetMsgSubject(subject))
			} else {
				msg, getErr = stream.GetMsg(ctx, seq)
			}
			if getErr != nil {
				if errors.Is(getErr, jetstream.ErrMsgNotFound) {
					continue
				}
				Error(w, http.StatusInternalServerError, "failed to get message: "+getErr.Error())
				return
			}
			messages = append(messages, msgFromRaw(msg))

			if seq == 0 {
				break
			}
		}
	}

	JSON(w, http.StatusOK, listMessagesResponse{
		Messages: messages,
		FirstSeq: firstSeq,
		LastSeq:  lastSeq,
		Total:    total,
	})
}

func (h *StreamHandler) GetLastBySubject(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "stream")
	subject := r.URL.Query().Get("subject")
	if subject == "" {
		Error(w, http.StatusBadRequest, "subject query parameter is required")
		return
	}

	ctx := r.Context()
	stream, err := js.Stream(ctx, name)
	if err != nil {
		streamError(w, err)
		return
	}

	info, err := stream.Info(ctx)
	if err != nil {
		streamError(w, err)
		return
	}

	msg, err := stream.GetMsg(ctx, info.State.LastSeq, jetstream.WithGetMsgSubject(subject))
	if err != nil {
		if errors.Is(err, jetstream.ErrMsgNotFound) {
			Error(w, http.StatusNotFound, "message not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to get message: "+natsErrMsg(err))
		return
	}

	JSON(w, http.StatusOK, msgFromRaw(msg))
}

// GetMessage handles GET /api/v1/streams/{stream}/messages/{seq}
func (h *StreamHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "stream")
	seqStr := chi.URLParam(r, "seq")
	seq, err := strconv.ParseUint(seqStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid sequence number")
		return
	}

	ctx := r.Context()
	stream, err := js.Stream(ctx, name)
	if err != nil {
		streamError(w, err)
		return
	}

	msg, err := stream.GetMsg(ctx, seq)
	if err != nil {
		streamError(w, err)
		return
	}

	JSON(w, http.StatusOK, msgFromRaw(msg))
}

// DeleteMessage handles DELETE /api/v1/streams/{stream}/messages/{seq}
func (h *StreamHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "stream")
	seqStr := chi.URLParam(r, "seq")
	seq, err := strconv.ParseUint(seqStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid sequence number")
		return
	}

	ctx := r.Context()
	stream, err := js.Stream(ctx, name)
	if err != nil {
		streamError(w, err)
		return
	}

	if err := stream.DeleteMsg(ctx, seq); err != nil {
		streamError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
