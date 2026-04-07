package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/seralaci/nats-insight/internal/nats"
	"github.com/seralaci/nats-insight/internal/store"
	wshandlers "github.com/seralaci/nats-insight/internal/ws"
)

func NewRouter(store *store.Store, manager *nats.Manager) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	connHandler := NewConnectionHandler(store, manager)
	serverHandler := NewServerHandler(manager)
	kvHandler := NewKvHandler(manager)
	streamHandler := NewStreamHandler(manager)
	consumerHandler := NewConsumerHandler(manager)
	objectHandler := NewObjectHandler(manager)
	tailHandler := wshandlers.NewTailHandler(manager)
	kvWatchHandler := wshandlers.NewKvWatchHandler(manager)
	metricsWS := wshandlers.NewMetricsHandler(manager)

	// WebSocket routes — registered outside the compressed route group so the
	// HTTP upgrade handshake is not interfered with by the Compress middleware.
	r.Get("/api/v1/ws/tail", tailHandler.ServeHTTP)
	r.Get("/api/v1/ws/kv/{bucket}/watch", kvWatchHandler.ServeHTTP)
	r.Get("/api/v1/ws/metrics", metricsWS.ServeHTTP)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			JSON(w, 200, map[string]string{"status": "ok"})
		})

		// Connection management
		r.Route("/connections", func(r chi.Router) {
			r.Get("/", connHandler.List)
			r.Post("/", connHandler.Create)
			r.Get("/{id}", connHandler.Get)
			r.Put("/{id}", connHandler.Update)
			r.Delete("/{id}", connHandler.Delete)
			r.Post("/{id}/connect", connHandler.Connect)
			r.Post("/{id}/disconnect", connHandler.Disconnect)
		})

		// Active connection status
		r.Get("/connection/status", connHandler.Status)

		// Server monitoring (requires active connection)
		r.Route("/server", func(r chi.Router) {
			r.Get("/varz", serverHandler.Varz)
			r.Get("/jsz", serverHandler.Jsz)
			r.Get("/connz", serverHandler.Connz)
			r.Get("/healthz", serverHandler.Healthz)
			r.Get("/accountz", serverHandler.Accountz)
			r.Get("/account-info", serverHandler.AccountInfo)
		})

		// JetStream Streams
		r.Route("/streams", func(r chi.Router) {
			r.Get("/", streamHandler.List)
			r.Post("/", streamHandler.Create)
			r.Get("/{stream}", streamHandler.Get)
			r.Put("/{stream}", streamHandler.Update)
			r.Delete("/{stream}", streamHandler.Delete)
			r.Post("/{stream}/purge", streamHandler.Purge)
			r.Get("/{stream}/messages", streamHandler.ListMessages)
			r.Get("/{stream}/messages/last", streamHandler.GetLastBySubject)
			r.Get("/{stream}/messages/{seq}", streamHandler.GetMessage)
			r.Delete("/{stream}/messages/{seq}", streamHandler.DeleteMessage)
			r.Get("/{stream}/consumers", consumerHandler.List)
			r.Get("/{stream}/consumers/{consumer}", consumerHandler.Get)
			r.Delete("/{stream}/consumers/{consumer}", consumerHandler.Delete)
			r.Post("/{stream}/consumers/{consumer}/pause", consumerHandler.Pause)
			r.Post("/{stream}/consumers/{consumer}/resume", consumerHandler.Resume)
		})

		// KV Store
		r.Route("/kv", func(r chi.Router) {
			r.Get("/buckets", kvHandler.ListBuckets)
			r.Post("/buckets", kvHandler.CreateBucket)
			r.Get("/buckets/{bucket}", kvHandler.GetBucket)
			r.Delete("/buckets/{bucket}", kvHandler.DeleteBucket)
			r.Get("/buckets/{bucket}/keys", kvHandler.ListKeys)
			r.Route("/buckets/{bucket}/keys/{key}", func(r chi.Router) {
				r.Get("/", kvHandler.GetKey)
				r.Put("/", kvHandler.PutKey)
				r.Delete("/", kvHandler.DeleteKey)
				r.Post("/purge", kvHandler.PurgeKey)
				r.Get("/history", kvHandler.GetKeyHistory)
			})
		})

		// Object Store
		r.Route("/objects/stores", func(r chi.Router) {
			r.Get("/", objectHandler.ListStores)
			r.Post("/", objectHandler.CreateStore)
			r.Get("/{store}", objectHandler.GetStore)
			r.Delete("/{store}", objectHandler.DeleteStore)
			r.Get("/{store}/objects", objectHandler.ListObjects)
			r.Post("/{store}/objects", objectHandler.UploadObject)
			r.Get("/{store}/objects/{name}", objectHandler.GetObjectInfo)
			r.Get("/{store}/objects/{name}/data", objectHandler.DownloadObject)
			r.Delete("/{store}/objects/{name}", objectHandler.DeleteObject)
		})
	})

	return r
}
