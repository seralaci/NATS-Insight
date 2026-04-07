package api

import (
	"encoding/base64"
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go/jetstream"
	natsmgr "github.com/seralaci/nats-insight/internal/nats"
)

type KvHandler struct {
	manager *natsmgr.Manager
}

func NewKvHandler(m *natsmgr.Manager) *KvHandler {
	return &KvHandler{manager: m}
}

func (h *KvHandler) requireJetStream(w http.ResponseWriter) (jetstream.JetStream, bool) {
	js, err := h.manager.JetStream()
	if err != nil {
		Error(w, http.StatusServiceUnavailable, "JetStream not available: "+err.Error())
		return nil, false
	}
	return js, true
}

func kvError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, jetstream.ErrBucketNotFound):
		Error(w, http.StatusNotFound, "bucket not found")
	case errors.Is(err, jetstream.ErrKeyNotFound):
		Error(w, http.StatusNotFound, "key not found")
	case errors.Is(err, jetstream.ErrKeyDeleted):
		Error(w, http.StatusNotFound, "key has been deleted")
	default:
		Error(w, http.StatusInternalServerError, natsErrMsg(err))
	}
}

// bucketInfo is the JSON representation of a KV bucket status.
type bucketInfo struct {
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Values       uint64            `json:"values"`
	Bytes        uint64            `json:"bytes"`
	History      int64             `json:"history"`
	TTL          string            `json:"ttl,omitempty"`
	Storage      string            `json:"storage"`
	Replicas     int               `json:"replicas"`
	IsCompressed bool              `json:"isCompressed"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

func bucketInfoFromStatus(s jetstream.KeyValueStatus) bucketInfo {
	cfg := s.Config()
	info := bucketInfo{
		Name:         s.Bucket(),
		Values:       s.Values(),
		Bytes:        s.Bytes(),
		History:      s.History(),
		Replicas:     cfg.Replicas,
		IsCompressed: s.IsCompressed(),
		Metadata:     s.Metadata(),
	}
	if s.TTL() > 0 {
		info.TTL = s.TTL().String()
	}
	switch cfg.Storage {
	case jetstream.MemoryStorage:
		info.Storage = "memory"
	default:
		info.Storage = "file"
	}
	return info
}

// keyEntry is the JSON representation of a KV entry.
type keyEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	ValueText string    `json:"valueText,omitempty"`
	Revision  uint64    `json:"revision"`
	Created   time.Time `json:"created"`
	Operation string    `json:"operation"`
	Delta     uint64    `json:"delta"`
	Bucket    string    `json:"bucket"`
}

func entryFromKV(e jetstream.KeyValueEntry) keyEntry {
	encoded := base64.StdEncoding.EncodeToString(e.Value())
	ke := keyEntry{
		Key:       e.Key(),
		Value:     encoded,
		Revision:  e.Revision(),
		Created:   e.Created(),
		Operation: e.Operation().String(),
		Delta:     e.Delta(),
		Bucket:    e.Bucket(),
	}
	if utf8.Valid(e.Value()) {
		ke.ValueText = string(e.Value())
	}
	return ke
}

// ListBuckets handles GET /api/v1/kv/buckets
func (h *KvHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	ctx := r.Context()
	lister := js.KeyValueStores(ctx)

	var buckets []bucketInfo
	for status := range lister.Status() {
		buckets = append(buckets, bucketInfoFromStatus(status))
	}
	if err := lister.Error(); err != nil {
		Error(w, http.StatusInternalServerError, "failed to list buckets: "+natsErrMsg(err))
		return
	}

	if buckets == nil {
		buckets = []bucketInfo{}
	}
	JSON(w, http.StatusOK, buckets)
}

// createBucketRequest is the JSON body for creating a KV bucket.
type createBucketRequest struct {
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	History              int64             `json:"history"`
	TTL                  string            `json:"ttl"`
	LimitMarkerTTL       string            `json:"limitMarkerTtl"`
	Storage              string            `json:"storage"`
	Replicas             int               `json:"replicas"`
	MaxBucketSize        string            `json:"maxBucketSize"`
	MaxValueSize         string            `json:"maxValueSize"`
	Compression          bool              `json:"compression"`
	Metadata             map[string]string `json:"metadata"`
	Tags                 []string          `json:"tags"`
	RepublishSource      string            `json:"republishSource"`
	RepublishDestination string            `json:"republishDestination"`
}

// CreateBucket handles POST /api/v1/kv/buckets
func (h *KvHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	var req createBucketRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		Error(w, http.StatusBadRequest, "name is required")
		return
	}

	cfg := jetstream.KeyValueConfig{
		Bucket:      req.Name,
		Description: req.Description,
		Metadata:    req.Metadata,
	}

	if req.History > 0 {
		cfg.History = uint8(req.History)
	} else {
		cfg.History = 1
	}

	if req.Replicas > 0 {
		cfg.Replicas = req.Replicas
	} else {
		cfg.Replicas = 1
	}

	if req.TTL != "" {
		d, err := time.ParseDuration(req.TTL)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid ttl duration: "+err.Error())
			return
		}
		cfg.TTL = d
	}

	if req.LimitMarkerTTL != "" {
		d, err := time.ParseDuration(req.LimitMarkerTTL)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid limit marker ttl: "+err.Error())
			return
		}
		cfg.LimitMarkerTTL = d
	}

	switch req.Storage {
	case "memory":
		cfg.Storage = jetstream.MemoryStorage
	default:
		cfg.Storage = jetstream.FileStorage
	}

	if req.MaxBucketSize != "" {
		size, err := ParseByteSize(req.MaxBucketSize)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid max bucket size: "+err.Error())
			return
		}
		cfg.MaxBytes = size
	}
	if req.MaxValueSize != "" {
		size, err := ParseByteSize(req.MaxValueSize)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid max value size: "+err.Error())
			return
		}
		cfg.MaxValueSize = int32(size)
	}
	if req.Compression {
		cfg.Compression = true
	}

	if req.RepublishSource != "" || req.RepublishDestination != "" {
		cfg.RePublish = &jetstream.RePublish{
			Source:      req.RepublishSource,
			Destination: req.RepublishDestination,
		}
	}

	ctx := r.Context()
	kv, err := js.CreateKeyValue(ctx, cfg)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create bucket: "+natsErrMsg(err))
		return
	}

	status, err := kv.Status(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "bucket created but failed to get status: "+natsErrMsg(err))
		return
	}

	JSON(w, http.StatusCreated, bucketInfoFromStatus(status))
}

// GetBucket handles GET /api/v1/kv/buckets/{bucket}
func (h *KvHandler) GetBucket(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "bucket")
	ctx := r.Context()

	kv, err := js.KeyValue(ctx, name)
	if err != nil {
		kvError(w, err)
		return
	}

	status, err := kv.Status(ctx)
	if err != nil {
		kvError(w, err)
		return
	}

	JSON(w, http.StatusOK, bucketInfoFromStatus(status))
}

// DeleteBucket handles DELETE /api/v1/kv/buckets/{bucket}
func (h *KvHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "bucket")
	ctx := r.Context()

	if err := js.DeleteKeyValue(ctx, name); err != nil {
		kvError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListKeys handles GET /api/v1/kv/buckets/{bucket}/keys
func (h *KvHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "bucket")
	ctx := r.Context()

	kv, err := js.KeyValue(ctx, name)
	if err != nil {
		kvError(w, err)
		return
	}

	filter := r.URL.Query().Get("q")

	var lister jetstream.KeyLister
	if filter != "" {
		lister, err = kv.ListKeysFiltered(ctx, filter)
	} else {
		lister, err = kv.ListKeys(ctx)
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list keys: "+natsErrMsg(err))
		return
	}
	defer lister.Stop()

	keys := []string{}
	for key := range lister.Keys() {
		keys = append(keys, key)
	}

	JSON(w, http.StatusOK, keys)
}

// GetKey handles GET /api/v1/kv/buckets/{bucket}/keys/{key}
func (h *KvHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")
	ctx := r.Context()

	kv, err := js.KeyValue(ctx, bucket)
	if err != nil {
		kvError(w, err)
		return
	}

	entry, err := kv.Get(ctx, key)
	if err != nil {
		kvError(w, err)
		return
	}

	JSON(w, http.StatusOK, entryFromKV(entry))
}

// putKeyRequest is the JSON body for putting a key value.
type putKeyRequest struct {
	Value string `json:"value"`
}

// PutKey handles PUT /api/v1/kv/buckets/{bucket}/keys/{key}
func (h *KvHandler) PutKey(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")
	ctx := r.Context()

	var req putKeyRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	kv, err := js.KeyValue(ctx, bucket)
	if err != nil {
		kvError(w, err)
		return
	}

	if _, err := kv.PutString(ctx, key, req.Value); err != nil {
		kvError(w, err)
		return
	}

	entry, err := kv.Get(ctx, key)
	if err != nil {
		kvError(w, err)
		return
	}

	JSON(w, http.StatusOK, entryFromKV(entry))
}

// DeleteKey handles DELETE /api/v1/kv/buckets/{bucket}/keys/{key}
func (h *KvHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")
	ctx := r.Context()

	kv, err := js.KeyValue(ctx, bucket)
	if err != nil {
		kvError(w, err)
		return
	}

	if err := kv.Delete(ctx, key); err != nil {
		kvError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// PurgeKey handles POST /api/v1/kv/buckets/{bucket}/keys/{key}/purge
func (h *KvHandler) PurgeKey(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")
	ctx := r.Context()

	kv, err := js.KeyValue(ctx, bucket)
	if err != nil {
		kvError(w, err)
		return
	}

	if err := kv.Purge(ctx, key); err != nil {
		kvError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "purged"})
}

// GetKeyHistory handles GET /api/v1/kv/buckets/{bucket}/keys/{key}/history
func (h *KvHandler) GetKeyHistory(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "key")
	ctx := r.Context()

	kv, err := js.KeyValue(ctx, bucket)
	if err != nil {
		kvError(w, err)
		return
	}

	entries, err := kv.History(ctx, key)
	if err != nil {
		kvError(w, err)
		return
	}

	result := make([]keyEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, entryFromKV(e))
	}

	JSON(w, http.StatusOK, result)
}
