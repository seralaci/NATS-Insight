package api

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go/jetstream"
	natsmgr "github.com/seralaci/nats-insight/internal/nats"
)

type ObjectHandler struct {
	manager *natsmgr.Manager
}

func NewObjectHandler(m *natsmgr.Manager) *ObjectHandler {
	return &ObjectHandler{manager: m}
}

func (h *ObjectHandler) requireJetStream(w http.ResponseWriter) (jetstream.JetStream, bool) {
	js, err := h.manager.JetStream()
	if err != nil {
		Error(w, http.StatusServiceUnavailable, "JetStream not available: "+err.Error())
		return nil, false
	}
	return js, true
}

func objectStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, jetstream.ErrBucketNotFound):
		Error(w, http.StatusNotFound, "object store not found")
	case errors.Is(err, jetstream.ErrObjectNotFound):
		Error(w, http.StatusNotFound, "object not found")
	default:
		Error(w, http.StatusInternalServerError, natsErrMsg(err))
	}
}

type objectStoreInfo struct {
	Bucket      string            `json:"bucket"`
	Description string            `json:"description,omitempty"`
	Sealed      bool              `json:"sealed"`
	Size        uint64            `json:"size"`
	Storage     string            `json:"storage"`
	Replicas    int               `json:"replicas"`
	TTL         string            `json:"ttl,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func objectStoreInfoFromStatus(s jetstream.ObjectStoreStatus) objectStoreInfo {
	info := objectStoreInfo{
		Bucket:      s.Bucket(),
		Description: s.Description(),
		Sealed:      s.Sealed(),
		Size:        s.Size(),
		Replicas:    s.Replicas(),
		Metadata:    s.Metadata(),
	}
	if s.TTL() > 0 {
		info.TTL = s.TTL().String()
	}
	switch s.Storage() {
	case jetstream.MemoryStorage:
		info.Storage = "memory"
	default:
		info.Storage = "file"
	}
	return info
}

type objectInfo struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Size        uint64            `json:"size"`
	Chunks      uint32            `json:"chunks"`
	Digest      string            `json:"digest,omitempty"`
	Modified    time.Time         `json:"modified"`
	Deleted     bool              `json:"deleted,omitempty"`
	Headers     map[string][]string `json:"headers,omitempty"`
}

func objectInfoFromJetStream(o *jetstream.ObjectInfo) objectInfo {
	info := objectInfo{
		Name:        o.Name,
		Description: o.Description,
		Size:        o.Size,
		Chunks:      o.Chunks,
		Digest:      o.Digest,
		Modified:    o.ModTime,
		Deleted:     o.Deleted,
	}
	if len(o.Headers) > 0 {
		info.Headers = map[string][]string(o.Headers)
	}
	return info
}

type createObjectStoreRequest struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	MaxChunkSize int32             `json:"maxChunkSize"`
	Storage      string            `json:"storage"`
	Replicas     int               `json:"replicas"`
	Metadata     map[string]string `json:"metadata"`
	MaxBytes     string            `json:"maxBytes"`
}

func (h *ObjectHandler) ListStores(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	ctx := r.Context()
	lister := js.ObjectStores(ctx)

	var stores []objectStoreInfo
	for status := range lister.Status() {
		stores = append(stores, objectStoreInfoFromStatus(status))
	}
	if err := lister.Error(); err != nil {
		Error(w, http.StatusInternalServerError, "failed to list object stores: "+natsErrMsg(err))
		return
	}

	if stores == nil {
		stores = []objectStoreInfo{}
	}
	JSON(w, http.StatusOK, stores)
}

func (h *ObjectHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	var req createObjectStoreRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		Error(w, http.StatusBadRequest, "name is required")
		return
	}

	cfg := jetstream.ObjectStoreConfig{
		Bucket:      req.Name,
		Description: req.Description,
		Metadata:    req.Metadata,
	}

	switch req.Storage {
	case "memory":
		cfg.Storage = jetstream.MemoryStorage
	default:
		cfg.Storage = jetstream.FileStorage
	}

	if req.Replicas > 0 {
		cfg.Replicas = req.Replicas
	} else {
		cfg.Replicas = 1
	}

	if req.MaxBytes != "" {
		size, err := ParseByteSize(req.MaxBytes)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid max bytes: "+err.Error())
			return
		}
		cfg.MaxBytes = size
	}

	ctx := r.Context()
	store, err := js.CreateObjectStore(ctx, cfg)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create object store: "+natsErrMsg(err))
		return
	}

	status, err := store.Status(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "store created but failed to get status: "+natsErrMsg(err))
		return
	}

	JSON(w, http.StatusCreated, objectStoreInfoFromStatus(status))
}

func (h *ObjectHandler) GetStore(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "store")
	ctx := r.Context()

	store, err := js.ObjectStore(ctx, name)
	if err != nil {
		objectStoreError(w, err)
		return
	}

	status, err := store.Status(ctx)
	if err != nil {
		objectStoreError(w, err)
		return
	}

	JSON(w, http.StatusOK, objectStoreInfoFromStatus(status))
}

func (h *ObjectHandler) DeleteStore(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "store")
	ctx := r.Context()

	if err := js.DeleteObjectStore(ctx, name); err != nil {
		objectStoreError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *ObjectHandler) ListObjects(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	name := chi.URLParam(r, "store")
	ctx := r.Context()

	store, err := js.ObjectStore(ctx, name)
	if err != nil {
		objectStoreError(w, err)
		return
	}

	objects, err := store.List(ctx)
	if err != nil {
		if errors.Is(err, jetstream.ErrNoObjectsFound) {
			JSON(w, http.StatusOK, []objectInfo{})
			return
		}
		objectStoreError(w, err)
		return
	}

	result := make([]objectInfo, 0, len(objects))
	for _, o := range objects {
		result = append(result, objectInfoFromJetStream(o))
	}

	JSON(w, http.StatusOK, result)
}

func (h *ObjectHandler) GetObjectInfo(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	storeName := chi.URLParam(r, "store")
	rawName := chi.URLParam(r, "name")
	objectName, err := url.PathUnescape(rawName)
	if err != nil {
		objectName = rawName
	}

	ctx := r.Context()

	store, err := js.ObjectStore(ctx, storeName)
	if err != nil {
		objectStoreError(w, err)
		return
	}

	info, err := store.GetInfo(ctx, objectName)
	if err != nil {
		objectStoreError(w, err)
		return
	}

	JSON(w, http.StatusOK, objectInfoFromJetStream(info))
}

func (h *ObjectHandler) DownloadObject(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	storeName := chi.URLParam(r, "store")
	rawName := chi.URLParam(r, "name")
	objectName, err := url.PathUnescape(rawName)
	if err != nil {
		objectName = rawName
	}

	ctx := r.Context()

	store, err := js.ObjectStore(ctx, storeName)
	if err != nil {
		objectStoreError(w, err)
		return
	}

	result, err := store.Get(ctx, objectName)
	if err != nil {
		objectStoreError(w, err)
		return
	}
	defer result.Close()

	info, err := result.Info()
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get object info: "+natsErrMsg(err))
		return
	}

	contentType := "application/octet-stream"
	if info != nil && info.Headers != nil {
		if ct := info.Headers.Get("Content-Type"); ct != "" {
			contentType = ct
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+objectName+"\"")
	io.Copy(w, result)
}

func (h *ObjectHandler) UploadObject(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	storeName := chi.URLParam(r, "store")
	ctx := r.Context()

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		Error(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		Error(w, http.StatusBadRequest, "missing file field: "+err.Error())
		return
	}
	defer file.Close()

	description := r.FormValue("description")

	meta := jetstream.ObjectMeta{
		Name:        header.Filename,
		Description: description,
	}

	store, err := js.ObjectStore(ctx, storeName)
	if err != nil {
		objectStoreError(w, err)
		return
	}

	info, err := store.Put(ctx, meta, file)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to upload object: "+natsErrMsg(err))
		return
	}

	JSON(w, http.StatusCreated, objectInfoFromJetStream(info))
}

func (h *ObjectHandler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	js, ok := h.requireJetStream(w)
	if !ok {
		return
	}

	storeName := chi.URLParam(r, "store")
	rawName := chi.URLParam(r, "name")
	objectName, err := url.PathUnescape(rawName)
	if err != nil {
		objectName = rawName
	}

	ctx := r.Context()

	store, err := js.ObjectStore(ctx, storeName)
	if err != nil {
		objectStoreError(w, err)
		return
	}

	if err := store.Delete(ctx, objectName); err != nil {
		objectStoreError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
