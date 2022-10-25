package api

import (
	"net/http"

	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/broswen/notifi/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type API struct {
	Producer     producer.Producer
	Notification repository.NotificationRepository
}

func (api *API) Router(accessClient AccessClient) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://notifi.broswen.com", "http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, nil, ErrNotFound)
	})

	r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, http.StatusOK, "OK")
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(CloudflareAccessVerifier(accessClient))
		r.Use(CloudflareAccessIdentityLogger(accessClient))
		r.Post("/api/notifications", api.HandleCreateNotification())
		r.Get("/api/notifications/{notificationId}", api.HandleGetNotification())
		r.Delete("/api/notifications/{notificationId}", api.HandleDeleteNotification())

	})

	return r
}
